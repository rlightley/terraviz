package ui

import (
	"fmt"
	"strings"
	"terraviz/model"
	"terraviz/parser"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FocusPanel represents which panel is currently focused
type FocusPanel int

const (
	FocusColumns FocusPanel = iota
	FocusDetail
)

// Model represents the main application model
type Model struct {
	columnsView      ColumnsView
	detail           DetailView
	styles           *Styles
	graph            *model.Graph
	allResources     map[string]*model.Resource
	allEdges         []model.Edge
	parseResult      *parser.ParseResult
	focus            FocusPanel
	width            int
	height           int
	providers        []string // List of available providers
	providerIndex    int      // Current provider tab index
}

// NewModel creates a new application model
func NewModel(parseResult *parser.ParseResult) Model {
	styles := NewStyles()

	// Extract list of providers
	providerSet := make(map[string]bool)
	for _, res := range parseResult.Graph.Resources {
		if res.Provider != "" {
			providerSet[res.Provider] = true
		}
	}

	// Build sorted provider list
	providers := make([]string, 0, len(providerSet))
	for provider := range providerSet {
		providers = append(providers, provider)
	}

	// Sort providers alphabetically
	for i := 0; i < len(providers); i++ {
		for j := i + 1; j < len(providers); j++ {
			if providers[i] > providers[j] {
				providers[i], providers[j] = providers[j], providers[i]
			}
		}
	}

	// Default to first provider if available
	startIndex := 0
	if len(providers) == 0 {
		providers = []string{"all"}
	}

	m := Model{
		columnsView:   NewColumnsView(styles),
		detail:        NewDetailView(styles),
		styles:        styles,
		graph:         parseResult.Graph,
		allResources:  parseResult.Graph.Resources,
		allEdges:      parseResult.Graph.Edges,
		parseResult:   parseResult,
		focus:         FocusColumns,
		providers:     providers,
		providerIndex: startIndex,
	}

	m.updateColumnsView()
	m.updateDetail()

	return m
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.focus == FocusColumns {
				m.columnsView.MoveUp()
				m.columnsView.HandleSelection()
				m.updateDetail()
			} else {
				var cmd tea.Cmd
				m.detail, cmd = m.detail.Update(msg)
				return m, cmd
			}

		case "down", "j":
			if m.focus == FocusColumns {
				m.columnsView.MoveDown()
				m.columnsView.HandleSelection()
				m.updateDetail()
			} else {
				var cmd tea.Cmd
				m.detail, cmd = m.detail.Update(msg)
				return m, cmd
			}

		case "left", "h":
			if m.focus == FocusColumns {
				m.columnsView.MoveLeft()
				m.updateDetail()
			} else {
				m.focus = FocusColumns
			}

		case "right", "l":
			if m.focus == FocusColumns {
				m.columnsView.MoveRight()
				m.updateDetail()
			} else {
				m.focus = FocusDetail
			}

		case "tab":
			if m.focus == FocusColumns {
				m.focus = FocusDetail
			} else {
				m.focus = FocusColumns
			}

		case "n", "]":
			// Next provider tab
			m.nextProvider()

		case "p", "[":
			// Previous provider tab
			m.previousProvider()

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Jump to specific provider tab by number
			idx := int(msg.String()[0] - '1')
			if idx < len(m.providers) {
				m.providerIndex = idx
				m.updateColumnsView()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
	}

	return m, nil
}

// updateColumnsView updates the columns view with current provider filter
func (m *Model) updateColumnsView() {
	resources, edges := m.getCurrentResourcesAndEdges()
	m.columnsView.SetResources(resources, edges)
	m.updateDetail()
}

// nextProvider switches to the next provider tab
func (m *Model) nextProvider() {
	if len(m.providers) == 0 {
		return
	}

	m.providerIndex++
	if m.providerIndex >= len(m.providers) {
		m.providerIndex = 0 // Wrap around
	}

	m.updateColumnsView()
}

// previousProvider switches to the previous provider tab
func (m *Model) previousProvider() {
	if len(m.providers) == 0 {
		return
	}

	m.providerIndex--
	if m.providerIndex < 0 {
		m.providerIndex = len(m.providers) - 1 // Wrap around
	}

	m.updateColumnsView()
}

// getCurrentResourcesAndEdges returns the current resource set and edges based on provider
func (m *Model) getCurrentResourcesAndEdges() (map[string]*model.Resource, []model.Edge) {
	if m.providerIndex < 0 || m.providerIndex >= len(m.providers) {
		return m.allResources, m.allEdges
	}

	currentProvider := m.providers[m.providerIndex]
	filtered := make(map[string]*model.Resource)

	// Filter by current provider
	for id, res := range m.allResources {
		if res.Provider == currentProvider {
			filtered[id] = res
		}
	}

	// Filter edges to only include those between resources in the filtered set
	filteredEdges := make([]model.Edge, 0)
	for _, edge := range m.allEdges {
		if _, hasFrom := filtered[edge.From]; hasFrom {
			if _, hasTo := filtered[edge.To]; hasTo {
				filteredEdges = append(filteredEdges, edge)
			}
		}
	}

	return filtered, filteredEdges
}

// updateDetail updates the detail view with the currently selected resource
func (m *Model) updateDetail() {
	selected := m.columnsView.GetSelectedResource()
	m.detail.SetResource(selected)
}

// layoutPanels calculates and sets the sizes of the panels
func (m *Model) layoutPanels() {
	// Reserve space for provider tabs (2 lines) and status bar (1 line)
	tabsHeight := 2
	statusBarHeight := 1
	availableHeight := m.height - tabsHeight - statusBarHeight

	// Split width: 70% columns, 30% detail
	columnsWidth := m.width * 70 / 100
	detailWidth := m.width - columnsWidth

	m.columnsView.SetSize(columnsWidth, availableHeight)
	m.detail.SetSize(detailWidth, availableHeight)
}

// View renders the application
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render provider tabs
	tabs := m.renderProviderTabs()

	// Render panels
	columnsPanel := m.columnsView.View()
	detailPanel := m.renderDetailPanel()

	// Combine panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, columnsPanel, detailPanel)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Combine everything
	return lipgloss.JoinVertical(lipgloss.Left, tabs, panels, statusBar)
}

