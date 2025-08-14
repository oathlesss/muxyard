package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"muxyard/internal/config"
	"muxyard/internal/git"
	"muxyard/internal/tmux"
)

type viewState int

const (
	sessionListView viewState = iota
	createModeView
	repoListView
	manualCreateView
	manualDirectoryView
	templateSelectView
	renameSessionView
	loadingView
	confirmDeleteView
)

type listItem struct {
	title string
	desc  string
	data  any
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

type MainModel struct {
	cfg              *config.Config
	styles           Styles
	state            viewState
	list             list.Model
	spinner          spinner.Model
	nameInput        textinput.Model
	pathInput        textinput.Model
	sessions         []tmux.Session
	filteredSessions []tmux.Session
	repos            []git.Repository
	filteredRepos    []git.Repository
	templates        []config.SessionTemplate
	selectedRepo     *git.Repository
	selectedTemplate *config.SessionTemplate
	selectedSession  *tmux.Session
	selectedSessions map[int]bool
	error            string
	success          string
	sessionName      string
	sessionPath      string
	filterQuery      string
	repoFilterQuery  string
	quitting         bool
	width            int
	height           int
	inputFocused     bool
	visualMode       bool
	visualStart      int
	confirmDelete    bool
	deleteTarget     string
}

type sessionsLoadedMsg []tmux.Session
type reposLoadedMsg []git.Repository
type errorMsg string
type successMsg string

func NewMainModel(cfg *config.Config) MainModel {
	styles := NewStyles(cfg.Colors)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	nameInput := textinput.New()
	nameInput.Placeholder = "Session name"
	nameInput.CharLimit = 50

	pathInput := textinput.New()
	pathInput.Placeholder = "Directory path (e.g., ~/projects/myapp)"
	pathInput.CharLimit = 200

	// Custom list with disabled default filtering
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tmux Sessions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // We'll handle filtering ourselves
	l.KeyMap.Quit.SetKeys("q", "ctrl+c")
	// Remove conflicting keybindings
	l.KeyMap.CursorUp.SetKeys("up", "k")
	l.KeyMap.CursorDown.SetKeys("down", "j")

	return MainModel{
		cfg:              cfg,
		styles:           styles,
		state:            sessionListView,
		list:             l,
		spinner:          s,
		nameInput:        nameInput,
		pathInput:        pathInput,
		templates:        cfg.Templates,
		selectedSessions: make(map[int]bool),
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadSessions,
	)
}

func loadSessions() tea.Msg {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return errorMsg(fmt.Sprintf("Failed to load sessions: %v", err))
	}
	return sessionsLoadedMsg(sessions)
}

