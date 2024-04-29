package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pacman"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func (pconfig *PackageConfig) CleanPkgrelBumpVersions(pkgver string) error {
	if pkgver == "" {
		return nil
	}

	if pconfig.Overrides == nil {
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

func GetMergedPkgnames(pkgbase string) ([]string, error) {
	basePath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildPkgnames(pkgbuildPath)
}

func GetMergedSources(pkgbase string) (map[string][]string, error) {
	basePath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildSources(pkgbuildPath)
}

func GetMergedVcsPkgver(pkgbase string) (string, int, int, error) {
	basePath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildVcsPkgver(pkgbuildPath)
}

func GetMergedVersion(pkgbase string) (string, error) {
	epoch, pkgver, pkgrel, subpkgrel, err := GetMergedVersionParts(pkgbase)

	if err != nil {
		return "", err
	}

	return GetVersionString(epoch, pkgver, pkgrel, subpkgrel), nil
}

func GetMergedVersionParts(pkgbase string) (string, string, int, int, error) {
	basePath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildVersionParts(pkgbuildPath)
}

func GetUpstreamPkgnames(pkgbase string) ([]string, error) {
	basePath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildPkgnames(pkgbuildPath)
}

func GetUpstreamVersion(pkgbase string) (string, error) {
	epoch, pkgver, pkgrel, subpkgrel, err := GetUpstreamVersionParts(pkgbase)

	if err != nil {
		return "", err
	}

	return GetVersionString(epoch, pkgver, pkgrel, subpkgrel), nil
}

func GetUpstreamVersionParts(pkgbase string) (string, string, int, int, error) {
	basePath := config.GetUpstreamPath(pkgbase)
	pkgbuildPath := path.Join(basePath, "PKGBUILD")

	return getPkgbuildVersionParts(pkgbuildPath)
}

func GetVersionString(epoch string, pkgver string, pkgrel int, subpkgrel int) string {
	pkgrelstr := fmt.Sprintf("%d", pkgrel)

	if subpkgrel > 0 {
		pkgrelstr = fmt.Sprintf("%s.%d", pkgrelstr, subpkgrel)
	}

	version := fmt.Sprintf("%s-%s", pkgver, pkgrelstr)

	if epoch != "" {
		version = fmt.Sprintf("%s:%s", epoch, version)
	}

	return version
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

func getPkgbuildSources(pkgbuildPath string) (map[string][]string, error) {
	cmdText := `
set -e

source "${1}"

mapfile -t SOURCE_ARRAYS < <(compgen -v source)

for SOURCE_ARRAY in "${SOURCE_ARRAYS[@]}"; do
	mapfile -t SOURCE_ITEMS < <(IFS=$'\n'; eval echo '"'"\${${SOURCE_ARRAY}[*]}"'"')

	for ITEM in "${SOURCE_ITEMS[@]}"; do
		echo "$SOURCE_ARRAY" "$ITEM"
	done
done
`
	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText, "bash", pkgbuildPath)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	lines := strings.Split(stdoutBuf.String(), "\n")
	result := map[string][]string{}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		lineParts := strings.SplitN(line, " ", 2)

		if len(lineParts) < 1 {
			return nil, errors.New(fmt.Sprintf("Invalid output: %s", line))
		}

		sourceType := lineParts[0]
		source := lineParts[1]

		if value, ok := result[sourceType]; ok {
			result[sourceType] = append(value, source)
		} else {
			result[sourceType] = []string{source}
		}
	}

	return result, nil
}

func getPkgbuildVcsPkgver(pkgbuildPath string) (string, int, int, error) {
	cmdText := `
set -e

source "${1}"

_r_pkgver="$([[ "$(type -t pkgver || true)" == "function" ]] && pkgver || echo "$pkgver")"
_r_pkgrel="1"

if [[ "$pkgver" == "$_r_pkgver" ]]; then
	_r_pkgrel="$pkgrel"
fi

echo "${_r_pkgver}"
echo "${_r_pkgrel}"
`

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText, "bash", path.Base(pkgbuildPath))
	cmd.Dir = path.Dir(pkgbuildPath)
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		return "", 0, 0, err
	}

	parts := misc.FilterEmptyString(strings.Split(stdoutBuf.String(), "\n"))

	if len(parts) != 2 {
		return "", 0, 0, errors.New(fmt.Sprintf("invalid pkgver result: %s", stdoutBuf.String()))
	}

	pkgrel, subpkgrel, err := getPkgrelParts(parts[1])

	if err != nil {
		return "", 0, 0, err
	}

	return parts[0], pkgrel, subpkgrel, nil
}

func getPkgbuildVersionParts(pkgbuildPath string) (string, string, int, int, error) {
	cmdText := `
set -e

source "${1}"

echo "${epoch}"
echo "${pkgver}"
echo "${pkgrel}"
`

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText, "bash", pkgbuildPath)
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		return "", "", 0, 0, err
	}

	parts := strings.Split(stdoutBuf.String(), "\n")
	pkgrel, subpkgrel, err := getPkgrelParts(parts[2])

	if err != nil {
		return "", "", 0, 0, err
	}

	return parts[0], parts[1], pkgrel, subpkgrel, nil
}

func getPkgrelParts(pkgrel string) (int, int, error) {
	intpkgrel := 0
	intsubpkgrel := 0

	if strings.Contains(pkgrel, ".") {
		pkgrelParts := strings.Split(pkgrel, ".")

		if len(pkgrelParts) != 2 {
			return 0, 0, errors.New("invalid pkgrel")
		}

		if val, err := strconv.Atoi(pkgrelParts[0]); err != nil {
			return 0, 0, err
		} else {
			intpkgrel = val
		}

		if val, err := strconv.Atoi(pkgrelParts[1]); err != nil {
			return 0, 0, err
		} else {
			intsubpkgrel = val
		}
	} else {
		if val, err := strconv.Atoi(pkgrel); err != nil {
			return 0, 0, err
		} else {
			intpkgrel = val
		}
	}

	return intpkgrel, intsubpkgrel, nil
}
