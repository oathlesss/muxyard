package tmux

import (
	"testing"
)

func TestIsTmuxAvailable(t *testing.T) {
	available := IsTmuxAvailable()
	if !available {
		t.Skip("tmux not available on this system")
	}
	t.Log("tmux is available")
}

func TestIsInsideTmux(t *testing.T) {
	inside := IsInsideTmux()
	t.Logf("Inside tmux: %v", inside)
}

func TestGenerateSessionName(t *testing.T) {
	sessions := []Session{
		{Name: "test", Windows: 1, Attached: false},
		{Name: "test_2", Windows: 1, Attached: false},
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/myproject", "myproject"},
		{"/home/user/test", "test_3"},
		{"/tmp/newproject", "newproject"},
	}

	for _, tt := range tests {
		result := GenerateSessionName(tt.path, sessions)
		if result != tt.expected {
			t.Errorf("GenerateSessionName(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}
