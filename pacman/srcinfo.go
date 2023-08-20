package pacman

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

type SrcinfoPkg struct {
	Pkgbase   string
	Pkgname   string
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

	Source       []SrcInfoArchItem
	Depends      []SrcInfoArchItem
	CheckDepends []SrcInfoArchItem
	MakeDepends  []SrcInfoArchItem
	OptDepends   []SrcInfoArchItem
	Provides     []SrcInfoArchItem
	Conflicts    []SrcInfoArchItem
	Replaces     []SrcInfoArchItem
	CkSums       []SrcInfoArchItem
	Md5Sums      []SrcInfoArchItem
	Sha1Sums     []SrcInfoArchItem
	Sha224Sums   []SrcInfoArchItem
	Sha256Sums   []SrcInfoArchItem
	Sha384Sums   []SrcInfoArchItem
	Sha512Sums   []SrcInfoArchItem
	B2Sums       []SrcInfoArchItem
}

type SrcInfoArchItem struct {
	Arch  string
	Value string
}

func (srcinfo *SrcinfoPkg) GetAllBuildDepends(arch ...string) []string {
	var allDeps []SrcInfoArchItem
	var result []string

	allDeps = append(allDeps, srcinfo.Depends[:]...)
	allDeps = append(allDeps, srcinfo.MakeDepends[:]...)
	allDeps = append(allDeps, srcinfo.CheckDepends[:]...)

	for _, item := range allDeps {
		if len(arch) > 0 && item.Arch != "" && !slices.Contains(arch, item.Arch) {
			continue
		}

		if !slices.Contains(result, item.Value) {
			result = append(result, item.Value)
		}
	}

	return result
}

func (srcinfo *SrcinfoPkg) GetFullVersion() string {
	if srcinfo.Epoch > 0 {
		return fmt.Sprintf("%d:%s-%d", srcinfo.Epoch, srcinfo.Pkgver, srcinfo.Pkgrel)
	}

	return fmt.Sprintf("%s-%d", srcinfo.Pkgver, srcinfo.Pkgrel)
}

func (srcinfo *SrcinfoPkg) Load(srcinfolines []string) error {
	for _, line := range srcinfolines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " = ", 2)
		field, value, arch := parts[0], parts[1], ""

		if strings.Contains(field, "_") {
			parts = strings.SplitN(field, "_", 2)
			field, arch = parts[0], parts[1]
		}

		fmt.Sprintf("%s %s %s", field, value, arch)

		switch field {
		case "pkgbase":
			srcinfo.Pkgbase = value
		case "pkgname":
			srcinfo.Pkgname = value
		case "pkgver":
			srcinfo.Pkgver = value
		case "pkgrel":
			srcinfo.Pkgrel, _ = strconv.Atoi(value)
		case "epoch":
			srcinfo.Epoch, _ = strconv.Atoi(value)
		case "pkgdesc":
			srcinfo.Pkgdesc = value
		case "url":
			srcinfo.Url = value
		case "install":
			srcinfo.Install = value
		case "changelog":
			srcinfo.Changelog = value
		case "arch":
			srcinfo.Arch = append(srcinfo.Arch, value)
		case "groups":
			srcinfo.Groups = append(srcinfo.Groups, value)
		case "license":
			srcinfo.License = append(srcinfo.License, value)
		case "noextract":
			srcinfo.NoExtract = append(srcinfo.NoExtract, value)
		case "options":
			srcinfo.Options = append(srcinfo.Options, value)
		case "backup":
			srcinfo.Backup = append(srcinfo.Backup, value)
		case "validpgpkeys":
			srcinfo.ValidPgpKeys = append(srcinfo.ValidPgpKeys, value)
		case "source":
			srcinfo.Source = append(srcinfo.Source, SrcInfoArchItem{arch, value})
		case "depends":
			srcinfo.Depends = append(srcinfo.Depends, SrcInfoArchItem{arch, value})
		case "checkdepends":
			srcinfo.CheckDepends = append(srcinfo.CheckDepends, SrcInfoArchItem{arch, value})
		case "makedepends":
			srcinfo.MakeDepends = append(srcinfo.MakeDepends, SrcInfoArchItem{arch, value})
		case "optdepends":
			srcinfo.OptDepends = append(srcinfo.OptDepends, SrcInfoArchItem{arch, value})
		case "provides":
			srcinfo.Provides = append(srcinfo.Provides, SrcInfoArchItem{arch, value})
		case "conflicts":
			srcinfo.Conflicts = append(srcinfo.Conflicts, SrcInfoArchItem{arch, value})
		case "replaces":
			srcinfo.Replaces = append(srcinfo.Replaces, SrcInfoArchItem{arch, value})
		case "cksums":
			srcinfo.CkSums = append(srcinfo.CkSums, SrcInfoArchItem{arch, value})
		case "md5sums":
			srcinfo.Md5Sums = append(srcinfo.Md5Sums, SrcInfoArchItem{arch, value})
		case "sha1sums":
			srcinfo.Sha1Sums = append(srcinfo.Sha1Sums, SrcInfoArchItem{arch, value})
		case "sha224sums":
			srcinfo.Sha224Sums = append(srcinfo.Sha224Sums, SrcInfoArchItem{arch, value})
		case "sha256sums":
			srcinfo.Sha256Sums = append(srcinfo.Sha256Sums, SrcInfoArchItem{arch, value})
		case "sha384sums":
			srcinfo.Sha384Sums = append(srcinfo.Sha384Sums, SrcInfoArchItem{arch, value})
		case "sha512sums":
			srcinfo.Sha512Sums = append(srcinfo.Sha512Sums, SrcInfoArchItem{arch, value})
		case "b2sums":
			srcinfo.B2Sums = append(srcinfo.B2Sums, SrcInfoArchItem{arch, value})
		}
	}

	return nil
}

func LoadSrcinfo(pkgbase string) ([]SrcinfoPkg, error) {
	mergedPath := config.GetMergedPath(pkgbase)
	srcinfoPath := path.Join(mergedPath, ".SRCINFO")
	dataBytes, err := os.ReadFile(srcinfoPath)

	if err != nil {
		return nil, err
	}

	var packages []SrcinfoPkg
	var pkgbaseLines []string
	var pkgLines []string
	isPkgbase := false

	appendPkg := func() error {
		var newlines []string
		newlines = append(newlines, pkgbaseLines[:]...)
		newlines = append(newlines, pkgLines[:]...)

		newpkg := SrcinfoPkg{}
		err := newpkg.Load(newlines)

		if err != nil {
			return err
		}

		packages = append(packages, newpkg)
		pkgLines = nil

		return nil
	}

	for _, line := range strings.Split(string(dataBytes), "\n") {
		line = strings.Trim(line, " \t")

		if strings.HasPrefix(line, "pkgbase = ") {
			isPkgbase = true
			pkgbaseLines = nil
			pkgLines = nil
		}

		if strings.HasPrefix(line, "pkgname = ") {
			if !isPkgbase {
				err := appendPkg()

				if err != nil {
					return nil, err
				}
			}

			isPkgbase = false
		}

		if isPkgbase {
			pkgbaseLines = append(pkgbaseLines, line)
		} else {
			pkgLines = append(pkgLines, line)
		}
	}

	if len(pkgLines) > 0 {
		err := appendPkg()

		if err != nil {
			return nil, err
		}
	}

	return packages, nil
}
