package impenv

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pkg"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
)

var (
	localRemoteVersionScript = "onlocalremoteversion.sh"
	localUpdateScript        = "onlocalupdate.sh"
)

type LocalImportEnv struct {
}

func (ienv LocalImportEnv) IsLocalEnv() bool {
	return true
}

func (ienv LocalImportEnv) GetPackageInfo(pkgname []string) ([]misc.PackageInfo, error) {
	var result []misc.PackageInfo

	packages, err := pkg.GetPackages()

	if err != nil {
		return nil, err
	}

	for _, pkgbase := range packages {
		pkgconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			continue
		}

		if pkgconfig.Source != "local" {
			continue
		}

		pkgnames, err := pkg.GetLocalPkgnames(pkgbase)

		if err != nil {
			continue
		}

		var version string

		for _, item := range pkgnames {
			if !slices.Contains(pkgname, item) {
				continue
			}

			if version == "" {
				version, err = ienv.getRemoteVersion(pkgbase)

				if err != nil {
					fmt.Printf("Failed to get remote version for package %s\n, skipping", pkgbase)
					break
				}
			}

			result = append(result, misc.PackageInfo{
				Pkgbase:     pkgbase,
				Pkgname:     item,
				FullVersion: version,
			})
		}
	}

	return result, nil
}

func (ienv LocalImportEnv) PackageExists(pkgbase string) (bool, error) {
	return false, errors.New("import not supported")
}

func (ienv LocalImportEnv) PackageImport(pkgbase string, version string) error {
	pkgPath := config.GetLocalPath(pkgbase)

	if _, err := os.Stat(pkgPath); err != nil {
		return errors.New("local only supports package updates")
	}

	scriptPath := path.Join(config.GetScriptsPath(pkgbase), localUpdateScript)

	if _, err := os.Stat(scriptPath); err != nil {
		return errors.New(fmt.Sprintf("local script %s does not exist for package %s", localUpdateScript, pkgbase))
	}

	var out bytes.Buffer
	var outErr bytes.Buffer

	cmd := exec.Command(scriptPath, pkgbase, version)
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	cmd.Dir = pkgPath

	return cmd.Run()
}

func (ienv LocalImportEnv) cleanVersion(pkgbase string, version string) (string, error) {
	epoch, _, _, _, err := pkg.GetLocalVersionParts(pkgbase)

	if err != nil {
		return "", err
	}

	return pkg.GetVersionString(epoch, version, 1, 0), nil
}

func (ienv LocalImportEnv) getRemoteVersion(pkgbase string) (string, error) {
	pkgPath := config.GetLocalPath(pkgbase)

	if _, err := os.Stat(pkgPath); err != nil {
		return "", errors.New(fmt.Sprintf("local dir does not exist for package %s", pkgbase))
	}

	scriptPath := path.Join(config.GetScriptsPath(pkgbase), localRemoteVersionScript)

	if _, err := os.Stat(scriptPath); err != nil {
		return "", errors.New(fmt.Sprintf("local script %s does not exist for package %s", localRemoteVersionScript, pkgbase))
	}

	var stdoutBuf bytes.Buffer

	cmd := exec.Command(scriptPath, pkgbase)
	cmd.Dir = pkgPath
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	version := strings.Replace(stdoutBuf.String(), "\n", "", -1)

	return ienv.cleanVersion(pkgbase, version)
}
