package pkg

import (
	"errors"
	"github.com/ryanpetris/aur-builder/config"
	"gopkg.in/yaml.v3"
	"os"
)

type PackageConfig struct {
	Source     string                 `yaml:"source,omitempty"`
	BuildFirst bool                   `yaml:"buildFirst,omitempty"`
	Overrides  PackageConfigOverrides `yaml:"overrides,omitempty"`
}

type PackageConfigOverrides struct {
	AppendPkgbuild       string                        `yaml:"appendPkgbuild,omitempty"`
	BumpPkgrel           map[string]int                `yaml:"bumpPkgrel,omitempty"`
	ClearDependsVersions bool                          `yaml:"clearDependsVersions,omitempty"`
	ClearPkgverFunc      bool                          `yaml:"clearPkgverFunc,omitempty"`
	ClearSignatures      bool                          `yaml:"clearSignatures,omitempty"`
	DeleteFile           []string                      `yaml:"deleteFile,omitempty"`
	ModifySection        []PackageConfigModifySection  `yaml:"modifySection,omitempty"`
	RemoveSource         []string                      `yaml:"removeSource,omitempty"`
	RenameFile           []PackageConfigOverrideFromTo `yaml:"renameFile,omitempty"`
	RenamePackage        []PackageConfigOverrideFromTo `yaml:"renamePackage,omitempty"`
	ReplacePkgbuild      []PackageConfigOverrideFromTo `yaml:"replacePkgbuild,omitempty"`
}

type PackageConfigOverrideFromTo struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`
}

type PackageConfigModifySection struct {
	Section  string                        `yaml:"section,omitempty"`
	Sections []string                      `yaml:"sections,omitempty"`
	Package  string                        `yaml:"package,omitempty"`
	Packages []string                      `yaml:"packages,omitempty"`
	Append   string                        `yaml:"append,omitempty"`
	Prepend  string                        `yaml:"prepend,omitempty"`
	Replace  []PackageConfigOverrideFromTo `yaml:"replace,omitempty"`
}

func LoadConfig(pkgbase string) (*PackageConfig, error) {
	packageConfig := &PackageConfig{}

	err := packageConfig.Load(pkgbase)

	if err != nil {
		return nil, err
	}

	return packageConfig, err
}

func (pconfig *PackageConfig) Load(pkgbase string) error {
	configPath := config.GetConfigPath(pkgbase)
	_, err := os.Stat(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	data, err := os.ReadFile(configPath)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, pconfig)
}

func (pconfig *PackageConfig) Write(pkgbase string) error {
	configPath := config.GetConfigPath(pkgbase)
	data, err := yaml.Marshal(pconfig)

	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0666)
}
