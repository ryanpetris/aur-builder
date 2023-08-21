package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/ryanpetris/aur-builder/config"
	"os"
	"path"
)

func CloneUpstream(pkgbase string, url string, tag string) error {
	upstreamPath := config.GetUpstreamPath(pkgbase)
	gitPath := path.Join(upstreamPath, ".git")

	if _, err := os.Stat(upstreamPath); err != nil {
		if err = os.RemoveAll(upstreamPath); err != nil {
			return err
		}
	}

	cloneOptions := &git.CloneOptions{
		URL:   url,
		Depth: 1,
	}

	if tag != "" {
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(CleanTagName(tag))
		cloneOptions.SingleBranch = true
	}

	_, err := git.PlainClone(upstreamPath, false, cloneOptions)

	if err != nil {
		return err
	}

	if err = os.RemoveAll(gitPath); err != nil {
		return err
	}

	return nil
}
