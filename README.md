# Terraviz

An interactive terminal UI that visualizes Terraform infrastructure dependencies using **Miller Columns** (like macOS Finder). Browse your infrastructure by following dependency relationships from root resources through their dependents.

## Features

### 📂 **Miller Columns Navigation**

Browse dependencies like browsing folders in Finder:

```
┌─ Root Resources ─────┐  ┌─ Dependents ──────────┐  ┌─ Dependencies ────────┐
│ aws_vpc.main     (2) │  │ aws_subnet.public    │  │ aws_instance.web     │
│ google_compute... (0)│  │ aws_security_grou... │  │                       │
│                       │  │                       │  │                       │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
```

- **Column 1**: Resources with no dependencies (roots like VPCs, networks)
- **Column 2**: Resources that depend on the selected root
- **Column 3**: Resources that depend on those, and so on...
- Numbers show child count: `(2)` means 2 resources depend on this

### 🏷️ **Provider Tabs**

Separate cloud providers into clean tabs:

```
  AWS    │  GOOGLE  │  AZURE
────────────────────────────────
```

- Most solutions use only 1 provider anyway
- No more mixing AWS, Google, and Azure resources
- Press `n`/`p` or `1-9` to switch tabs

### 🎯 **Multiple Input Modes**

- **Primary**: Parse `.tf` files (HCL parsing) - see what you're planning
- **Enhanced**: Automatically uses `terraform graph` for better dependency info
- **State Files**: Parse `.tfstate` files - see what's deployed

### 📊 **Detail Panel**

Right panel shows full resource details:
- Resource type, name, provider, module
- All attributes as key-value pairs
- Full list of dependencies

## Installation

```bash
go build -o terraviz
```

## Usage

### Basic Usage

```bash
# Visualize Terraform files (PRIMARY USE CASE)
# See dependencies before deploying!
./terraviz ./terraform

# Visualize deployed infrastructure from state
./terraviz terraform.tfstate

# Pull and visualize remote state
./terraviz --from-remote ./terraform
```

### Filtering

```bash
# Show only AWS resources
./terraviz --filter aws_ ./terraform

# Show resources from a specific module
./terraviz --module vpc ./terraform
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`↓` or `k`/`j` | Navigate up/down in current column |
| `←` or `h` | Move to previous column (go back up dependency tree) |
| `→` or `l` | Move to next column (expand children/dependents) |
| `Tab` | Switch focus to detail panel |
| `n` or `]` | Next provider tab |
| `p` or `[` | Previous provider tab |
| `1-9` | Jump to specific provider tab |
| `Q` or `Ctrl+C` | Quit |

## How It Works

### Dependency Flow

Terraviz shows **who depends on whom**:

1. **Start with roots**: Resources with no dependencies (VPCs, networks, resource groups)
2. **Select a resource**: See all resources that depend on it in the next column
3. **Keep browsing**: Navigate right to see dependents of dependents
4. **Go back**: Navigate left to move back up the tree

### Example Flow

```
VPC (selected)
  ↓ depends on
Subnet, Security Group
  ↓ depends on
EC2 Instance
```

In Terraviz:
1. See `aws_vpc.main` in first column
2. Press `→` to see `aws_subnet.public` and `aws_security_group.web`
3. Select subnet, press `→` to see `aws_instance.web`
4. Press `←` to go back to subnet/security group
5. Press `←` again to go back to VPC

## Project Structure

```
terraviz/
├── main.go           # CLI interface
├── go.mod
├── parser/
│   ├── hcl.go        # HCL parser (primary - parse .tf files)
│   ├── dot.go        # DOT parser (enhanced - terraform graph)
│   └── graph.go      # Parsing coordinator
├── ui/
│   ├── app.go        # Main TUI application
│   ├── columns.go    # Miller Columns view
│   ├── detail.go     # Resource detail view
│   └── styles.go     # UI styling with provider colors
└── model/
    └── resource.go   # Core data structures
```

## Parsing Modes

### Primary: HCL Parsing

Parse `.tf` files directly to see planned infrastructure:
- See what you're about to deploy
- Works without Terraform installed
- Reads resource blocks and `depends_on` declarations

### Enhanced: Terraform Graph

If `terraform` is available, automatically enhances with:
- More accurate dependency edges
- Computed dependencies
- Better module resolution
- Status bar shows "terraform graph (enhanced)"

### State Files

Parse `.tfstate` JSON to see deployed resources:
- Local state files
- Remote state via `--from-remote`
- Shows actual deployed configuration

## Example

```bash
./terraviz ./example
```

You'll see:
- Provider tabs at top (AWS, GOOGLE, AZURE)
- First column shows root resources (VPC, compute network, resource group)
- Navigate with arrow keys to explore dependencies
- Detail panel shows full resource info

Start with AWS tab showing `aws_vpc.main`, press `→` to see subnet and security group, press `→` again on subnet to see the EC2 instance!

## Requirements

- Go 1.21 or later
- Terraform (optional, for enhanced mode)

## Libraries

- [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [github.com/hashicorp/hcl/v2](https://github.com/hashicorp/hcl) - HCL parsing
- [github.com/zclconf/go-cty](https://github.com/zclconf/go-cty) - HCL values

## License

MIT
