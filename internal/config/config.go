package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SessionTemplate struct {
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Windows       []WindowConfig `yaml:"windows"`
	FocusedWindow string         `yaml:"focused_window,omitempty"`
}

type WindowConfig struct {
	Name    string `yaml:"name,omitempty"`
	Command string `yaml:"command,omitempty"`
}

type ColorConfig struct {
	Title        ColorPair `yaml:"title"`
	Selected     string    `yaml:"selected"`
	Dimmed       string    `yaml:"dimmed"`
	Help         string    `yaml:"help"`
	Error        string    `yaml:"error"`
	Success      string    `yaml:"success"`
	Border       string    `yaml:"border"`
	Input        string    `yaml:"input"`
	FocusedInput string    `yaml:"focused_input"`
	Spinner      string    `yaml:"spinner"`
	Highlight    string    `yaml:"highlight"`
	FilterBorder string    `yaml:"filter_border"`
}

type ColorPair struct {
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
}

type Config struct {
	RepoDirectories []string          `yaml:"repo_directories"`
	Templates       []SessionTemplate `yaml:"templates"`
	Colors          ColorConfig       `yaml:"colors,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		RepoDirectories: []string{
			filepath.Join(os.Getenv("HOME"), "src"),
			filepath.Join(os.Getenv("HOME"), "code"),
			filepath.Join(os.Getenv("HOME"), "projects"),
		},
		Colors: ColorConfig{
			Title: ColorPair{
				Foreground: "#FAFAFA",
				Background: "#7D56F4",
			},
			Selected:     "#EE6FF8",
			Dimmed:       "#626262",
			Help:         "#626262",
			Error:        "#FF0000",
			Success:      "#00FF00",
			Border:       "#874BFD",
			Input:        "#874BFD",
			FocusedInput: "#FF75B7",
			Spinner:      "205",
			Highlight:    "#FF75B7",
			FilterBorder: "#FF75B7",
		},
		Templates: []SessionTemplate{
			{
				Name:        "basic",
				Description: "Single window with shell",
				Windows: []WindowConfig{
					{Name: "main", Command: ""},
				},
			},
			{
				Name:          "coding",
				Description:   "Editor, server, and shell windows",
				FocusedWindow: "editor",
				Windows: []WindowConfig{
					{Name: "editor", Command: "nvim ."},
					{Name: "server", Command: ""},
					{Name: "shell", Command: ""},
				},
			},
			{
				Name:        "monitor",
				Description: "App, logs, and monitoring",
				Windows: []WindowConfig{
					{Name: "app", Command: ""},
					{Name: "logs", Command: "tail -f *.log"},
					{Name: "monitor", Command: "htop"},
				},
			},
		},
	}
}

func (c *Config) GetTemplate(name string) (*SessionTemplate, error) {
	for _, template := range c.Templates {
		if template.Name == name {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

func configDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "muxyard"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
