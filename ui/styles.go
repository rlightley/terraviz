package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the styling for the TUI
type Styles struct {
	// Panel styles
	BorderStyle      lipgloss.Style
	FocusedBorder    lipgloss.Style
	UnfocusedBorder  lipgloss.Style
	SelectedRow      lipgloss.Style
	UnselectedRow    lipgloss.Style
	ModuleHeader     lipgloss.Style

	// Provider colors
	ProviderAWS      lipgloss.Style
	ProviderGoogle   lipgloss.Style
	ProviderAzure    lipgloss.Style
	ProviderDefault  lipgloss.Style

	// Detail view styles
	DetailHeader     lipgloss.Style
	DetailKey        lipgloss.Style
	DetailValue      lipgloss.Style
	SectionTitle     lipgloss.Style

	// Status bar styles
	StatusBar        lipgloss.Style
	StatusBarKey     lipgloss.Style
	StatusBarWarning lipgloss.Style
}

// NewStyles creates and initializes all styles
func NewStyles() *Styles {
	// Define colors
	mutedGray := lipgloss.Color("240")
	highlightBg := lipgloss.Color("237")
	selectedBg := lipgloss.Color("62")

	awsOrange := lipgloss.Color("214")
	googleBlue := lipgloss.Color("69")
	azureTeal := lipgloss.Color("45")
	defaultGray := lipgloss.Color("246")

	statusBg := lipgloss.Color("235")
	statusFg := lipgloss.Color("252")
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
			Bold(true),

		UnselectedRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		ModuleHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("141")).
			MarginTop(1).
			MarginBottom(0),

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
			Foreground(lipgloss.Color("212")).
			MarginBottom(1).
			Padding(0, 1),

		DetailKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("111")).
			Bold(true),

		DetailValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("141")).
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