func loadRepositories(directories []string) tea.Cmd {
	return func() tea.Msg {
		repos, err := git.FindRepositories(directories)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to find repositories: %v", err))
		}
		return reposLoadedMsg(repos)
	}
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Clear messages after some interactions
	if m.error != "" || m.success != "" {
		switch msg.(type) {
		case tea.KeyMsg:
			m.error = ""
			m.success = ""
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}

		switch m.state {
		case sessionListView:
			return m.handleSessionListKeys(msg)
		case createModeView:
			return m.handleCreateModeKeys(msg)
		case repoListView:
			return m.handleRepoListKeys(msg)
		case manualCreateView:
			return m.handleManualCreateKeys(msg)
		case manualDirectoryView:
			return m.handleManualDirectoryKeys(msg)
		case templateSelectView:
			return m.handleTemplateSelectKeys(msg)
		case renameSessionView:
			return m.handleRenameSessionKeys(msg)
		case confirmDeleteView:
			return m.handleConfirmDeleteKeys(msg)
		}

	case sessionsLoadedMsg:
		m.sessions = []tmux.Session(msg)
		m.filteredSessions = m.sessions
		return m.updateSessionList(), nil

	case reposLoadedMsg:
		m.repos = []git.Repository(msg)
		m.filteredRepos = m.repos
		return m.updateRepoList(), nil

	case errorMsg:
		m.error = string(msg)
		return m, nil

	case successMsg:
		m.success = string(msg)
		return m, nil

	case spinner.TickMsg:
		if m.state == loadingView {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m MainModel) handleSessionListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle input when in filter mode first
	if m.inputFocused {
		switch msg.String() {
		case "esc":
			m.inputFocused = false
			m.nameInput.Blur()
			m.filterQuery = ""
			m.filteredSessions = m.sessions
			return m.updateSessionList(), nil
		case "enter":
			m.filterQuery = m.nameInput.Value()
			m.filteredSessions = m.fuzzyFilterSessions(m.filterQuery)
			m.inputFocused = false
			m.nameInput.Blur()
			return m.updateSessionList(), nil
		default:
			// Update input and apply real-time filtering
			m.nameInput, cmd = m.nameInput.Update(msg)
			m.filterQuery = m.nameInput.Value()
			m.filteredSessions = m.fuzzyFilterSessions(m.filterQuery)
			return m.updateSessionList(), cmd
		}
	}

	// Handle normal navigation and commands when not in filter mode
	switch msg.String() {
	case "q", "ctrl+c":
		if m.visualMode {
			// Exit visual mode
			m.visualMode = false
			m.selectedSessions = make(map[int]bool)
			return m.updateSessionList(), nil
		}
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.visualMode {
			// Exit visual mode
			m.visualMode = false
			m.selectedSessions = make(map[int]bool)
			return m.updateSessionList(), nil
		}

	case "/":
		if !m.visualMode {
			// Enter filter mode
			m.inputFocused = true
			m.nameInput.Focus()
			m.nameInput.SetValue(m.filterQuery)
			return m, nil
		}

	case "enter", "l":
		if !m.visualMode {
			// Attach to session
			if len(m.filteredSessions) > 0 {
				selectedIdx := m.list.Index()
				if selectedIdx >= 0 && selectedIdx < len(m.filteredSessions) {
					session := m.filteredSessions[selectedIdx]
					err := tmux.AttachToSession(session.Name)
					if err != nil {
						m.error = fmt.Sprintf("Failed to attach: %v", err)
					} else {
						m.quitting = true
						return m, tea.Quit
					}
				}
			}
		}

	case "c", "n":
		if !m.inputFocused && !m.visualMode {
			m.state = createModeView
			return m.updateCreateModeList(), nil
		}

	case "r":
		if !m.inputFocused && !m.visualMode && len(m.filteredSessions) > 0 {
			selectedIdx := m.list.Index()
			if selectedIdx >= 0 && selectedIdx < len(m.filteredSessions) {
				m.selectedSession = &m.filteredSessions[selectedIdx]
				m.state = renameSessionView
				m.nameInput.SetValue(m.selectedSession.Name)
				m.nameInput.Focus()
				m.inputFocused = false
				return m, nil
			}
		}

	case "ctrl+v":
		if !m.inputFocused {
			// Toggle visual mode
			m.visualMode = !m.visualMode
			if m.visualMode {
				m.visualStart = m.list.Index()
				m.selectedSessions = make(map[int]bool)
				if m.visualStart >= 0 && m.visualStart < len(m.filteredSessions) {
					m.selectedSessions[m.visualStart] = true
				}
			} else {
				m.selectedSessions = make(map[int]bool)
			}
			return m.updateSessionList(), nil
		}

	case "j", "down":
		if !m.inputFocused {
			if m.visualMode {
				// Move cursor and update selection
				newIdx := m.list.Index() + 1
				if newIdx < len(m.filteredSessions) {
					m.list.CursorDown()
					m.updateVisualSelection()
					return m.updateSessionList(), nil
				}
			} else {
				m.list.CursorDown()
			}
		}
		return m, nil

	case "k", "up":
		if !m.inputFocused {
			if m.visualMode {
				// Move cursor and update selection
				newIdx := m.list.Index() - 1
				if newIdx >= 0 {
					m.list.CursorUp()
					m.updateVisualSelection()
					return m.updateSessionList(), nil
				}
			} else {
				m.list.CursorUp()
			}
		}
		return m, nil

	case "d", "x":
		if !m.inputFocused {
			if m.visualMode {
				// Delete selected sessions
				return m.deleteSelectedSessions()
			} else if len(m.filteredSessions) > 0 {
				// Delete single session
				selectedIdx := m.list.Index()
				if selectedIdx >= 0 && selectedIdx < len(m.filteredSessions) {
					session := m.filteredSessions[selectedIdx]

					// Check if session is attached and confirm deletion
					if session.Attached {
						m.deleteTarget = session.Name
						m.state = confirmDeleteView
						return m, nil
					}

					// Delete non-attached session directly
					err := tmux.KillSession(session.Name)
					if err != nil {
						m.error = fmt.Sprintf("Failed to kill session: %v", err)
					} else {
						m.success = fmt.Sprintf("Killed session: %s", session.Name)
						return m, loadSessions
					}
				}
			}
		}
	}

	// Handle other list navigation when not in filter mode and not in visual mode
	if !m.inputFocused && !m.visualMode {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m MainModel) handleCreateModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "h":
		m.state = sessionListView
		return m.updateSessionList(), nil

	case "enter", "l":
		selectedIdx := m.list.Index()
		if selectedIdx == 0 {
			m.state = loadingView
			return m, loadRepositories(m.cfg.RepoDirectories)
		} else if selectedIdx == 1 {
			m.state = manualCreateView
			m.nameInput.SetValue("")
			m.nameInput.Placeholder = "Session name"
			m.nameInput.Focus()
			return m, nil
		}

	case "j", "down":
		m.list.CursorDown()

	case "k", "up":
		m.list.CursorUp()
	}

	return m, nil
}

