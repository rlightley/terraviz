package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"terraviz/model"
	"terraviz/parser"
	"terraviz/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

func main() {
	// Define flags
	filterFlag := flag.String("filter", "", "Filter resources by type prefix")
	moduleFlag := flag.String("module", "", "Scope view to a specific module")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	fromRemoteFlag := flag.Bool("from-remote", false, "Pull state from remote using terraform state pull")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: terraviz [options] <path>\n\n")
		fmt.Fprintf(os.Stderr, "Terraviz - Interactive terminal UI for visualizing Terraform infrastructure\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path>    Path to directory containing .tf files or a .tfstate file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  terraviz ./terraform              # Visualize .tf files in directory\n")
		fmt.Fprintf(os.Stderr, "  terraviz terraform.tfstate        # Visualize state file\n")
		fmt.Fprintf(os.Stderr, "  terraviz --from-remote ./terraform # Pull and visualize remote state\n")
		fmt.Fprintf(os.Stderr, "  terraviz --filter aws_ ./terraform # Show only AWS resources\n")
	}

	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("Terraviz version %s\n", version)
		os.Exit(0)
	}

	// Get path argument
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: path argument is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	path := args[0]

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: path does not exist: %s\n", path)
		os.Exit(1)
	}

	// Parse and build graph
	parseResult := parser.BuildGraph(path, *fromRemoteFlag)

	// Apply filters if specified
	if *filterFlag != "" {
		parseResult.Graph.Resources = filterByType(parseResult.Graph.Resources, *filterFlag)
	}

	if *moduleFlag != "" {
		parseResult.Graph.Resources = filterByModule(parseResult.Graph.Resources, *moduleFlag)
	}

	// Check if we have any resources
	if len(parseResult.Graph.Resources) == 0 {
		fmt.Fprintf(os.Stderr, "No resources found in %s\n", path)
		if parseResult.ErrorMessage != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", parseResult.ErrorMessage)
		}
		os.Exit(1)
	}

	// Create and run the TUI
	model := ui.NewModel(parseResult)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// filterByType filters resources by type prefix
func filterByType(resources map[string]*model.Resource, prefix string) map[string]*model.Resource {
	filtered := make(map[string]*model.Resource)
	prefix = strings.ToLower(prefix)

	for id, res := range resources {
		if strings.HasPrefix(strings.ToLower(res.Type), prefix) {
			filtered[id] = res
		}
	}

	return filtered
}

// filterByModule filters resources by module name
func filterByModule(resources map[string]*model.Resource, moduleName string) map[string]*model.Resource {
	filtered := make(map[string]*model.Resource)

	for id, res := range resources {
		// Match exact module or root
		if moduleName == "" && res.Module == "" {
			filtered[id] = res
		} else if strings.Contains(res.Module, moduleName) {
			filtered[id] = res
		}
	}

	return filtered
}
