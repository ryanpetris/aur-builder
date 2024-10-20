package git

import (
	"github.com/go-git/go-git/v5"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
)

func GetLastCommitLog() (*gitobject.Commit, error) {
	repo, err := git.PlainOpen(".")

	if err != nil {
		return nil, err
	}

	logIter, err := repo.Log(&git.LogOptions{})

	if err != nil {
		return nil, err
	}

	logItem, err := logIter.Next()

	if err != nil {
		return nil, err
	}

	return logItem, nil
}
