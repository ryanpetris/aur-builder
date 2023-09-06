package pkg

import (
	"errors"
	"github.com/ryanpetris/aur-builder/config"
	"gopkg.in/yaml.v3"
	"os"
)

type PackageConfig struct {
	Source    string                 `yaml:"source,omitempty"`
	Overrides PackageConfigOverrides `yaml:"overrides,omitempty"`
}

type PackageConfigOverrides struct {
	AppendPkgbuild       string                        `yaml:"appendPkgbuild,omitempty"`
	BumpPkgrel           map[string]int                `yaml:"bumpPkgrel,omitempty"`
	ClearConflicts       bool                          `yaml:"clearConflicts,omitempty"`
	ClearDependsVersions bool                          `yaml:"clearDependsVersions,omitempty"`
	ClearPkgverFunc      bool                          `yaml:"clearPkgverFunc,omitempty"`
	ClearProvides        bool                          `yaml:"clearProvides,omitempty"`
	ClearReplaces        bool                          `yaml:"clearReplaces,omitempty"`
	ClearSignatures      bool                          `yaml:"clearSignatures,omitempty"`
	DeleteFile           []string                      `yaml:"deleteFile,omitempty"`
	RenameFile           []PackageConfigOverrideFromTo `yaml:"renameFile,omitempty"`
	RenameFunction       []PackageConfigOverrideFromTo `yaml:"renameFunction,omitempty"`
	RenamePackage        []PackageConfigOverrideFromTo `yaml:"renamePackage,omitempty"`
	ReplacePkgbuild      []PackageConfigOverrideFromTo `yaml:"replacePkgbuild,omitempty"`
}

type PackageConfigOverrideFromTo struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`
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