func (m MainModel) handleRepoListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle input when in filter mode first
	if m.inputFocused {
		switch msg.String() {
		case "esc":
			m.inputFocused = false
			m.nameInput.Blur()
			m.repoFilterQuery = ""
			m.filteredRepos = m.repos
			return m.updateRepoList(), nil
		case "enter":
			m.repoFilterQuery = m.nameInput.Value()
			m.filteredRepos = m.fuzzyFilterRepos(m.repoFilterQuery)
			m.inputFocused = false
			m.nameInput.Blur()
			return m.updateRepoList(), nil
		default:
			// Update input and apply real-time filtering
			m.nameInput, cmd = m.nameInput.Update(msg)
			m.repoFilterQuery = m.nameInput.Value()
			m.filteredRepos = m.fuzzyFilterRepos(m.repoFilterQuery)
			return m.updateRepoList(), cmd
		}
	}

	// Handle normal navigation and commands when not in filter mode
	switch msg.String() {
	case "esc", "q":
		m.state = sessionListView
		return m.updateSessionList(), nil

	case "/":
		// Enter filter mode
		m.inputFocused = true
		m.nameInput.Focus()
		m.nameInput.SetValue(m.repoFilterQuery)
		return m, nil

	case "enter", "l":
		if len(m.filteredRepos) > 0 {
			selectedIdx := m.list.Index()
			if selectedIdx >= 0 && selectedIdx < len(m.filteredRepos) {
				m.selectedRepo = &m.filteredRepos[selectedIdx]
				m.state = templateSelectView
				return m.updateTemplateList(), nil
			}
		}

	case "j", "down":
		if !m.inputFocused {
			m.list.CursorDown()
		}
		return m, nil

	case "k", "up":
		if !m.inputFocused {
			m.list.CursorUp()
		}
		return m, nil
	}

	// Handle other list navigation when not in filter mode
	if !m.inputFocused {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m MainModel) handleManualCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = createModeView
		return m.updateCreateModeList(), nil

	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if name != "" {
			m.sessionName = name
			m.state = manualDirectoryView
			m.pathInput.SetValue(getDefaultPath())
			m.pathInput.Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m MainModel) handleManualDirectoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = manualCreateView
		m.nameInput.Focus()
		return m, nil

	case "enter":
		path := strings.TrimSpace(m.pathInput.Value())
		if path != "" {
			// Expand ~ to home directory
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}

			// Check if directory exists
			if info, err := os.Stat(path); err != nil || !info.IsDir() {
				m.error = fmt.Sprintf("Directory does not exist: %s", path)
				return m, nil
			}

			m.sessionPath = path
			m.state = templateSelectView
			return m.updateTemplateList(), nil
		}
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m MainModel) handleTemplateSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "h":
		if m.selectedRepo != nil {
			m.state = repoListView
			return m.updateRepoList(), nil
		} else {
			m.state = manualDirectoryView
			m.pathInput.Focus()
			return m, nil
		}

	case "enter", "l":
		if len(m.templates) > 0 {
			selectedIdx := m.list.Index()
			if selectedIdx >= 0 && selectedIdx < len(m.templates) {
				template := m.templates[selectedIdx]
				return m.createSession(&template)
			}
		}

	case "j", "down":
		m.list.CursorDown()

	case "k", "up":
		m.list.CursorUp()
	}

	return m, nil
}

