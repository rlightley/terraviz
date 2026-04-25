package ui

import "github.com/charmbracelet/lipgloss"

func borderedBox(style lipgloss.Style, totalWidth, totalHeight int, content string) string {
	innerWidth := totalWidth - style.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	innerHeight := totalHeight - style.GetVerticalFrameSize()
	if innerHeight < 1 {
		innerHeight = 1
	}

	framedContent := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)
	return style.Width(innerWidth).Height(innerHeight).Render(framedContent)
}

// Styles holds all the styling for the TUI
type Styles struct {
	// Panel styles
	BorderStyle     lipgloss.Style
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
	SelectedRow     lipgloss.Style
	UnselectedRow   lipgloss.Style
	ModuleHeader    lipgloss.Style
	Breadcrumb      lipgloss.Style
	BreadcrumbDim   lipgloss.Style

	// Provider colors
	ProviderAWS     lipgloss.Style
	ProviderGoogle  lipgloss.Style
	ProviderAzure   lipgloss.Style
	ProviderDefault lipgloss.Style

	// Detail view styles
	DetailHeader lipgloss.Style
	DetailKey    lipgloss.Style
	DetailValue  lipgloss.Style
	SectionTitle lipgloss.Style

	// Status bar styles
	StatusBar        lipgloss.Style
	StatusBarKey     lipgloss.Style
	StatusBarWarning lipgloss.Style
}

// NewStyles creates and initializes all styles
func NewStyles() *Styles {
	// Define colors
	mutedGray := lipgloss.Color("243")
	highlightBg := lipgloss.Color("238")
	selectedBg := lipgloss.Color("31")

	awsOrange := lipgloss.Color("214")
	googleBlue := lipgloss.Color("69")
	azureTeal := lipgloss.Color("38")
	defaultGray := lipgloss.Color("246")

	statusBg := lipgloss.Color("236")
	statusFg := lipgloss.Color("254")
	warningYellow := lipgloss.Color("220")

	return &Styles{
		BorderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedGray),

		FocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(selectedBg),

		UnfocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedGray),

		SelectedRow: lipgloss.NewStyle().
			Background(selectedBg).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1),

		UnselectedRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		ModuleHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("222")).
			MarginTop(0).
			MarginBottom(0),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")).
			Padding(0, 1),

		BreadcrumbDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),

		ProviderAWS: lipgloss.NewStyle().
			Foreground(awsOrange).
			Bold(true),

		ProviderGoogle: lipgloss.NewStyle().
			Foreground(googleBlue).
			Bold(true),

		ProviderAzure: lipgloss.NewStyle().
			Foreground(azureTeal).
			Bold(true),

		ProviderDefault: lipgloss.NewStyle().
			Foreground(defaultGray),

		DetailHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("223")).
			MarginBottom(1).
			Padding(0, 1),

		DetailKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true),

		DetailValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("186")).
			MarginTop(1).
			MarginBottom(0),

		StatusBar: lipgloss.NewStyle().
			Background(statusBg).
			Foreground(statusFg).
			Padding(0, 1),

		StatusBarKey: lipgloss.NewStyle().
			Background(highlightBg).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1).
			MarginRight(1),

		StatusBarWarning: lipgloss.NewStyle().
			Background(statusBg).
			Foreground(warningYellow).
			Bold(true),
	}
}

// GetProviderStyle returns the appropriate style for a given provider
func (s *Styles) GetProviderStyle(provider string) lipgloss.Style {
	switch provider {
	case "aws":
		return s.ProviderAWS
	case "google", "gcp":
		return s.ProviderGoogle
	case "azurerm", "azure":
		return s.ProviderAzure
	default:
		return s.ProviderDefault
	}
}
