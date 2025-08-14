package ui

import (
	"github.com/charmbracelet/lipgloss"
	"muxyard/internal/config"
)

type Styles struct {
	Title        lipgloss.Style
	Selected     lipgloss.Style
	Dimmed       lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	Border       lipgloss.Style
	Input        lipgloss.Style
	FocusedInput lipgloss.Style
	Spinner      lipgloss.Style
	Highlight    lipgloss.Style
	FilterBorder lipgloss.Style
}

func NewStyles(colors config.ColorConfig) Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.Title.Foreground)).
			Background(lipgloss.Color(colors.Title.Background)).
			Padding(0, 1),

		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.Selected)),

		Dimmed: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Dimmed)),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Help)).
			Margin(1, 0),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Border)).
			Padding(1, 2),

		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colors.Input)).
			Padding(0, 1),

		FocusedInput: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colors.FocusedInput)).
			Padding(0, 1),

		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Spinner)),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Highlight)).
			Bold(true),

		FilterBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colors.FilterBorder)).
			Padding(0, 1),
	}
}
