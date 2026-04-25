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

// DetailView displays detailed information about a selected resource
type DetailView struct {
	viewport viewport.Model
	styles   *Styles
	resource *model.Resource
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
	d.viewport.Width = maxInt(width-2, 1) // Account for border
	d.viewport.Height = height - 2        // Account for border
	if d.viewport.Height < 1 {
		d.viewport.Height = 1
	}

	if d.resource != nil {
		d.SetResource(d.resource)
	}
}

// SetResource updates the detail view to show a specific resource
func (d *DetailView) SetResource(resource *model.Resource) {
	d.resource = resource

	if resource == nil {
		d.viewport.SetContent(d.styles.DetailValue.Render("No resource selected"))
		return
	}

	var content strings.Builder
	contentWidth := d.contentWidth()

	// Header block
	header := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
	headerStyle := d.styles.DetailHeader.Width(contentWidth)
	headerTextWidth := maxInt(contentWidth-2, 1)
	content.WriteString(headerStyle.Render(truncateDetailText(header, headerTextWidth)))
	content.WriteString("\n\n")

	// Provider
	providerStyle := d.styles.GetProviderStyle(resource.Provider)
	content.WriteString(d.renderDetailPair("Provider", resource.Provider, providerStyle))
	content.WriteString("\n")

	// Module
	if resource.Module != "" {
		content.WriteString(d.renderDetailPair("Module", resource.Module, d.styles.DetailValue))
		content.WriteString("\n")
	}

	// Type
	content.WriteString(d.renderDetailPair("Type", resource.Type, d.styles.DetailValue))
	content.WriteString("\n")

	if len(resource.ModuleInputs) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Module Inputs"))
		content.WriteString("\n\n")

		inputNames := make([]string, 0, len(resource.ModuleInputs))
		for inputName := range resource.ModuleInputs {
			inputNames = append(inputNames, inputName)
		}
		sort.Strings(inputNames)

		for _, inputName := range inputNames {
			content.WriteString(d.renderDetailPair(inputName, strings.Join(resource.ModuleInputs[inputName], ", "), d.styles.DetailValue))
			content.WriteString("\n")
		}
	}

	if len(resource.VariableRefs) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Variables"))
		content.WriteString("\n\n")

		for _, variableRef := range resource.VariableRefs {
			content.WriteString(d.renderBullet(variableRef))
			content.WriteString("\n")
		}
	}

	if len(resource.LocalRefs) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Locals"))
		content.WriteString("\n\n")

		for _, localRef := range resource.LocalRefs {
			content.WriteString(d.renderBullet(localRef))
			content.WriteString("\n")
		}
	}

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
			content.WriteString(d.renderDetailPair(key, valueStr, d.styles.DetailValue))
			content.WriteString("\n")
		}
	}

	// Dependencies section
	if len(resource.Dependencies) > 0 {
		content.WriteString("\n")
		content.WriteString(d.styles.SectionTitle.Render("Dependencies"))
		content.WriteString("\n\n")

		for _, dep := range resource.Dependencies {
			content.WriteString(d.renderBullet(dep))
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

func (d *DetailView) contentWidth() int {
	if d.viewport.Width > 0 {
		return d.viewport.Width
	}

	return maxInt(d.width-2, 1)
}

func (d *DetailView) renderDetailPair(key, value string, valueStyle lipgloss.Style) string {
	contentWidth := d.contentWidth()
	keyWidth := contentWidth / 3
	if keyWidth < 10 {
		keyWidth = 10
	}
	if keyWidth > 18 {
		keyWidth = 18
	}

	keyLabel := fmt.Sprintf("  %-*s ", keyWidth, truncateDetailText(key, keyWidth))
	remainingWidth := contentWidth - lipgloss.Width(keyLabel)
	if remainingWidth < 1 {
		remainingWidth = 1
	}

	return d.styles.DetailKey.Render(keyLabel) + valueStyle.Render(truncateDetailText(value, remainingWidth))
}

func (d *DetailView) renderBullet(value string) string {
	prefix := "  • "
	remainingWidth := d.contentWidth() - lipgloss.Width(prefix)
	if remainingWidth < 1 {
		remainingWidth = 1
	}

	return d.styles.DetailValue.Render(prefix + truncateDetailText(value, remainingWidth))
}

func truncateDetailText(value string, maxWidth int) string {
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
