package pacman

import (
	"bytes"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type PkgInfo struct {
	Pkgbase   string
	Pkgname   []string
	Pkgver    string
	Pkgrel    int
	Epoch     int
	Pkgdesc   string
	Url       string
	Install   string
	Changelog string

	Arch         []string
	Groups       []string
	License      []string
	NoExtract    []string
	Options      []string
	Backup       []string
	ValidPgpKeys []string

	Source       []PkgInfoArchItem
	Depends      []PkgInfoArchItem
	CheckDepends []PkgInfoArchItem
	MakeDepends  []PkgInfoArchItem
	OptDepends   []PkgInfoArchItem
	Provides     []PkgInfoArchItem
	Conflicts    []PkgInfoArchItem
	Replaces     []PkgInfoArchItem
	CkSums       []PkgInfoArchItem
	Md5Sums      []PkgInfoArchItem
	Sha1Sums     []PkgInfoArchItem
	Sha224Sums   []PkgInfoArchItem
	Sha256Sums   []PkgInfoArchItem
	Sha384Sums   []PkgInfoArchItem
	Sha512Sums   []PkgInfoArchItem
	B2Sums       []PkgInfoArchItem
}

type PkgInfoArchItem struct {
	Arch  string
	Value string
}

func (pkginfo *PkgInfo) GetAllBuildDepends(arch ...string) []string {
	var allDeps []PkgInfoArchItem
	var result []string

	allDeps = append(allDeps, pkginfo.Depends[:]...)
	allDeps = append(allDeps, pkginfo.MakeDepends[:]...)
	allDeps = append(allDeps, pkginfo.CheckDepends[:]...)

	for _, item := range allDeps {
		if len(arch) > 0 && item.Arch != "" && !slices.Contains(arch, item.Arch) {
			continue
		}

		parts := regexp.MustCompile("[<>=]+").Split(item.Value, 2)

		if !slices.Contains(result, parts[0]) {
			result = append(result, parts[0])
		}
	}

	return result
}

func (pkginfo *PkgInfo) GetFullVersion() string {
	if pkginfo.Epoch > 0 {
		return fmt.Sprintf("%d:%s-%d", pkginfo.Epoch, pkginfo.Pkgver, pkginfo.Pkgrel)
	}

	return fmt.Sprintf("%s-%d", pkginfo.Pkgver, pkginfo.Pkgrel)
}

func (pkginfo *PkgInfo) Load(lines []string) error {
	var field, arch string

	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "%") && strings.HasSuffix(line, "%") {
			field = line[1 : len(line)-1]
			arch = ""

			if strings.Contains(field, "_") {
				parts := strings.SplitN(field, "_", 2)
				field, arch = parts[0], parts[1]
			}

			continue
		}

		if field == "" {
			continue
		}

		switch field {
		case "pkgbase":
			pkginfo.Pkgbase = line
		case "pkgname":
			pkginfo.Pkgname = append(pkginfo.Pkgname, line)
		case "pkgver":
			pkginfo.Pkgver = line
		case "pkgrel":
			pkginfo.Pkgrel, _ = strconv.Atoi(line)
		case "epoch":
			pkginfo.Epoch, _ = strconv.Atoi(line)
		case "pkgdesc":
			pkginfo.Pkgdesc = line
		case "url":
			pkginfo.Url = line
		case "install":
			pkginfo.Install = line
		case "changelog":
			pkginfo.Changelog = line
		case "arch":
			pkginfo.Arch = append(pkginfo.Arch, line)
		case "groups":
			pkginfo.Groups = append(pkginfo.Groups, line)
		case "license":
			pkginfo.License = append(pkginfo.License, line)
		case "noextract":
			pkginfo.NoExtract = append(pkginfo.NoExtract, line)
		case "options":
			pkginfo.Options = append(pkginfo.Options, line)
		case "backup":
			pkginfo.Backup = append(pkginfo.Backup, line)
		case "validpgpkeys":
			pkginfo.ValidPgpKeys = append(pkginfo.ValidPgpKeys, line)
		case "source":
			pkginfo.Source = append(pkginfo.Source, PkgInfoArchItem{arch, line})
		case "depends":
			pkginfo.Depends = append(pkginfo.Depends, PkgInfoArchItem{arch, line})
		case "checkdepends":
			pkginfo.CheckDepends = append(pkginfo.CheckDepends, PkgInfoArchItem{arch, line})
		case "makedepends":
			pkginfo.MakeDepends = append(pkginfo.MakeDepends, PkgInfoArchItem{arch, line})
		case "optdepends":
			pkginfo.OptDepends = append(pkginfo.OptDepends, PkgInfoArchItem{arch, line})
		case "provides":
			pkginfo.Provides = append(pkginfo.Provides, PkgInfoArchItem{arch, line})
		case "conflicts":
			pkginfo.Conflicts = append(pkginfo.Conflicts, PkgInfoArchItem{arch, line})
		case "replaces":
			pkginfo.Replaces = append(pkginfo.Replaces, PkgInfoArchItem{arch, line})
		case "cksums":
			pkginfo.CkSums = append(pkginfo.CkSums, PkgInfoArchItem{arch, line})
		case "md5sums":
			pkginfo.Md5Sums = append(pkginfo.Md5Sums, PkgInfoArchItem{arch, line})
		case "sha1sums":
			pkginfo.Sha1Sums = append(pkginfo.Sha1Sums, PkgInfoArchItem{arch, line})
		case "sha224sums":
			pkginfo.Sha224Sums = append(pkginfo.Sha224Sums, PkgInfoArchItem{arch, line})
		case "sha256sums":
			pkginfo.Sha256Sums = append(pkginfo.Sha256Sums, PkgInfoArchItem{arch, line})
		case "sha384sums":
			pkginfo.Sha384Sums = append(pkginfo.Sha384Sums, PkgInfoArchItem{arch, line})
		case "sha512sums":
			pkginfo.Sha512Sums = append(pkginfo.Sha512Sums, PkgInfoArchItem{arch, line})
		case "b2sums":
			pkginfo.B2Sums = append(pkginfo.B2Sums, PkgInfoArchItem{arch, line})
		}
	}

	return nil
}

func LoadPkgInfo(pkgbase string) (*PkgInfo, error) {
	cmdText := `
OLDENVARS=()
NEWENVARS=()
PKGENVARS=()

mapfile -t OLDENVARS < <(compgen -v)
source "${PKGBUILD_PATH}"
mapfile -t NEWENVARS < <(compgen -v)
mapfile -t PKGENVARS < <(comm -13 <(IFS=$'\n'; echo "${OLDENVARS[*]}" | sort) <(IFS=$'\n'; echo "${NEWENVARS[*]}" | sort))

for PKGENVAR in "${PKGENVARS[@]}"; do
	echo "%${PKGENVAR}%"
	(IFS=$'\n'; eval echo '"'"\${${PKGENVAR}[*]}"'"')
	echo ""
done
`

	mergedPath := config.GetMergedPath(pkgbase)

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", cmdText)
	cmd.Stdout = &stdoutBuf
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PKGBUILD_PATH=%s/PKGBUILD", mergedPath))

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	pkginfo := &PkgInfo{}

	if err := pkginfo.Load(strings.Split(stdoutBuf.String(), "\n")); err != nil {
		return nil, err
	}

	return pkginfo, nil
}
