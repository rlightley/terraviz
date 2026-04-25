package ui

import (
	"fmt"
	"sort"
	"strings"
	"terraviz/model"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TreeView displays resources in a hierarchical tree based on dependency layers
type TreeView struct {
	viewport viewport.Model
	styles   *Styles
	width    int
	height   int
	lines    []treeLine
	selected int
}

type treeLine struct {
	text       string
	resource   *model.Resource
	isHeader   bool
	layerDepth int
}

// NewTreeView creates a new tree view
func NewTreeView(styles *Styles) TreeView {
	vp := viewport.New(0, 0)
	return TreeView{
		viewport: vp,
		styles:   styles,
		lines:    make([]treeLine, 0),
		selected: 0,
	}
}

// SetSize updates the dimensions of the tree view
func (t *TreeView) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.viewport.Width = width - 4
	t.viewport.Height = height - 2
	t.render()
}

// SetResources updates the tree with a new set of resources
func (t *TreeView) SetResources(resources map[string]*model.Resource, edges []model.Edge) {
	// Calculate dependency depths
	depthMap := calculateDepths(resources, edges)

	// Group resources by depth
	layers := make(map[int][]*model.Resource)
	maxDepth := 0

	for _, res := range resources {
		depth := depthMap[res.ID]
		layers[depth] = append(layers[depth], res)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Sort resources within each layer
	for depth := range layers {
		sort.Slice(layers[depth], func(i, j int) bool {
			if layers[depth][i].Provider != layers[depth][j].Provider {
				return layers[depth][i].Provider < layers[depth][j].Provider
			}
			if layers[depth][i].Type != layers[depth][j].Type {
				return layers[depth][i].Type < layers[depth][j].Type
			}
			return layers[depth][i].Name < layers[depth][j].Name
		})
	}

	// Build tree lines
	t.lines = make([]treeLine, 0)

	for depth := 0; depth <= maxDepth; depth++ {
		layerResources := layers[depth]
		if len(layerResources) == 0 {
			continue
		}

		// Layer header
		headerText := fmt.Sprintf("├─ Layer %d", depth)
		if depth == 0 {
			headerText += " (No Dependencies)"
		} else if depth == 1 {
			headerText += " (1 step from root)"
		} else {
			headerText += fmt.Sprintf(" (%d steps from root)", depth)
		}

		t.lines = append(t.lines, treeLine{
			text:       headerText,
			isHeader:   true,
			layerDepth: depth,
		})

		// Resources in this layer
		for i, res := range layerResources {
			isLast := i == len(layerResources)-1

			// Build resource line
			prefix := "│  ├─ "
			if isLast {
				prefix = "│  └─ "
			}

			// Resource type and name with color
			resText := fmt.Sprintf("%s%s.%s", prefix, res.Type, res.Name)

			// Add dependencies if any
			if len(res.Dependencies) > 0 {
				depNames := make([]string, 0, len(res.Dependencies))
				for _, depID := range res.Dependencies {
					if dep, ok := resources[depID]; ok {
						depNames = append(depNames, fmt.Sprintf("%s.%s", dep.Type, dep.Name))
					} else {
						depNames = append(depNames, depID)
					}
				}
				resText += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" → " + strings.Join(depNames, ", "))
			}

			t.lines = append(t.lines, treeLine{
				text:       resText,
				resource:   res,
				isHeader:   false,
				layerDepth: depth,
			})
		}

		// Blank line between layers
		if depth < maxDepth {
			t.lines = append(t.lines, treeLine{
				text:     "│",
				isHeader: true,
			})
		}
	}

	// Ensure selected index is valid
	if t.selected >= len(t.lines) {
		t.selected = len(t.lines) - 1
	}
	if t.selected < 0 && len(t.lines) > 0 {
		t.selected = 0
	}

	// Find first non-header line for initial selection
	if t.selected == 0 {
		for i, line := range t.lines {
			if !line.isHeader && line.resource != nil {
				t.selected = i
				break
			}
		}
	}

	t.render()
}

