package ui

import (
	"fmt"
	"sort"
	"strings"
	"terraviz/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ColumnsView displays resources in Miller Columns (like macOS Finder)
type ColumnsView struct {
	styles           *Styles
	width            int
	height           int
	columns          [][]columnItem
	selectedColumn   int
	selectedRow      []int // Selected row index for each column
	dependents       map[string][]string
	resources        map[string]*model.Resource
	currentProvider  string
}

type columnItem struct {
	resource   *model.Resource
	childCount int
}

// NewColumnsView creates a new Miller Columns view
func NewColumnsView(styles *Styles) ColumnsView {
	return ColumnsView{
		styles:         styles,
		columns:        make([][]columnItem, 0),
		selectedRow:    make([]int, 0),
		dependents:     make(map[string][]string),
		resources:      make(map[string]*model.Resource),
		selectedColumn: 0,
	}
}

// SetSize updates the dimensions
func (c *ColumnsView) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetResources updates the view with new resources and builds dependency map
func (c *ColumnsView) SetResources(resources map[string]*model.Resource, edges []model.Edge) {
	c.resources = resources

	// We don't need a dependents map anymore - we'll use the Dependencies field directly
	c.dependents = make(map[string][]string)

	// Build initial column: all resources (we'll show dependencies when selected)
	c.buildInitialColumn()
}

// buildInitialColumn builds the first column with all resources
func (c *ColumnsView) buildInitialColumn() {
	allResources := make([]*model.Resource, 0)

	for _, res := range c.resources {
		allResources = append(allResources, res)
	}

	// Sort all resources
	sort.Slice(allResources, func(i, j int) bool {
		if allResources[i].Type != allResources[j].Type {
			return allResources[i].Type < allResources[j].Type
		}
		return allResources[i].Name < allResources[j].Name
	})

	// Build column items
	items := make([]columnItem, 0)
	for _, res := range allResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(res.Dependencies), // Number of dependencies
		})
	}

	c.columns = [][]columnItem{items}
	c.selectedRow = []int{0}
	c.selectedColumn = 0
}

// GetSelectedResource returns the currently selected resource
func (c *ColumnsView) GetSelectedResource() *model.Resource {
	if c.selectedColumn < 0 || c.selectedColumn >= len(c.columns) {
		return nil
	}
	if len(c.columns[c.selectedColumn]) == 0 {
		return nil
	}
	rowIdx := c.selectedRow[c.selectedColumn]
	if rowIdx < 0 || rowIdx >= len(c.columns[c.selectedColumn]) {
		return nil
	}
	return c.columns[c.selectedColumn][rowIdx].resource
}

// MoveUp moves selection up in current column
func (c *ColumnsView) MoveUp() {
	if c.selectedColumn < 0 || c.selectedColumn >= len(c.selectedRow) {
		return
	}
	if c.selectedRow[c.selectedColumn] > 0 {
		c.selectedRow[c.selectedColumn]--
	}
}

// MoveDown moves selection down in current column
func (c *ColumnsView) MoveDown() {
	if c.selectedColumn < 0 || c.selectedColumn >= len(c.columns) {
		return
	}
	maxRow := len(c.columns[c.selectedColumn]) - 1
	if c.selectedRow[c.selectedColumn] < maxRow {
		c.selectedRow[c.selectedColumn]++
	}
}

// MoveLeft moves to previous column
func (c *ColumnsView) MoveLeft() {
	if c.selectedColumn > 0 {
		c.selectedColumn--
		// Remove columns to the right
		c.columns = c.columns[:c.selectedColumn+1]
		c.selectedRow = c.selectedRow[:c.selectedColumn+1]
	}
}

// MoveRight moves to next column (expanding dependencies)
func (c *ColumnsView) MoveRight() {
	if c.selectedColumn >= len(c.columns) {
		return
	}

	selected := c.GetSelectedResource()
	if selected == nil {
		return
	}

	// Check if this resource has dependencies
	dependencies := selected.Dependencies
	if len(dependencies) == 0 {
		return
	}

	// Build next column with dependencies
	depResources := make([]*model.Resource, 0)
	for _, depID := range dependencies {
		if res, ok := c.resources[depID]; ok {
			depResources = append(depResources, res)
		}
	}

	// Sort dependencies
	sort.Slice(depResources, func(i, j int) bool {
		if depResources[i].Type != depResources[j].Type {
			return depResources[i].Type < depResources[j].Type
		}
		return depResources[i].Name < depResources[j].Name
	})

	// Build column items
	items := make([]columnItem, 0)
	for _, res := range depResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(res.Dependencies), // Number of dependencies this has
		})
	}

	// Add or replace the next column
	if c.selectedColumn+1 < len(c.columns) {
		c.columns[c.selectedColumn+1] = items
		c.selectedRow[c.selectedColumn+1] = 0
		// Remove columns beyond this
		c.columns = c.columns[:c.selectedColumn+2]
		c.selectedRow = c.selectedRow[:c.selectedColumn+2]
	} else {
		c.columns = append(c.columns, items)
		c.selectedRow = append(c.selectedRow, 0)
	}

	c.selectedColumn++
}

