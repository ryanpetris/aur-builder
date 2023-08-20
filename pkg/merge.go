package pkg

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os"
	"os/exec"
	"path"
)

func (pconfig *PackageConfig) Merge(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Merging %s", pkgbase))

	basePath := config.GetPackagePath(pkgbase)
	localPath := config.GetLocalPath(pkgbase)
	mergedPath := config.GetMergedPath(pkgbase)
	scriptsPath := config.GetScriptsPath(pkgbase)
	upstreamPath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
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

	if pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		return err
	} else {
		if _, err = pkgbuild.WriteString("\n"); err != nil {
			_ = pkgbuild.Close()

			return err
		}

		if err = pkgbuild.Close(); err != nil {
			return err
		}
	}

	if err := pconfig.ProcessOverrides(pkgbase); err != nil {
		return err
	}

	if _, err := os.Stat(onmergeScriptPath); err == nil {
		cmd := exec.Command(onmergeScriptPath)
		cmd.Dir = mergedPath

		if err = cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
