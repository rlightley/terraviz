package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"terraviz/model"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type moduleParseContext struct {
	prefix string
	inputs map[string][]string
}

// ParseHCL walks a directory and parses all .tf files, returning a Graph
func ParseHCL(dir string) (*model.Graph, []string, error) {
	graph := model.NewGraph()
	failedFiles := make([]string, 0)
	visited := make(map[string]bool)

	err := parseModuleDir(filepath.Clean(dir), moduleParseContext{}, graph, &failedFiles, visited)

	// Build edges from dependencies
	for id, res := range graph.Resources {
		for _, depID := range res.Dependencies {
			graph.AddEdge(id, depID)
		}
	}

	return graph, failedFiles, err
}

func parseModuleDir(dir string, context moduleParseContext, graph *model.Graph, failedFiles *[]string, visited map[string]bool) error {
	visitKey := dir + "|" + context.prefix
	if visited[visitKey] {
		return nil
	}
	visited[visitKey] = true

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := parseFile(path, dir, context, graph, failedFiles, visited); err != nil {
			*failedFiles = append(*failedFiles, filepath.Base(path))
		}
	}

	return nil
}

// parseFile parses a single .tf file and adds resources to the graph
func parseFile(path, currentDir string, context moduleParseContext, graph *model.Graph, failedFiles *[]string, visited map[string]bool) error {
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
		switch block.Type {
		case "resource":
			if len(block.Labels) < 2 {
				continue
			}

			resource := parseResourceBlock(block, context)
			graph.AddResource(resource)
		case "module":
			if err := parseModuleBlock(block, currentDir, context, graph, failedFiles, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseResourceBlock parses a resource block into a Resource
func parseResourceBlock(block *hclsyntax.Block, context moduleParseContext) *model.Resource {
	resourceType := block.Labels[0]
	resourceName := block.Labels[1]
	id := fmt.Sprintf("%s.%s", resourceType, resourceName)
	if context.prefix != "" {
		id = context.prefix + "." + id
	}

	resource := &model.Resource{
		ID:           id,
		Type:         resourceType,
		Name:         resourceName,
		Module:       context.prefix,
		Attributes:   make(map[string]any),
		Dependencies: make([]string, 0),
		LocalRefs:    make([]string, 0),
		VariableRefs: make([]string, 0),
		ModuleInputs: make(map[string][]string),
	}

	// Derive provider from type prefix
	if idx := strings.Index(resourceType, "_"); idx > 0 {
		resource.Provider = resourceType[:idx]
	} else {
		resource.Provider = resourceType
	}

	collectResourceMetadata(block.Body, resource, true)
	applyModuleInputs(resource, context.inputs)

	return resource
}

func parseModuleBlock(block *hclsyntax.Block, currentDir string, parentContext moduleParseContext, graph *model.Graph, failedFiles *[]string, visited map[string]bool) error {
	if len(block.Labels) == 0 {
		return nil
	}

	moduleName := block.Labels[0]
	moduleSource := moduleSourcePath(block.Body)
	if moduleSource == "" || !isLocalModuleSource(moduleSource) {
		return nil
	}

	moduleDir := filepath.Clean(filepath.Join(currentDir, moduleSource))
	modulePrefix := "module." + moduleName
	if parentContext.prefix != "" {
		modulePrefix = parentContext.prefix + "." + modulePrefix
	}

	childContext := moduleParseContext{
		prefix: modulePrefix,
		inputs: moduleInputRefs(block.Body, parentContext.inputs),
	}

	return parseModuleDir(moduleDir, childContext, graph, failedFiles, visited)
}

func moduleSourcePath(body *hclsyntax.Body) string {
	if body == nil {
		return ""
	}

	attr, ok := body.Attributes["source"]
	if !ok {
		return ""
	}

	return exprToString(attr.Expr)
}

func isLocalModuleSource(source string) bool {
	return strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || filepath.IsAbs(source)
}

func moduleInputRefs(body *hclsyntax.Body, inheritedInputs map[string][]string) map[string][]string {
	refs := make(map[string][]string)
	if body == nil {
		return refs
	}

	metaArgs := map[string]struct{}{
		"source":     {},
		"version":    {},
		"providers":  {},
		"depends_on": {},
		"count":      {},
		"for_each":   {},
	}

	for name, attr := range body.Attributes {
		if _, ok := metaArgs[name]; ok {
			continue
		}

		refs[name] = resolveModuleInputRefs(attr.Expr, inheritedInputs)
	}

	return refs
}

func resolveModuleInputRefs(expr hclsyntax.Expression, inheritedInputs map[string][]string) []string {
	locals, variables := extractReferenceSets(expr)
	resolved := make([]string, 0, len(locals)+len(variables))
	resolved = appendUnique(resolved, locals...)

	for _, variableRef := range variables {
		inputName := strings.TrimPrefix(variableRef, "var.")
		if inheritedRefs, ok := inheritedInputs[inputName]; ok && len(inheritedRefs) > 0 {
			resolved = appendUnique(resolved, inheritedRefs...)
			continue
		}
		resolved = appendUnique(resolved, variableRef)
	}

	sort.Strings(resolved)
	return resolved
}

func applyModuleInputs(resource *model.Resource, moduleInputs map[string][]string) {
	if len(moduleInputs) == 0 || len(resource.VariableRefs) == 0 {
		return
	}

	for _, variableRef := range resource.VariableRefs {
		inputName := strings.TrimPrefix(variableRef, "var.")
		inputRefs, ok := moduleInputs[inputName]
		if !ok || len(inputRefs) == 0 {
			continue
		}
		resource.ModuleInputs[inputName] = appendUnique(resource.ModuleInputs[inputName], inputRefs...)
	}
}

func collectResourceMetadata(body *hclsyntax.Body, resource *model.Resource, topLevel bool) {
	if body == nil {
		return
	}

	for name, attr := range body.Attributes {
		if topLevel && name == "depends_on" {
			resource.Dependencies = extractDependencies(attr.Expr)
		} else if topLevel {
			resource.Attributes[name] = exprToString(attr.Expr)
		}

		locals, variables := extractReferenceSets(attr.Expr)
		resource.LocalRefs = appendUnique(resource.LocalRefs, locals...)
		resource.VariableRefs = appendUnique(resource.VariableRefs, variables...)
	}

	for _, nestedBlock := range body.Blocks {
		collectResourceMetadata(nestedBlock.Body, resource, false)
	}
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

func extractReferenceSets(expr hclsyntax.Expression) ([]string, []string) {
	localRefs := make([]string, 0)
	variableRefs := make([]string, 0)
	collectExpressionReferences(expr, &localRefs, &variableRefs)
	return localRefs, variableRefs
}

func collectExpressionReferences(expr hclsyntax.Expression, localRefs, variableRefs *[]string) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		ref := traversalToString(e.Traversal)
		if strings.HasPrefix(ref, "local.") {
			*localRefs = appendUnique(*localRefs, ref)
		}
		if strings.HasPrefix(ref, "var.") {
			*variableRefs = appendUnique(*variableRefs, ref)
		}
	case *hclsyntax.RelativeTraversalExpr:
		collectExpressionReferences(e.Source, localRefs, variableRefs)
	case *hclsyntax.FunctionCallExpr:
		for _, arg := range e.Args {
			collectExpressionReferences(arg, localRefs, variableRefs)
		}
	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			collectExpressionReferences(part, localRefs, variableRefs)
		}
	case *hclsyntax.TemplateWrapExpr:
		collectExpressionReferences(e.Wrapped, localRefs, variableRefs)
	case *hclsyntax.TupleConsExpr:
		for _, elem := range e.Exprs {
			collectExpressionReferences(elem, localRefs, variableRefs)
		}
	case *hclsyntax.ObjectConsExpr:
		for _, item := range e.Items {
			collectExpressionReferences(item.KeyExpr, localRefs, variableRefs)
			collectExpressionReferences(item.ValueExpr, localRefs, variableRefs)
		}
	case *hclsyntax.BinaryOpExpr:
		collectExpressionReferences(e.LHS, localRefs, variableRefs)
		collectExpressionReferences(e.RHS, localRefs, variableRefs)
	case *hclsyntax.UnaryOpExpr:
		collectExpressionReferences(e.Val, localRefs, variableRefs)
	case *hclsyntax.ParenthesesExpr:
		collectExpressionReferences(e.Expression, localRefs, variableRefs)
	case *hclsyntax.ConditionalExpr:
		collectExpressionReferences(e.Condition, localRefs, variableRefs)
		collectExpressionReferences(e.TrueResult, localRefs, variableRefs)
		collectExpressionReferences(e.FalseResult, localRefs, variableRefs)
	case *hclsyntax.ForExpr:
		collectExpressionReferences(e.CollExpr, localRefs, variableRefs)
		if e.KeyExpr != nil {
			collectExpressionReferences(e.KeyExpr, localRefs, variableRefs)
		}
		collectExpressionReferences(e.ValExpr, localRefs, variableRefs)
		if e.CondExpr != nil {
			collectExpressionReferences(e.CondExpr, localRefs, variableRefs)
		}
	case *hclsyntax.IndexExpr:
		collectExpressionReferences(e.Collection, localRefs, variableRefs)
		collectExpressionReferences(e.Key, localRefs, variableRefs)
	case *hclsyntax.SplatExpr:
		collectExpressionReferences(e.Source, localRefs, variableRefs)
		if e.Each != nil {
			collectExpressionReferences(e.Each, localRefs, variableRefs)
		}
	case *hclsyntax.AnonSymbolExpr, *hclsyntax.LiteralValueExpr:
		return
	}
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}

	for _, addition := range additions {
		if addition == "" {
			continue
		}
		if _, ok := seen[addition]; ok {
			continue
		}
		seen[addition] = struct{}{}
		values = append(values, addition)
	}

	return values
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