// HandleSelection rebuilds columns when selection changes
func (c *ColumnsView) HandleSelection() {
	// Rebuild next column based on current selection
	selected := c.GetSelectedResource()
	if selected == nil {
		return
	}

	dependencies := selected.Dependencies
	if len(dependencies) == 0 {
		// No dependencies - remove columns to the right
		c.columns = c.columns[:c.selectedColumn+1]
		c.selectedRow = c.selectedRow[:c.selectedColumn+1]
		return
	}

	// Build next column with dependencies
	depResources := make([]*model.Resource, 0)
	for _, depID := range dependencies {
		if res, ok := c.resources[depID]; ok {
			depResources = append(depResources, res)
		}
	}

	sort.Slice(depResources, func(i, j int) bool {
		if depResources[i].Type != depResources[j].Type {
			return depResources[i].Type < depResources[j].Type
		}
		return depResources[i].Name < depResources[j].Name
	})

	items := make([]columnItem, 0)
	for _, res := range depResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(res.Dependencies),
		})
	}

	// Update or add next column
	if c.selectedColumn+1 < len(c.columns) {
		c.columns[c.selectedColumn+1] = items
		if c.selectedRow[c.selectedColumn+1] >= len(items) {
			c.selectedRow[c.selectedColumn+1] = len(items) - 1
		}
		if c.selectedRow[c.selectedColumn+1] < 0 {
			c.selectedRow[c.selectedColumn+1] = 0
		}
		// Remove columns beyond
		c.columns = c.columns[:c.selectedColumn+2]
		c.selectedRow = c.selectedRow[:c.selectedColumn+2]
	} else {
		c.columns = append(c.columns, items)
		c.selectedRow = append(c.selectedRow, 0)
	}
}

// Update handles messages
func (c *ColumnsView) Update(msg tea.Msg) (ColumnsView, tea.Cmd) {
	return *c, nil
}

// View renders the Miller Columns
func (c *ColumnsView) View() string {
	if len(c.columns) == 0 {
		return c.styles.DetailValue.Render("No resources to display")
	}

	// Calculate column width (divide available width by number of visible columns)
	visibleColumns := len(c.columns)
	if visibleColumns > 4 {
		visibleColumns = 4 // Show max 4 columns
	}

	columnWidth := (c.width - 8) / visibleColumns
	if columnWidth < 20 {
		columnWidth = 20
	}

	var columns []string

	// Determine which columns to show (show current column and context)
	startCol := c.selectedColumn
	if startCol > len(c.columns)-visibleColumns && len(c.columns) >= visibleColumns {
		startCol = len(c.columns) - visibleColumns
	}
	if startCol < 0 {
		startCol = 0
	}

	for colIdx := startCol; colIdx < len(c.columns) && colIdx < startCol+visibleColumns; colIdx++ {
		columns = append(columns, c.renderColumn(colIdx, columnWidth))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

// renderColumn renders a single column
func (c *ColumnsView) renderColumn(colIdx int, width int) string {
	var lines []string

	// Column header
	header := "Dependencies"
	if colIdx == 0 {
		header = "All Resources"
	} else if colIdx == 1 {
		header = "Depends On"
	}

	headerStyle := c.styles.ModuleHeader.Width(width - 2).Align(lipgloss.Center)
	lines = append(lines, headerStyle.Render(header))
	lines = append(lines, strings.Repeat("─", width))

	// Column items
	items := c.columns[colIdx]
	selectedRow := c.selectedRow[colIdx]
	isFocused := colIdx == c.selectedColumn

	availableHeight := c.height - 4 // Reserve space for header and borders
	startRow := 0
	if selectedRow >= availableHeight {
		startRow = selectedRow - availableHeight + 1
	}

	for i := startRow; i < len(items) && i < startRow+availableHeight; i++ {
		item := items[i]
		isSelected := i == selectedRow && isFocused

		// Build line
		resourceText := fmt.Sprintf("%s.%s", item.resource.Type, item.resource.Name)
		if len(resourceText) > width-6 {
			resourceText = resourceText[:width-6] + "..."
		}

		// Add child indicator
		suffix := ""
		if item.childCount > 0 {
			suffix = fmt.Sprintf(" (%d)", item.childCount)
		}

		line := resourceText + suffix

		// Apply styling
		providerStyle := c.styles.GetProviderStyle(item.resource.Provider)

		if isSelected {
			line = c.styles.SelectedRow.Width(width - 2).Render(line)
		} else {
			line = providerStyle.Width(width - 2).Render(line)
		}

		lines = append(lines, line)
	}

	// Fill remaining space
	for len(lines) < c.height-2 {
		lines = append(lines, strings.Repeat(" ", width))
	}

	// Build column with border
	borderStyle := c.styles.UnfocusedBorder
	if isFocused {
		borderStyle = c.styles.FocusedBorder
	}

	return borderStyle.
		Width(width).
		Height(c.height - 2).
		Render(strings.Join(lines, "\n"))
}
