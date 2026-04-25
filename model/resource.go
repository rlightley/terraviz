package model

// Resource represents a Terraform resource with its attributes and dependencies
type Resource struct {
	ID           string
	Type         string
	Name         string
	Provider     string
	Module       string
	Attributes   map[string]any
	Dependencies []string
	LocalRefs    []string
	VariableRefs []string
	ModuleInputs map[string][]string
}

// Edge represents a directed edge between two resources in the dependency graph
type Edge struct {
	From string
	To   string
}

// Graph represents the complete infrastructure graph
type Graph struct {
	Resources map[string]*Resource
	Edges     []Edge
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		Resources: make(map[string]*Resource),
		Edges:     make([]Edge, 0),
	}
}

// AddResource adds a resource to the graph
func (g *Graph) AddResource(r *Resource) {
	g.Resources[r.ID] = r
}

// AddEdge adds a dependency edge to the graph
func (g *Graph) AddEdge(from, to string) {
	g.Edges = append(g.Edges, Edge{From: from, To: to})
}