// renderProviderTabs renders the provider tabs at the top
func (m Model) renderProviderTabs() string {
	var tabs []string

	for i, provider := range m.providers {
		isActive := i == m.providerIndex

		providerName := strings.ToUpper(provider)
		tabStyle := lipgloss.NewStyle().
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true)

		if isActive {
			tabStyle = tabStyle.
				BorderForeground(m.styles.GetProviderStyle(provider).GetForeground()).
				Foreground(m.styles.GetProviderStyle(provider).GetForeground()).
				Bold(true)
		} else {
			tabStyle = tabStyle.
				BorderForeground(lipgloss.Color("240")).
				Foreground(lipgloss.Color("246"))
		}

		tabs = append(tabs, tabStyle.Render(providerName))
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Add separator line
	separator := strings.Repeat("─", m.width)

	return lipgloss.JoinVertical(lipgloss.Left, tabsRow, separator)
}

// renderDetailPanel renders the detail panel with border
func (m Model) renderDetailPanel() string {
	borderStyle := m.styles.UnfocusedBorder
	if m.focus == FocusDetail {
		borderStyle = m.styles.FocusedBorder
	}

	return borderStyle.
		Width(m.width*60/100 - 2).
		Height(m.height - 3).
		Render(m.detail.View())
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	var parts []string

	// Mode indicator
	mode := string(m.parseResult.Mode)
	parts = append(parts, fmt.Sprintf("Mode: %s", mode))

	// Resource count
	resources, _ := m.getCurrentResourcesAndEdges()
	parts = append(parts, fmt.Sprintf("Resources: %d", len(resources)))

	// Failed files
	if len(m.parseResult.FailedFiles) > 0 {
		failedMsg := fmt.Sprintf("Failed: %s", strings.Join(m.parseResult.FailedFiles, ", "))
		parts = append(parts, m.styles.StatusBarWarning.Render(failedMsg))
	}

	leftSide := strings.Join(parts, " │ ")

	// Keybindings
	keys := []string{
		m.styles.StatusBarKey.Render("↑↓") + "navigate",
		m.styles.StatusBarKey.Render("←→") + "columns",
		m.styles.StatusBarKey.Render("Tab") + "detail",
		m.styles.StatusBarKey.Render("N/P") + "tabs",
		m.styles.StatusBarKey.Render("Q") + "quit",
	}
	rightSide := strings.Join(keys, " ")

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(rightSide)
	padding := m.width - leftWidth - rightWidth - 4

	if padding < 0 {
		padding = 0
	}

	statusContent := leftSide + strings.Repeat(" ", padding) + rightSide

	return m.styles.StatusBar.
		Width(m.width).
		Render(statusContent)
}
