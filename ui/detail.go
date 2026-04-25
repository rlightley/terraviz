package ui

import (
	"fmt"
	"sort"
	"strings"
	"terraviz/model"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// DetailView displays detailed information about a selected resource
type DetailView struct {
	viewport viewport.Model
	styles   *Styles
	width    int
	height   int
}

// NewDetailView creates a new detail view
func NewDetailView(styles *Styles) DetailView {
	vp := viewport.New(0, 0)
	return DetailView{
		viewport: vp,
		styles:   styles,
	}
}

// SetSize updates the dimensions of the detail view
func (d *DetailView) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.viewport.Width = width - 4  // Account for padding and border
	d.viewport.Height = height - 2 // Account for border
}

// SetResource updates the detail view to show a specific resource
func (d *DetailView) SetResource(resource *model.Resource) {
	if resource == nil {
		d.viewport.SetContent(d.styles.DetailValue.Render("No resource selected"))
		return
	}

	var content strings.Builder

	// Header block
	header := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
	content.WriteString(d.styles.DetailHeader.Render(header))
	content.WriteString("\n\n")

	// Provider
	providerStyle := d.styles.GetProviderStyle(resource.Provider)
	content.WriteString(d.styles.DetailKey.Render("Provider: "))
	content.WriteString(providerStyle.Render(resource.Provider))
	content.WriteString("\n")

	// Module
	module := resource.Module
	if module == "" {
		module = "(root)"
	}
	content.WriteString(d.styles.DetailKey.Render("Module:   "))
	content.WriteString(d.styles.DetailValue.Render(module))
	content.WriteString("\n")

	// Type
	content.WriteString(d.styles.DetailKey.Render("Type:     "))
	content.WriteString(d.styles.DetailValue.Render(resource.Type))
	content.WriteString("\n")

	// Attributes section
	if len(resource.Attributes) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Attributes"))
		content.WriteString("\n\n")

		// Sort attributes for consistent display
		keys := make([]string, 0, len(resource.Attributes))
		for k := range resource.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := resource.Attributes[key]
			valueStr := fmt.Sprintf("%v", value)

			// Truncate long values
			if len(valueStr) > 60 {
				valueStr = valueStr[:57] + "..."
			}

			content.WriteString(d.styles.DetailKey.Render(fmt.Sprintf("  %-20s ", key)))
			content.WriteString(d.styles.DetailValue.Render(valueStr))
			content.WriteString("\n")
		}
	}

	// Dependencies section
	if len(resource.Dependencies) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Dependencies"))
		content.WriteString("\n\n")

		for _, dep := range resource.Dependencies {
			content.WriteString(d.styles.DetailValue.Render(fmt.Sprintf("  → %s", dep)))
			content.WriteString("\n")
		}
	}

	d.viewport.SetContent(content.String())
	d.viewport.GotoTop()
}

// Update handles messages for the detail view
func (d *DetailView) Update(msg tea.Msg) (DetailView, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return *d, cmd
}

// View renders the detail view
func (d *DetailView) View() string {
	return d.viewport.View()
}
