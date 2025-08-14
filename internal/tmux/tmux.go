package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"muxyard/internal/config"
)

type Session struct {
	Name     string
	Windows  int
	Attached bool
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func ListSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}:#{session_windows}:#{session_attached}")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return []Session{}, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	sessions := make([]Session, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			continue
		}

		session := Session{
			Name:     parts[0],
			Attached: parts[2] == "1",
		}

		if parts[1] != "" {
			// parts[1] is already the window count from #{session_windows}
			if windowCount, err := strconv.Atoi(parts[1]); err == nil {
				session.Windows = windowCount
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func SessionExists(name string) (bool, error) {
	sessions, err := ListSessions()
	if err != nil {
		return false, err
	}

	for _, session := range sessions {
		if session.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func CreateSession(name, path string, template *config.SessionTemplate) error {
	if len(template.Windows) == 0 {
		return fmt.Errorf("template must have at least one window")
	}

	firstWindow := template.Windows[0]
	args := []string{"new-session", "-d", "-s", name, "-c", path}

	if firstWindow.Name != "" {
		args = append(args, "-n", firstWindow.Name)
	}

	// Use shell wrapper for command to keep shell alive after command exits
	if firstWindow.Command != "" {
		shellCmd := fmt.Sprintf("%s; exec $SHELL", firstWindow.Command)
		args = append(args, "sh", "-c", shellCmd)
	}

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create additional windows
	for i, window := range template.Windows[1:] {
		windowArgs := []string{"new-window", "-t", name, "-c", path}

		if window.Name != "" {
			windowArgs = append(windowArgs, "-n", window.Name)
		}

		// Use shell wrapper for command to keep shell alive after command exits
		if window.Command != "" {
			shellCmd := fmt.Sprintf("%s; exec $SHELL", window.Command)
			windowArgs = append(windowArgs, "sh", "-c", shellCmd)
		}

		cmd := exec.Command("tmux", windowArgs...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create window %d: %w", i+2, err)
		}
	}

	// Focus the specified window if provided
	if template.FocusedWindow != "" {
		focusCmd := exec.Command("tmux", "select-window", "-t", name+":"+template.FocusedWindow)
		focusCmd.Run() // Don't fail if this doesn't work
	}

	return nil
}

func AttachToSession(name string) error {
	var cmd *exec.Cmd
	if IsInsideTmux() {
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", name)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RenameSession(oldName, newName string) error {
	cmd := exec.Command("tmux", "rename-session", "-t", oldName, newName)
	return cmd.Run()
}

func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

func GenerateSessionName(repoPath string, existingSessions []Session) string {
	baseName := filepath.Base(repoPath)
	name := baseName

	counter := 1
	for {
		exists := false
		for _, session := range existingSessions {
			if session.Name == name {
				exists = true
				break
			}
		}

		if !exists {
			break
		}

		counter++
		name = fmt.Sprintf("%s_%d", baseName, counter)
	}

	return name
}
