package git

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Repository struct {
	Name string
	Path string
}

func (r Repository) String() string {
	return r.Name
}

func FindRepositories(directories []string) ([]Repository, error) {
	var repos []Repository
	seen := make(map[string]bool)

	for _, dir := range directories {
		if dir == "" {
			continue
		}

		expandedDir := expandHome(dir)
		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			continue
		}

		dirRepos, err := findReposInDirectory(expandedDir)
		if err != nil {
			continue
		}

		for _, repo := range dirRepos {
			if !seen[repo.Path] {
				repos = append(repos, repo)
				seen[repo.Path] = true
			}
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

func findReposInDirectory(dir string) ([]Repository, error) {
	var repos []Repository

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		if info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repoName := filepath.Base(repoPath)

			repos = append(repos, Repository{
				Name: repoName,
				Path: repoPath,
			})

			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}

		depth := strings.Count(strings.TrimPrefix(path, dir), string(filepath.Separator))
		if depth > 3 {
			return filepath.SkipDir
		}

		return nil
	})

	return repos, err
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
