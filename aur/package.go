package aur

import (
	"strconv"
	"strings"
)

type Package struct {
	ID            int64  `yaml:"ID,omitempty"`
	Name          string `yaml:"Name,omitempty"`
	Version       string `yaml:"Version,omitempty"`
	PackageBase   string `yaml:"PackageBase,omitempty"`
	PackageBaseID int64  `yaml:"PackageBaseID,omitempty"`

	Description string   `yaml:"Description,omitempty"`
	URL         string   `yaml:"URL,omitempty"`
	URLPath     string   `yaml:"URLPath,omitempty"`
	Keywords    []string `yaml:"Keywords,omitempty"`
	License     []string `yaml:"License,omitempty"`

	Depends      []string `yaml:"Depends,omitempty"`
	CheckDepends []string `yaml:"CheckDepends,omitempty"`
	MakeDepends  []string `yaml:"MakeDepends,omitempty"`
	OptDepends   []string `yaml:"OptDepends,omitempty"`

	Provides  []string `yaml:"Provides,omitempty"`
	Conflicts []string `yaml:"Conflicts,omitempty"`
	Replaces  []string `yaml:"Replaces,omitempty"`

	Submitter     string   `yaml:"Submitter,omitempty"`
	Maintainer    string   `yaml:"Maintainer,omitempty"`
	CoMaintainers []string `yaml:"CoMaintainers,omitempty"`

	FirstSubmitted int64 `yaml:"FirstSubmitted,omitempty"`
	LastModified   int64 `yaml:"LastModified,omitempty"`

	NumVotes   int64   `yaml:"NumVotes,omitempty"`
	Popularity float32 `yaml:"Popularity,omitempty"`

	OutOfDate bool `yaml:"OutOfDate,omitempty"`
}

type PackageSearchResults struct {
	ResultCount int       `yaml:"resultcount,omitempty"`
	Results     []Package `yaml:"results,omitempty"`
	Type        string    `yaml:"type,omitempty"`
	Version     int       `yaml:"version,omitempty"`
}

func (pkg *Package) GetEpoch() int {
	parts := strings.SplitN(pkg.Version, ":", 2)

	if len(parts) == 2 {
		result, _ := strconv.Atoi(parts[0])

		return result
	}

	return 0
}

func (pkg *Package) GetPkgrel() int {
	version := pkg.Version
	parts := strings.SplitN(version, ":", 2)

	if len(parts) == 2 {
		version = parts[1]
	}

	parts = strings.SplitN(version, "-", 2)
	result, _ := strconv.Atoi(parts[1])

	return result
}

func (pkg *Package) GetPkgver() string {
	version := pkg.Version
	parts := strings.SplitN(version, ":", 2)

	if len(parts) == 2 {
		version = parts[1]
	}

	parts = strings.SplitN(version, "-", 2)

	return parts[0]
}