func (m MainModel) handleRenameSessionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = sessionListView
		m.nameInput.Blur()
		return m.updateSessionList(), nil

	case "enter":
		newName := strings.TrimSpace(m.nameInput.Value())
		if newName != "" && newName != m.selectedSession.Name {
			err := tmux.RenameSession(m.selectedSession.Name, newName)
			if err != nil {
				m.error = fmt.Sprintf("Failed to rename session: %v", err)
			} else {
				m.success = fmt.Sprintf("Renamed session to: %s", newName)
				m.state = sessionListView
				m.nameInput.Blur()
				return m, loadSessions
			}
		} else {
			m.state = sessionListView
			m.nameInput.Blur()
			return m.updateSessionList(), nil
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m MainModel) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm deletion
		err := tmux.KillSession(m.deleteTarget)
		if err != nil {
			m.error = fmt.Sprintf("Failed to kill session: %v", err)
		} else {
			m.success = fmt.Sprintf("Killed session: %s", m.deleteTarget)
		}
		m.state = sessionListView
		m.deleteTarget = ""
		return m, loadSessions

	case "n", "N", "esc", "q":
		// Cancel deletion
		m.state = sessionListView
		m.deleteTarget = ""
		return m.updateSessionList(), nil
	}

	return m, nil
}

func (m MainModel) createSession(template *config.SessionTemplate) (tea.Model, tea.Cmd) {
	var sessionName, sessionPath string

	if m.selectedRepo != nil {
		sessionName = tmux.GenerateSessionName(m.selectedRepo.Path, m.sessions)
		sessionPath = m.selectedRepo.Path
	} else {
		sessionName = m.sessionName
		sessionPath = m.sessionPath
	}

	err := tmux.CreateSession(sessionName, sessionPath, template)
	if err != nil {
		m.error = fmt.Sprintf("Failed to create session: %v", err)
		return m, nil
	}

	err = tmux.AttachToSession(sessionName)
	if err != nil {
		m.error = fmt.Sprintf("Failed to attach to session: %v", err)
		return m, nil
	}

	m.quitting = true
	return m, tea.Quit
}

func (m MainModel) fuzzyFilterSessions(query string) []tmux.Session {
	if query == "" {
		return m.sessions
	}

	// Create a slice of session names for fuzzy matching
	sessionNames := make([]string, len(m.sessions))
	for i, session := range m.sessions {
		sessionNames[i] = session.Name
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, sessionNames)

	// Return sessions that match
	filtered := make([]tmux.Session, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, m.sessions[match.Index])
	}

	return filtered
}

func (m MainModel) fuzzyFilterRepos(query string) []git.Repository {
	if query == "" {
		return m.repos
	}

	// Create a slice combining repo names and paths for fuzzy matching
	repoSearchTerms := make([]string, len(m.repos))
	for i, repo := range m.repos {
		// Combine name and path for searching
		repoSearchTerms[i] = repo.Name + " " + repo.Path
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, repoSearchTerms)

	// Return repos that match
	filtered := make([]git.Repository, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, m.repos[match.Index])
	}

	return filtered
}

