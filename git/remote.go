package git

import (
	"github.com/go-git/go-git/v5"
)

func GetOriginUrl(path string) (string, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return "", err
	}

	remote, err := repo.Remote("origin")

	if err != nil {
		return "", err
	}

	return remote.Config().URLs[0], nil
}

func GetRevision(path string) (string, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return "", err
	}

	head, err := repo.Head()

	if err != nil {
		return "", err
	}

	return head.Hash().String(), nil
}
