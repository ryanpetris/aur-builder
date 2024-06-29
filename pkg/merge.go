package pkg

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os"
	"os/exec"
	"path"
)

func (pconfig *PackageConfig) ClearMerge(pkgbase string) error {
	basePath := config.GetPackagePath(pkgbase)
	mergedPath := config.GetMergedPath(pkgbase)
	scriptOverridePath := config.GetScriptOverridePath(pkgbase)

	if _, err := os.Stat(basePath); err != nil {
		return err
	}

	if _, err := os.Stat(mergedPath); err == nil {
		if err = os.RemoveAll(mergedPath); err != nil {
			return err
		}
	}

	if _, err := os.Stat(scriptOverridePath); err == nil {
		if err = os.RemoveAll(scriptOverridePath); err != nil {
			return err
		}
	}

	return nil
}

func (pconfig *PackageConfig) Merge(pkgbase string, processVcs bool) error {
	slog.Debug(fmt.Sprintf("Merging %s", pkgbase))

	basePath := config.GetPackagePath(pkgbase)
	localPath := config.GetLocalPath(pkgbase)
	mergedPath := config.GetMergedPath(pkgbase)
	scriptsPath := config.GetScriptsPath(pkgbase)
	scriptOverridePath := config.GetScriptOverridePath(pkgbase)
	upstreamPath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	onprepareScriptPath := path.Join(scriptsPath, "onprepare.sh")
	onmergeScriptPath := path.Join(scriptsPath, "onmerge.sh")

	if _, err := os.Stat(basePath); err != nil {
		return err
	}

	if _, err := os.Stat(mergedPath); err == nil {
		if err = os.RemoveAll(mergedPath); err != nil {
			return err
		}
	}

	if err := os.Mkdir(mergedPath, 0777); err != nil {
		return err
	}

	if _, err := os.Stat(onprepareScriptPath); err == nil {
		cmd := exec.Command(onprepareScriptPath)
		cmd.Dir = mergedPath

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if _, err := os.Stat(upstreamPath); err == nil {
		cmd := exec.Command("cp", "-ra", fmt.Sprintf("%s/.", upstreamPath), mergedPath)

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if _, err := os.Stat(localPath); err == nil {
		cmd := exec.Command("cp", "-ra", fmt.Sprintf("%s/.", localPath), mergedPath)

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if _, err := os.Stat(scriptOverridePath); err == nil {
		cmd := exec.Command("cp", "-ra", fmt.Sprintf("%s/.", scriptOverridePath), mergedPath)

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if err := formatPkgbuild(pkgbuildPath); err != nil {
		return err
	}

	if err := pconfig.ProcessOverrides(pkgbase); err != nil {
		return err
	}

	if processVcs {
		if err := pconfig.ProcessVcsOverrides(pkgbase); err != nil {
			return err
		}
	}

	if _, err := os.Stat(onmergeScriptPath); err == nil {
		cmd := exec.Command(onmergeScriptPath)
		cmd.Dir = mergedPath

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if err := formatPkgbuild(pkgbuildPath); err != nil {
		return err
	}

	return nil
}