func (m MainModel) highlightMatches(text, query string) string {
	if query == "" {
		return text
	}

	// Simple highlighting - find matches and wrap with styling
	matches := fuzzy.Find(query, []string{text})
	if len(matches) == 0 {
		return text
	}

	match := matches[0]
	highlighted := text

	// Apply highlighting to matched characters
	for i := len(match.MatchedIndexes) - 1; i >= 0; i-- {
		idx := match.MatchedIndexes[i]
		if idx < len(text) {
			char := string(text[idx])
			highlighted = highlighted[:idx] +
				m.styles.Highlight.Render(char) +
				highlighted[idx+1:]
		}
	}

	return highlighted
}

func (m *MainModel) updateVisualSelection() {
	start := m.visualStart
	current := m.list.Index()

	// Clear previous selection
	m.selectedSessions = make(map[int]bool)

	// Select range - ensure we include both endpoints
	minIdx := start
	maxIdx := current
	if current < start {
		minIdx = current
		maxIdx = start
	}

	// Select all items in range
	for i := minIdx; i <= maxIdx; i++ {
		if i >= 0 && i < len(m.filteredSessions) {
			m.selectedSessions[i] = true
		}
	}
}

func (m MainModel) deleteSelectedSessions() (tea.Model, tea.Cmd) {
	var sessionsToDelete []string
	var attachedSessions []string

	// Collect sessions to delete
	for idx := range m.selectedSessions {
		if idx >= 0 && idx < len(m.filteredSessions) {
			session := m.filteredSessions[idx]
			if session.Attached {
				attachedSessions = append(attachedSessions, session.Name)
			} else {
				sessionsToDelete = append(sessionsToDelete, session.Name)
			}
		}
	}

	// If there are attached sessions, show confirmation
	if len(attachedSessions) > 0 {
		m.deleteTarget = strings.Join(attachedSessions, ", ")
		m.state = confirmDeleteView
		// Store non-attached sessions for deletion after confirmation
		return m, nil
	}

	// Delete non-attached sessions
	var errors []string
	for _, sessionName := range sessionsToDelete {
		err := tmux.KillSession(sessionName)
		if err != nil {
			errors = append(errors, sessionName)
		}
	}

	if len(errors) > 0 {
		m.error = fmt.Sprintf("Failed to kill sessions: %s", strings.Join(errors, ", "))
	} else {
		m.success = fmt.Sprintf("Killed %d sessions", len(sessionsToDelete))
	}

	// Exit visual mode
	m.visualMode = false
	m.selectedSessions = make(map[int]bool)

	return m, loadSessions
}

func (m MainModel) updateSessionList() MainModel {
	items := make([]list.Item, len(m.filteredSessions))
	for i, session := range m.filteredSessions {
		status := "detached"
		if session.Attached {
			status = "attached"
		}

		title := session.Name
		if m.filterQuery != "" {
			title = m.highlightMatches(session.Name, m.filterQuery)
		}

		// Add visual selection indicator
		if m.visualMode && m.selectedSessions[i] {
			title = "● " + title
		}

		desc := fmt.Sprintf("%d windows, %s", session.Windows, status)
		if m.visualMode && m.selectedSessions[i] {
			desc = "✓ " + desc
		}

		items[i] = listItem{
			title: title,
			desc:  desc,
			data:  session,
		}
	}
	m.list.SetItems(items)

	listTitle := "Tmux Sessions"
	if m.visualMode {
		selectedCount := len(m.selectedSessions)
		listTitle = fmt.Sprintf("Tmux Sessions (Visual: %d selected)", selectedCount)
	}
	m.list.Title = listTitle
	return m
}

func (m MainModel) updateCreateModeList() MainModel {
	items := []list.Item{
		listItem{title: "From Git Repository", desc: "Select from configured repo directories"},
		listItem{title: "Manual Setup", desc: "Enter custom name and directory"},
	}
	m.list.SetItems(items)
	m.list.Title = "Create New Session"
	return m
}

