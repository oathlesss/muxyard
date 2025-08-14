package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepository(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	isRepo := IsGitRepository(projectRoot)
	t.Logf("Current project is git repo: %v", isRepo)

	tmpDir := t.TempDir()
	isNotRepo := IsGitRepository(tmpDir)
	if isNotRepo {
		t.Error("Expected temp directory to not be a git repository")
	}
}

func TestFindRepositories(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	projectParent := filepath.Dir(filepath.Join(wd, "..", ".."))
	repos, err := FindRepositories([]string{projectParent})
	if err != nil {
		t.Fatalf("FindRepositories failed: %v", err)
	}

	t.Logf("Found %d repositories in %s", len(repos), projectParent)
	for _, repo := range repos {
		t.Logf("  %s: %s", repo.Name, repo.Path)
	}
}
