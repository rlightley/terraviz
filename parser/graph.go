package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"terraviz/model"
)

// ParseMode indicates which parsing mode was used
type ParseMode string

const (
	ModeHCLParsing     ParseMode = "HCL parsing"
	ModeTerraformGraph ParseMode = "terraform graph (enhanced)"
	ModeStateFile      ParseMode = "state file"
)

// ParseResult contains the parsed graph and metadata about the parsing process
type ParseResult struct {
	Graph        *model.Graph
	Mode         ParseMode
	FailedFiles  []string
	ErrorMessage string
}

// BuildGraph is the main entry point for parsing. It determines the mode and delegates to the appropriate parser.
func BuildGraph(path string, fromRemote bool) *ParseResult {
	// Check if path is a state file
	if strings.HasSuffix(path, ".tfstate") {
		return parseStateFile(path)
	}

	// If fromRemote flag is set, use terraform state pull
	if fromRemote {
		return parseRemoteState(path)
	}

	// Otherwise, treat as directory - try terraform graph first
	return parseDirectory(path)
}

// parseDirectory parses .tf files directly (primary mode for planning infrastructure)
func parseDirectory(dir string) *ParseResult {
	// Primary mode: Parse HCL directly
	// This is what engineers want - to see what they're about to deploy
	graph, failedFiles, hclErr := ParseHCL(dir)

	mode := ModeHCLParsing
	errorMsg := ""

	if hclErr != nil {
		errorMsg = fmt.Sprintf("HCL parsing error: %v", hclErr)
	}

	// Try to enhance with terraform graph if available (provides better dependency info)
	cmd := exec.Command("terraform", "graph")
	cmd.Dir = dir
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		// Parse terraform graph to get enhanced dependency information
		tfGraph, parseErr := ParseDOT(string(output))
		if parseErr == nil {
			// Use terraform graph edges if available (more accurate)
			graph.Edges = tfGraph.Edges
			mode = ModeTerraformGraph

			// Also add any resources from terraform graph that we might have missed
			for id, res := range tfGraph.Resources {
				if _, exists := graph.Resources[id]; !exists {
					graph.Resources[id] = res
				}
			}
		}
	}

	return &ParseResult{
		Graph:       graph,
		Mode:        mode,
		FailedFiles: failedFiles,
		ErrorMessage: errorMsg,
	}
}

// parseStateFile parses a local .tfstate file
func parseStateFile(path string) *ParseResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return &ParseResult{
			Graph:        model.NewGraph(),
			Mode:         ModeStateFile,
			ErrorMessage: fmt.Sprintf("Failed to read state file: %v", err),
		}
	}

	graph, parseErr := parseStateJSON(content)
	result := &ParseResult{
		Graph: graph,
		Mode:  ModeStateFile,
	}

	if parseErr != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to parse state file: %v", parseErr)
	}

	return result
}

// parseRemoteState uses terraform state pull to get remote state
func parseRemoteState(dir string) *ParseResult {
	cmd := exec.Command("terraform", "state", "pull")
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		return &ParseResult{
			Graph:        model.NewGraph(),
			Mode:         ModeStateFile,
			ErrorMessage: fmt.Sprintf("Failed to pull remote state: %v", err),
		}
	}

	graph, parseErr := parseStateJSON(output)
	result := &ParseResult{
		Graph: graph,
		Mode:  ModeStateFile,
	}

	if parseErr != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to parse remote state: %v", parseErr)
	}

	return result
}

// TerraformState represents the structure of a Terraform state file
type TerraformState struct {
	Resources []StateResource `json:"resources"`
}

// StateResource represents a resource in the state file
type StateResource struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Module    string          `json:"module"`
	Instances []StateInstance `json:"instances"`
}

// StateInstance represents an instance of a resource
type StateInstance struct {
	Attributes   map[string]any `json:"attributes"`
	Dependencies []string       `json:"dependencies"`
}

// parseStateJSON parses JSON state file content into a Graph
func parseStateJSON(content []byte) (*model.Graph, error) {
	var state TerraformState
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, err
	}

	graph := model.NewGraph()

	for _, res := range state.Resources {
		if len(res.Instances) == 0 {
			continue
		}

		// Use only the first instance as specified
		instance := res.Instances[0]

		id := fmt.Sprintf("%s.%s", res.Type, res.Name)
		if res.Module != "" {
			id = fmt.Sprintf("%s.%s", res.Module, id)
		}

		resource := &model.Resource{
			ID:           id,
			Type:         res.Type,
			Name:         res.Name,
			Module:       res.Module,
			Attributes:   instance.Attributes,
			Dependencies: instance.Dependencies,
		}

		// Derive provider from type prefix
		if idx := strings.Index(res.Type, "_"); idx > 0 {
			resource.Provider = res.Type[:idx]
		} else {
			resource.Provider = res.Type
		}

		graph.AddResource(resource)

		// Add edges for dependencies
		for _, dep := range instance.Dependencies {
			graph.AddEdge(id, dep)
		}
	}

	return graph, nil
}
