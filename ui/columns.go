package ui

import (
	"fmt"
	"sort"
	"strings"
	"terraviz/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const minColumnWidth = 28
const maxVisibleColumns = 3

// ColumnsView displays resources in Miller Columns (like macOS Finder)
type ColumnsView struct {
	styles          *Styles
	width           int
	height          int
	columns         [][]columnItem
	selectedColumn  int
	selectedRow     []int // Selected row index for each column
	dependents      map[string][]string
	resources       map[string]*model.Resource
	currentProvider string
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

	c.dependents = make(map[string][]string)
	for _, edge := range edges {
		c.dependents[edge.To] = append(c.dependents[edge.To], edge.From)
	}

	// Build initial column from root resources and expand to dependents.
	c.buildInitialColumn()
}

// buildInitialColumn builds the first column with root resources
func (c *ColumnsView) buildInitialColumn() {
	rootResources := make([]*model.Resource, 0)

	for _, res := range c.resources {
		if len(res.Dependencies) == 0 {
			rootResources = append(rootResources, res)
		}
	}

	if len(rootResources) == 0 {
		for _, res := range c.resources {
			rootResources = append(rootResources, res)
		}
	}

	// Sort root resources
	sort.Slice(rootResources, func(i, j int) bool {
		if rootResources[i].Type != rootResources[j].Type {
			return rootResources[i].Type < rootResources[j].Type
		}
		return rootResources[i].Name < rootResources[j].Name
	})

	// Build column items
	items := make([]columnItem, 0)
	for _, res := range rootResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(c.dependents[res.ID]),
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

	children := c.dependents[selected.ID]
	if len(children) == 0 {
		return
	}

	childResources := make([]*model.Resource, 0, len(children))
	for _, childID := range children {
		if res, ok := c.resources[childID]; ok {
			childResources = append(childResources, res)
		}
	}

	// Sort dependents
	sort.Slice(childResources, func(i, j int) bool {
		if childResources[i].Type != childResources[j].Type {
			return childResources[i].Type < childResources[j].Type
		}
		return childResources[i].Name < childResources[j].Name
	})

	// Build column items
	items := make([]columnItem, 0)
	for _, res := range childResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(c.dependents[res.ID]),
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

	children := c.dependents[selected.ID]
	if len(children) == 0 {
		// No dependencies - remove columns to the right
		c.columns = c.columns[:c.selectedColumn+1]
		c.selectedRow = c.selectedRow[:c.selectedColumn+1]
		return
	}

	// Build next column with dependents
	childResources := make([]*model.Resource, 0, len(children))
	for _, childID := range children {
		if res, ok := c.resources[childID]; ok {
			childResources = append(childResources, res)
		}
	}

	sort.Slice(childResources, func(i, j int) bool {
		if childResources[i].Type != childResources[j].Type {
			return childResources[i].Type < childResources[j].Type
		}
		return childResources[i].Name < childResources[j].Name
	})

	items := make([]columnItem, 0)
	for _, res := range childResources {
		items = append(items, columnItem{
			resource:   res,
			childCount: len(c.dependents[res.ID]),
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

	availableWidth := c.width
	visibleColumns := c.visibleColumnCount(availableWidth)

	columnWidth := availableWidth / visibleColumns
	if columnWidth < 1 {
		columnWidth = 1
	}

	var columns []string

	startCol := c.windowStart(visibleColumns)

	columns = columns[:0]
	for colIdx := startCol; colIdx < len(c.columns) && colIdx < startCol+visibleColumns; colIdx++ {
		columns = append(columns, c.renderColumn(colIdx, columnWidth, c.height))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (c *ColumnsView) visibleColumnCount(availableWidth int) int {
	visibleColumns := availableWidth / minColumnWidth
	if visibleColumns < 1 {
		visibleColumns = 1
	}
	if visibleColumns > maxVisibleColumns {
		visibleColumns = maxVisibleColumns
	}
	if visibleColumns > len(c.columns) {
		visibleColumns = len(c.columns)
	}
	if visibleColumns == 0 {
		visibleColumns = 1
	}

	return visibleColumns
}

func (c *ColumnsView) BreadcrumbHeight() int {
	if len(c.columns) == 0 || c.height <= 6 {
		return 0
	}

	visibleColumns := c.visibleColumnCount(c.width)
	startCol := c.windowStart(visibleColumns)
	return lipgloss.Height(c.renderBreadcrumb(startCol, visibleColumns))
}

func (c *ColumnsView) Breadcrumb() string {
	if len(c.columns) == 0 || c.height <= 6 {
		return ""
	}

	visibleColumns := c.visibleColumnCount(c.width)
	startCol := c.windowStart(visibleColumns)
	return c.renderBreadcrumb(startCol, visibleColumns)
}

func (c *ColumnsView) windowStart(visibleColumns int) int {
	if visibleColumns <= 0 {
		return 0
	}

	startCol := c.selectedColumn - visibleColumns + 1
	if startCol < 0 {
		startCol = 0
	}

	maxStart := len(c.columns) - visibleColumns
	if maxStart < 0 {
		maxStart = 0
	}
	if startCol > maxStart {
		startCol = maxStart
	}

	return startCol
}

func (c *ColumnsView) renderBreadcrumb(startCol, visibleColumns int) string {
	segments := make([]string, 0, visibleColumns+2)
	endCol := startCol + visibleColumns
	if endCol > len(c.columns) {
		endCol = len(c.columns)
	}

	if startCol > 0 {
		segments = append(segments, c.styles.BreadcrumbDim.Render(fmt.Sprintf("... %d earlier", startCol)))
	}

	for colIdx := startCol; colIdx < endCol; colIdx++ {
		label := c.columnLabel(colIdx)
		style := c.styles.BreadcrumbDim
		if colIdx == c.selectedColumn {
			style = c.styles.Breadcrumb
		}
		segments = append(segments, style.Render(label))
	}

	if endCol < len(c.columns) {
		segments = append(segments, c.styles.BreadcrumbDim.Render(fmt.Sprintf("%d more ...", len(c.columns)-endCol)))
	}

	if len(segments) == 0 {
		segments = append(segments, c.styles.BreadcrumbDim.Render("No columns"))
	}

	row := strings.Join(segments, " ")
	return lipgloss.NewStyle().Width(c.width).MaxWidth(c.width).Render(truncateLabel(row, c.width))
}

func (c *ColumnsView) columnLabel(colIdx int) string {
	switch colIdx {
	case 0:
		return "Top Level"
	case 1:
		return "Children"
	default:
		return fmt.Sprintf("Level %d", colIdx)
	}
}

// renderColumn renders a single column
func (c *ColumnsView) renderColumn(colIdx int, width, height int) string {
	borderStyle := c.styles.UnfocusedBorder
	if colIdx == c.selectedColumn {
		borderStyle = c.styles.FocusedBorder
	}

	// Column header
	header := c.columnLabel(colIdx)

	innerWidth := width - borderStyle.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	headerStyle := c.styles.ModuleHeader.Width(innerWidth).Align(lipgloss.Center)
	headerBlock := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(truncateLabel(header, innerWidth)),
		truncateLabel(strings.Repeat("─", innerWidth), innerWidth),
	)

	// Column items
	items := c.columns[colIdx]
	selectedRow := c.selectedRow[colIdx]
	isFocused := colIdx == c.selectedColumn

	innerHeight := height - borderStyle.GetVerticalFrameSize()
	if innerHeight < 1 {
		innerHeight = 1
	}

	headHeight := lipgloss.Height(headerBlock)
	bodyHeight := innerHeight - headHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	startRow := c.scrollOffset(selectedRow, len(items), bodyHeight)
	bodyLines := make([]string, 0, bodyHeight)

	for i := startRow; i < len(items) && i < startRow+bodyHeight; i++ {
		item := items[i]
		isSelected := i == selectedRow && isFocused

		// Add child indicator
		suffix := ""
		if item.childCount > 0 {
			suffix = fmt.Sprintf(" (%d)", item.childCount)
		}

		resourceText := truncateLabel(fmt.Sprintf("%s.%s", item.resource.Type, item.resource.Name), innerWidth-lipgloss.Width(suffix))

		line := resourceText + suffix

		// Apply styling
		providerStyle := c.styles.GetProviderStyle(item.resource.Provider)

		if isSelected {
			line = c.styles.SelectedRow.Width(innerWidth).Render(line)
		} else {
			line = providerStyle.Width(innerWidth).Render(line)
		}

		bodyLines = append(bodyLines, line)
	}

	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, lipgloss.NewStyle().Width(innerWidth).Render(""))
	}

	bodyBlock := lipgloss.JoinVertical(lipgloss.Left, bodyLines...)
	content := lipgloss.JoinVertical(lipgloss.Left, headerBlock, bodyBlock)

	return borderedBox(borderStyle, width, height, content)
}

func (c *ColumnsView) scrollOffset(selectedRow, itemCount, bodyHeight int) int {
	if bodyHeight <= 0 || itemCount <= bodyHeight || selectedRow < bodyHeight {
		return 0
	}

	startRow := selectedRow - bodyHeight + 1
	maxStart := itemCount - bodyHeight
	if maxStart < 0 {
		maxStart = 0
	}
	if startRow > maxStart {
		startRow = maxStart
	}
	if startRow < 0 {
		startRow = 0
	}

	return startRow
}

func truncateLabel(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	if lipgloss.Width(value) <= maxWidth {
		return value
	}

	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	return value[:maxWidth-3] + "..."
}
