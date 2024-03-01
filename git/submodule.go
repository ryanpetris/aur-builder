package git

import (
	"github.com/go-git/go-git/v5"
)

type Submodule struct {
	Name   string
	Url    string
	Path   string
	Branch string
	Hash   string
}

func GetSubmodules(path string) (map[string]*Submodule, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()

	if err != nil {
		return nil, err
	}

	result := map[string]*Submodule{}

	for _, submodule := range submodules {
		submoduleConfig := submodule.Config()

		status, err := submodule.Status()

		if err != nil {
			return nil, err
		}

		result[submoduleConfig.Name] = &Submodule{
			Name:   submoduleConfig.Name,
			Path:   submoduleConfig.Path,
			Url:    submoduleConfig.URL,
			Branch: submoduleConfig.Branch,
			Hash:   status.Expected.String(),
		}
	}

	return result, nil
}