// calculateDepths assigns a depth level to each resource based on dependencies
func calculateDepths(resources map[string]*model.Resource, edges []model.Edge) map[string]int {
	depths := make(map[string]int)

	// Build reverse dependency map
	dependents := make(map[string][]string)
	for _, edge := range edges {
		dependents[edge.To] = append(dependents[edge.To], edge.From)
	}

	// Start with resources that have no dependencies
	queue := make([]string, 0)
	for id, res := range resources {
		if len(res.Dependencies) == 0 {
			depths[id] = 0
			queue = append(queue, id)
		}
	}

	// BFS to assign depths
	processed := make(map[string]bool)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if processed[current] {
			continue
		}
		processed[current] = true

		currentDepth := depths[current]

		// Update all resources that depend on current
		for _, dependent := range dependents[current] {
			newDepth := currentDepth + 1
			if existingDepth, ok := depths[dependent]; !ok || newDepth > existingDepth {
				depths[dependent] = newDepth
			}
			queue = append(queue, dependent)
		}
	}

	// Assign depth to any remaining resources
	for id := range resources {
		if _, ok := depths[id]; !ok {
			depths[id] = 0
		}
	}

	return depths
}

// GetSelectedResource returns the currently selected resource
func (t *TreeView) GetSelectedResource() *model.Resource {
	if t.selected < 0 || t.selected >= len(t.lines) {
		return nil
	}
	return t.lines[t.selected].resource
}

// MoveUp moves the selection up
func (t *TreeView) MoveUp() {
	if t.selected > 0 {
		t.selected--
		// Skip headers
		for t.selected > 0 && t.lines[t.selected].isHeader {
			t.selected--
		}
		t.ensureVisible()
		t.render()
	}
}

// MoveDown moves the selection down
func (t *TreeView) MoveDown() {
	if t.selected < len(t.lines)-1 {
		t.selected++
		// Skip headers
		for t.selected < len(t.lines)-1 && t.lines[t.selected].isHeader {
			t.selected++
		}
		t.ensureVisible()
		t.render()
	}
}

// ensureVisible scrolls to ensure the selected line is visible
func (t *TreeView) ensureVisible() {
	if t.selected < t.viewport.YOffset {
		t.viewport.YOffset = t.selected
	} else if t.selected >= t.viewport.YOffset+t.viewport.Height {
		t.viewport.YOffset = t.selected - t.viewport.Height + 1
	}
}

// render draws the tree
func (t *TreeView) render() {
	if len(t.lines) == 0 {
		t.viewport.SetContent(t.styles.DetailValue.Render("No resources to display"))
		return
	}

	var content strings.Builder

	for i, line := range t.lines {
		isSelected := i == t.selected

		if line.isHeader {
			// Render header line
			content.WriteString(t.styles.ModuleHeader.Render(line.text))
		} else {
			// Render resource line with provider color
			if line.resource != nil {
				providerStyle := t.styles.GetProviderStyle(line.resource.Provider)

				// Extract parts of the line (prefix, type.name, dependencies)
				parts := strings.SplitN(line.text, " ", 2)
				if len(parts) == 2 {
					prefix := parts[0]
					rest := parts[1]

					// Split resource from dependencies
					resParts := strings.Split(rest, " → ")
					resourcePart := resParts[0]

					styledLine := t.styles.UnselectedRow.Render(prefix + " ") +
						providerStyle.Render(resourcePart)

					if len(resParts) > 1 {
						styledLine += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" → " + resParts[1])
					}

					if isSelected {
						styledLine = t.styles.SelectedRow.Render(styledLine)
					}

					content.WriteString(styledLine)
				} else {
					if isSelected {
						content.WriteString(t.styles.SelectedRow.Render(line.text))
					} else {
						content.WriteString(t.styles.UnselectedRow.Render(line.text))
					}
				}
			} else {
				if isSelected {
					content.WriteString(t.styles.SelectedRow.Render(line.text))
				} else {
					content.WriteString(t.styles.UnselectedRow.Render(line.text))
				}
			}
		}

		content.WriteString("\n")
	}

	t.viewport.SetContent(content.String())
}

// Update handles messages for the tree view
func (t *TreeView) Update(msg tea.Msg) (TreeView, tea.Cmd) {
	var cmd tea.Cmd
	t.viewport, cmd = t.viewport.Update(msg)
	return *t, cmd
}

// View renders the tree view
func (t *TreeView) View() string {
	return t.viewport.View()
}
