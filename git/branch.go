package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)
import "github.com/go-git/go-git/v5/plumbing"
import "strings"

func PacakgeUpdateBranchExists(pkgbase string, pkgver string) (bool, error) {
	pkgver = strings.Replace(pkgver, ":", "--", 1)
	remotePath := fmt.Sprintf("origin/packages/%s/%s", pkgbase, pkgver)

	repo, err := git.PlainOpen(".")

	if err != nil {
		return false, err
	}

	_, err = repo.ResolveRevision(plumbing.Revision(remotePath))

	return err == nil, nil
}

func CreateAndSwitchToPackageUpdateBranch(pkgbase string, pkgver string) error {
	pkgver = strings.Replace(pkgver, ":", "--", 1)
	branchRef := plumbing.NewBranchReferenceName(fmt.Sprintf("packages/%s/%s", pkgbase, pkgver))

	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	headRef, err := repo.Head()

	if err != nil {
		return err
	}

	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
	}); err != nil {
		return err
	}

	return nil
}

func SwitchToMaster() error {
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("master"),
	}); err != nil {
		return err
	}

	return nil
}

func PushPackageBranch(pkgbase string, pkgver string) error {
	pkgver = strings.Replace(pkgver, ":", "--", 1)
	branchRef := fmt.Sprintf("packages/%s/%s", pkgbase, pkgver)

	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	if err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s", branchRef, branchRef))},
	}); err != nil {
		return err
	}

	return nil
}