func (m MainModel) updateRepoList() MainModel {
	m.state = repoListView
	items := make([]list.Item, len(m.filteredRepos))
	for i, repo := range m.filteredRepos {
		title := repo.Name
		desc := repo.Path
		if m.repoFilterQuery != "" {
			// Try to match and highlight in both name and path
			title = m.highlightMatches(repo.Name, m.repoFilterQuery)
			desc = m.highlightMatches(repo.Path, m.repoFilterQuery)
		}

		items[i] = listItem{
			title: title,
			desc:  desc,
			data:  repo,
		}
	}
	m.list.SetItems(items)
	m.list.Title = "Select Repository"
	return m
}

func (m MainModel) updateTemplateList() MainModel {
	items := make([]list.Item, len(m.templates))
	for i, template := range m.templates {
		items[i] = listItem{
			title: template.Name,
			desc:  template.Description,
			data:  template,
		}
	}
	m.list.SetItems(items)
	m.list.Title = "Select Template"
	return m
}

func getDefaultPath() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}

func (m MainModel) View() string {
	if m.quitting {
		return ""
	}

	var content string

	switch m.state {
	case sessionListView:
		content = m.list.View()

		if m.inputFocused {
			content += "\n\n" + m.styles.FilterBorder.Render("Filter: "+m.nameInput.View())
		}

		if m.error != "" {
			content += "\n" + m.styles.Error.Render("Error: "+m.error)
		}
		if m.success != "" {
			content += "\n" + m.styles.Success.Render(m.success)
		}

		helpText := "\n'c' create • 'r' rename • 'd/x' delete • '/' filter • 'enter/l' attach • 'ctrl+v' visual • 'q' quit"
		if m.inputFocused {
			helpText = "\n'enter' apply filter • 'esc' cancel filter"
		} else if m.visualMode {
			helpText = "\n'j/k' select • 'd/x' delete selected • 'esc/ctrl+v' exit visual • 'q' quit"
		}
		content += m.styles.Help.Render(helpText)

	case createModeView:
		content = m.list.View()
		content += m.styles.Help.Render("\n'enter/l' select • 'j/k' navigate • 'h/esc' back")

	case repoListView:
		content = m.list.View()

		if m.inputFocused {
			content += "\n\n" + m.styles.FilterBorder.Render("Filter: "+m.nameInput.View())
		}

		if m.error != "" {
			content += "\n" + m.styles.Error.Render("Error: "+m.error)
		}
		if m.success != "" {
			content += "\n" + m.styles.Success.Render(m.success)
		}

		helpText := "\n'enter/l' select • 'j/k' navigate • '/' filter • 'h/esc' back"
		if m.inputFocused {
			helpText = "\n'enter' apply filter • 'esc' cancel filter"
		}
		content += m.styles.Help.Render(helpText)

	case manualCreateView:
		content = "Enter session name:\n\n"
		content += m.styles.Input.Render(m.nameInput.View())
		content += m.styles.Help.Render("\n'enter' continue • 'esc' back")

	case manualDirectoryView:
		content = fmt.Sprintf("Session: %s\n\nEnter directory path:\n\n", m.sessionName)
		content += m.styles.Input.Render(m.pathInput.View())
		content += m.styles.Help.Render("\n'enter' continue • 'esc' back")

	case templateSelectView:
		content = m.list.View()
		content += m.styles.Help.Render("\n'enter/l' create session • 'j/k' navigate • 'h/esc' back")

	case renameSessionView:
		content = fmt.Sprintf("Rename session: %s\n\n", m.selectedSession.Name)
		content += m.styles.Input.Render(m.nameInput.View())
		content += m.styles.Help.Render("\n'enter' rename • 'esc' cancel")

	case loadingView:
		content = fmt.Sprintf("\n%s Loading repositories...\n", m.spinner.View())

	case confirmDeleteView:
		content = fmt.Sprintf("Delete attached session: %s?\n\n", m.deleteTarget)
		content += "This session is currently attached and deleting it will close all windows.\n\n"
		content += m.styles.Help.Render("'y' yes • 'n/esc' no")
	}

	return m.styles.Title.Render("Muxyard - Tmux Session Manager") + "\n\n" + content
}
