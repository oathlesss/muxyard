# Muxyard - Tmux Session Manager

A modern, interactive terminal UI for managing tmux sessions built with Go and Bubble Tea. Muxyard provides an intuitive interface for creating, listing, switching, and managing tmux sessions with support for session templates and Git repository integration.

## Features

- **Interactive TUI**: Full-screen terminal interface built with Bubble Tea
- **Session Management**: List, create, rename, kill, and attach to tmux sessions
- **Git Repository Integration**: Automatically discover and create sessions from Git repositories
- **Session Templates**: Pre-defined window layouts and commands for quick session setup
- **Focused Window Support**: Specify which window should be active when attaching to sessions
- **Two Creation Modes**:
  - **Git Repository Mode**: Select from configured repo directories
  - **Manual Mode**: Enter custom session name and directory
- **Advanced Filtering**: Search sessions and repositories by name and path
- **Visual Mode**: Multi-select sessions for batch operations
- **Persistent Windows**: Windows remain open after commands exit (no more closing when quitting nvim!)
- **Customizable Colors**: Full UI color theming support
- **Smart Session Naming**: Automatic name generation with conflict resolution
- **Configuration**: YAML-based configuration for repositories, templates, and UI colors

## Installation

### Prerequisites

- Go 1.21 or later
- tmux installed and available in PATH

### Build from Source

```bash
git clone https://github.com/yourusername/muxyard.git
cd muxyard
go build -o bin/muxyard ./cmd/muxyard
```

### Install Binary

```bash
# Install to /usr/local/bin (requires sudo)
sudo cp bin/muxyard /usr/local/bin/

# Or install to ~/bin (make sure it's in your PATH)
cp bin/muxyard ~/bin/
```

## Usage

Simply run `muxyard` to launch the interactive interface:

```bash
muxyard
```

### Key Bindings

#### Session List View
- `Enter` or `l` - Attach to selected session
- `c` or `n` - Create new session
- `r` - Rename selected session
- `d` or `x` - Delete selected session (confirmation for attached sessions)
- `/` - Filter/search sessions
- `Ctrl+V` - Toggle visual mode for multi-select
- `q` or `Ctrl+C` - Quit

#### Visual Mode (Multi-select)
- `j/k` or `↑/↓` - Extend selection
- `d` or `x` - Delete all selected sessions
- `Esc` or `Ctrl+V` - Exit visual mode

#### Repository List View
- `Enter` or `l` - Select repository
- `/` - Filter/search repositories (searches both name and path)
- `j/k` or `↑/↓` - Navigate
- `Esc` or `h` - Go back

#### Navigation
- `↑/↓` or `j/k` - Navigate lists
- `Enter` or `l` - Select item
- `Esc` or `h` - Go back
- `/` - Filter/search (in lists)

## Configuration

Muxyard creates a configuration file at `~/.config/muxyard/config.yaml` on first run.

### Example Configuration

```yaml
repo_directories:
  - ~/src
  - ~/code
  - ~/projects
  - ~/work

templates:
  - name: basic
    description: Single window with shell
    windows:
      - name: main
        command: ""

  - name: coding
    description: Editor, server, and shell windows
    focused_window: editor  # This window will be active when attaching
    windows:
      - name: editor
        command: nvim .
      - name: server
        command: ""
      - name: shell
        command: ""

  - name: monitor
    description: App, logs, and monitoring
    windows:
      - name: app
        command: ""
      - name: logs
        command: tail -f *.log
      - name: monitor
        command: htop

# UI Color Configuration
colors:
  title:
    foreground: "#FAFAFA"
    background: "#7D56F4"
  selected: "#EE6FF8"
  dimmed: "#626262"
  help: "#626262"
  error: "#FF0000"
  success: "#00FF00"
  border: "#874BFD"
  input: "#874BFD"
  focused_input: "#FF75B7"
  spinner: "205"
  highlight: "#FF75B7"
  filter_border: "#FF75B7"
```

### Configuration Options

- **repo_directories**: List of directories to scan for Git repositories
- **templates**: Session templates defining window layouts and commands
- **colors**: UI color theme configuration (optional)

### Session Templates

Templates define the structure of new sessions:

- **name**: Template identifier
- **description**: Human-readable description
- **focused_window**: Window name to focus when attaching (optional)
- **windows**: Array of window configurations
  - **name**: Window name (optional)
  - **command**: Command to run in window (optional, defaults to shell)

### Color Configuration

Customize the UI appearance with color themes:

- **title**: Title bar colors (foreground/background pair)
- **selected**: Selected items color
- **dimmed**: Inactive text color
- **help**: Help text color
- **error/success**: Message colors
- **border/input/focused_input**: Border colors
- **spinner/highlight/filter_border**: Accent colors

Colors can be specified as hex codes (`#FF0000`), color names (`red`), or ANSI codes (`205`).

## How It Works

### Session Creation Modes

1. **Git Repository Mode**:
   - Scans configured directories for Git repositories
   - Presents filterable list of found repositories (searches both name and path)
   - Auto-generates session names from repository names
   - Sets working directory to repository root

2. **Manual Mode**:
   - Prompts for custom session name
   - Prompts for custom directory path
   - Validates session name uniqueness and directory existence

### Session Management

- Lists all active tmux sessions with status (attached/detached)
- Shows number of windows per session
- Fast filtering/search by session name
- Visual mode for multi-select operations
- Supports rename and kill operations with confirmations
- Smart attach/switch based on context (inside tmux vs outside)

### Template System

- Pre-defined window layouts with custom commands
- Multiple built-in templates (basic, coding, monitor)
- Focused window support - specify which window should be active
- Persistent windows - commands run in shells that stay open after command exits
- Extensible through configuration file
- Commands run in session's working directory

## Architecture

The project follows clean architecture principles:

```
cmd/muxyard/          # Application entry point
internal/
  ├── config/         # Configuration management
  ├── tmux/           # Tmux command wrapper
  ├── git/            # Git repository discovery
  └── ui/             # Bubble Tea UI components
```

### Key Components

- **config**: YAML-based configuration loading and management
- **tmux**: Command wrapper for tmux operations (create, list, attach, etc.)
- **git**: Repository discovery and validation
- **ui**: Bubble Tea models and views with Lipgloss styling

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - Common UI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration parsing

## Compatibility

- **OS**: Linux, macOS (Windows not supported due to tmux requirement)
- **tmux**: Compatible with tmux 2.0+
- **Go**: Requires Go 1.21+

## Troubleshooting

### Common Issues

1. **"tmux not found"**: Ensure tmux is installed and in your PATH
2. **"No repositories found"**: Check your repo_directories configuration
3. **"Failed to create session"**: Verify session name doesn't already exist

### Debug Mode

Run with verbose output:
```bash
# Check tmux sessions
tmux list-sessions

# Test repository discovery
ls -la ~/.config/muxyard/
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Acknowledgments

- Inspired by [ThePrimeagen's tmux-sessionizer](https://github.com/ThePrimeagen/tmux-sessionizer)
- Built with the excellent [Charm](https://charm.sh) TUI libraries
