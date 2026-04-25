package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"terraviz/model"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// ParseHCL walks a directory and parses all .tf files, returning a Graph
func ParseHCL(dir string) (*model.Graph, []string, error) {
	graph := model.NewGraph()
	var failedFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		if err := parseFile(path, graph); err != nil {
			failedFiles = append(failedFiles, filepath.Base(path))
		}

		return nil
	})

	// Build edges from dependencies
	for id, res := range graph.Resources {
		for _, depID := range res.Dependencies {
			graph.AddEdge(id, depID)
		}
	}

	return graph, failedFiles, err
}

// parseFile parses a single .tf file and adds resources to the graph
func parseFile(path string, graph *model.Graph) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	file, diags := hclsyntax.ParseConfig(content, path, hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	for _, block := range body.Blocks {
		if block.Type == "resource" {
			if len(block.Labels) < 2 {
				continue
			}

			resource := parseResourceBlock(block)
			graph.AddResource(resource)
		}
	}

	return nil
}

// parseResourceBlock parses a resource block into a Resource
func parseResourceBlock(block *hclsyntax.Block) *model.Resource {
	resourceType := block.Labels[0]
	resourceName := block.Labels[1]
	id := fmt.Sprintf("%s.%s", resourceType, resourceName)

	resource := &model.Resource{
		ID:           id,
		Type:         resourceType,
		Name:         resourceName,
		Module:       "",
		Attributes:   make(map[string]any),
		Dependencies: make([]string, 0),
	}

	// Derive provider from type prefix
	if idx := strings.Index(resourceType, "_"); idx > 0 {
		resource.Provider = resourceType[:idx]
	} else {
		resource.Provider = resourceType
	}

	// Parse block attributes
	body := block.Body

	for name, attr := range body.Attributes {
		if name == "depends_on" {
			// Extract dependencies
			deps := extractDependencies(attr.Expr)
			resource.Dependencies = deps
		} else {
			// Store attribute as string
			value := exprToString(attr.Expr)
			resource.Attributes[name] = value
		}
	}

	return resource
}

// extractDependencies extracts dependency IDs from a depends_on expression
func extractDependencies(expr hclsyntax.Expression) []string {
	var deps []string

	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		for _, elem := range e.Exprs {
			if dep := exprToString(elem); dep != "" {
				deps = append(deps, dep)
			}
		}
	}

	return deps
}

// exprToString converts an HCL expression to a string representation
func exprToString(expr hclsyntax.Expression) string {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyValueToString(e.Val)
	case *hclsyntax.TemplateExpr:
		var parts []string
		for _, part := range e.Parts {
			parts = append(parts, exprToString(part))
		}
		return strings.Join(parts, "")
	case *hclsyntax.TemplateWrapExpr:
		return exprToString(e.Wrapped)
	case *hclsyntax.ScopeTraversalExpr:
		return traversalToString(e.Traversal)
	case *hclsyntax.FunctionCallExpr:
		return fmt.Sprintf("%s(...)", e.Name)
	case *hclsyntax.TupleConsExpr:
		var items []string
		for _, elem := range e.Exprs {
			items = append(items, exprToString(elem))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case *hclsyntax.ObjectConsExpr:
		return "{...}"
	default:
		// Return the source range as a fallback
		if e != nil {
			src := e.Range().SliceBytes([]byte{})
			return string(src)
		}
		return ""
	}
}

// ctyValueToString converts a cty.Value to a string
func ctyValueToString(val cty.Value) string {
	if val.IsNull() {
		return "null"
	}

	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return fmt.Sprintf("%v", f)
	case cty.Bool:
		return fmt.Sprintf("%v", val.True())
	default:
		return fmt.Sprintf("%#v", val)
	}
}

// traversalToString converts a traversal to a string (e.g., aws_instance.example)
func traversalToString(traversal hcl.Traversal) string {
	var parts []string
	for _, traverser := range traversal {
		switch t := traverser.(type) {
		case hcl.TraverseRoot:
			parts = append(parts, t.Name)
		case hcl.TraverseAttr:
			parts = append(parts, t.Name)
		case hcl.TraverseIndex:
			key := ctyValueToString(t.Key)
			parts = append(parts, fmt.Sprintf("[%s]", key))
		}
	}
	return strings.Join(parts, ".")
}
