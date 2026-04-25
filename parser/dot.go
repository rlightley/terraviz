package parser

import (
	"bufio"
	"regexp"
	"strings"
	"terraviz/model"
)

var (
	// Match DOT node declarations like: "[root] aws_instance.example" [label = "aws_instance.example",shape = "box"];
	nodeRegex = regexp.MustCompile(`^\s*"?\[root\]\s*([^"\[\]]+)"?\s*\[`)
	// Match DOT edge declarations like: "[root] aws_instance.example" -> "[root] aws_security_group.example";
	edgeRegex = regexp.MustCompile(`^\s*"?\[root\]\s*([^"\[\]]+)"?\s*->\s*"?\[root\]\s*([^"\[\]]+)"?`)
)

// ParseDOT parses DOT format output from terraform graph and returns a Graph
func ParseDOT(dotOutput string) (*model.Graph, error) {
	graph := model.NewGraph()
	scanner := bufio.NewScanner(strings.NewReader(dotOutput))

	nodes := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Try to match edge first (since edges also contain node names)
		if matches := edgeRegex.FindStringSubmatch(line); matches != nil {
			from := strings.TrimSpace(matches[1])
			to := strings.TrimSpace(matches[2])

			// Mark nodes as seen
			nodes[from] = true
			nodes[to] = true

			// Add edge
			graph.AddEdge(from, to)
			continue
		}

		// Try to match node
		if matches := nodeRegex.FindStringSubmatch(line); matches != nil {
			nodeName := strings.TrimSpace(matches[1])
			nodes[nodeName] = true
		}
	}

	// Create resources from nodes
	for nodeName := range nodes {
		resource := parseNodeName(nodeName)
		graph.AddResource(resource)
	}

	return graph, scanner.Err()
}

// parseNodeName parses a DOT node name into a Resource
// Expected format: "provider_resourcetype.name" or "module.modname.provider_resourcetype.name"
func parseNodeName(nodeName string) *model.Resource {
	parts := strings.Split(nodeName, ".")

	resource := &model.Resource{
		ID:           nodeName,
		Attributes:   make(map[string]any),
		Dependencies: make([]string, 0),
	}

	// Handle module prefix
	modulePrefix := ""
	typeAndName := nodeName

	if len(parts) >= 2 && parts[0] == "module" {
		// Find where the module path ends (before the resource type)
		for i := 1; i < len(parts)-1; i++ {
			if strings.Contains(parts[i+1], "_") || i == len(parts)-2 {
				// This is likely the end of module path
				modulePrefix = strings.Join(parts[:i+1], ".")
				typeAndName = strings.Join(parts[i+1:], ".")
				break
			}
		}
	}

	resource.Module = modulePrefix

	// Parse type and name from the remaining part
	lastDot := strings.LastIndex(typeAndName, ".")
	if lastDot > 0 {
		resource.Type = typeAndName[:lastDot]
		resource.Name = typeAndName[lastDot+1:]
	} else {
		resource.Type = typeAndName
		resource.Name = ""
	}

	// Derive provider from type (prefix before first underscore)
	if idx := strings.Index(resource.Type, "_"); idx > 0 {
		resource.Provider = resource.Type[:idx]
	} else {
		resource.Provider = resource.Type
	}

	return resource
}
