package pkg

import (
	"bytes"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/pacman"
	"os/exec"
	"path"
	"strings"
)

func (pconfig *PackageConfig) CleanPkgrelBumpVersions(pkgver string) error {
	if pkgver == "" {
		return nil
	}

	if pconfig.Overrides.BumpPkgrel == nil {
		return nil
	}

	for key, _ := range pconfig.Overrides.BumpPkgrel {
		if isNewer, _ := pacman.IsVersionNewer(key, pkgver); isNewer {
			delete(pconfig.Overrides.BumpPkgrel, key)
		}
	}

	if len(pconfig.Overrides.BumpPkgrel) == 0 {
		pconfig.Overrides.BumpPkgrel = nil
	}

	return nil
}

func GetMergedVersion(pkgbase string) (string, error) {
	basePath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildVersion(pkgbuildPath)
}

func GetUpstreamPkgnames(pkgbase string) ([]string, error) {
	basePath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildPkgnames(pkgbuildPath)
}

func GetUpstreamVersion(pkgbase string) (string, error) {
	basePath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildVersion(pkgbuildPath)
}

func getPkgbuildPkgnames(pkgbuildPath string) ([]string, error) {
	cmdText := `
set -e

source "${1}"

echo "${pkgname[@]}"
`

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText, "bash", pkgbuildPath)
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	line := strings.Split(stdoutBuf.String(), "\n")[0]

	return strings.Split(line, " "), nil
}

func getPkgbuildVersion(pkgbuildPath string) (string, error) {
	cmdText := `
set -e

source "${1}"

if [ -n "${epoch}" ]; then
  echo "${epoch}:${pkgver}-${pkgrel}"
else
  echo "${pkgver}-${pkgrel}"
fi
`

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText, "bash", pkgbuildPath)
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return strings.Split(stdoutBuf.String(), "\n")[0], nil
}
