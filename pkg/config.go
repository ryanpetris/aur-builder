package pkg

import (
	"bytes"
	"errors"
	"github.com/ryanpetris/aur-builder/config"
	"gopkg.in/yaml.v3"
	"os"
)

type PackageConfig struct {
	Source    string                 `yaml:"source,omitempty"`
	Overrides PackageConfigOverrides `yaml:"overrides,omitempty"`
	Ignore    bool                   `yaml:"ignore,omitempty"`
	Vcs       *PackageVcs            `yaml:"vcs,omitempty"`
}

type PackageConfigOverrides struct {
	BumpEpoch            int                           `yaml:"bumpEpoch,omitempty"`
	BumpPkgrel           map[string]int                `yaml:"bumpPkgrel,omitempty"`
	ClearDependsVersions bool                          `yaml:"clearDependsVersions,omitempty"`
	ClearSignatures      bool                          `yaml:"clearSignatures,omitempty"`
	DeleteFile           []string                      `yaml:"deleteFile,omitempty"`
	ModifySection        []PackageConfigModifySection  `yaml:"modifySection,omitempty"`
	RemoveSource         []string                      `yaml:"removeSource,omitempty"`
	RenameFile           []PackageConfigOverrideFromTo `yaml:"renameFile,omitempty"`
	RenamePackage        []PackageConfigOverrideFromTo `yaml:"renamePackage,omitempty"`
}

type PackageConfigOverrideFromTo struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`
}

type PackageConfigModifySection struct {
	Type     string                        `yaml:"type,omitempty"`
	Section  string                        `yaml:"section,omitempty"`
	Sections []string                      `yaml:"sections,omitempty"`
	Package  string                        `yaml:"package,omitempty"`
	Packages []string                      `yaml:"packages,omitempty"`
	Append   string                        `yaml:"append,omitempty"`
	Prepend  string                        `yaml:"prepend,omitempty"`
	Replace  []PackageConfigOverrideFromTo `yaml:"replace,omitempty"`
	Rename   string                        `yaml:"rename,omitempty"`
}

type PackageVcs struct {
	Pkgver          string                        `yaml:"pkgver,omitempty"`
	Pkgrel          int                           `yaml:"pkgrel,omitempty"`
	SourceOverrides []PackageConfigOverrideFromTo `yaml:"sourceOverrides,omitempty"`
}

func LoadConfig(pkgbase string) (*PackageConfig, error) {
	packageConfig := &PackageConfig{}

	err := packageConfig.Load(pkgbase)

	if err != nil {
		return nil, err
	}

	return packageConfig, err
}

func ConfigExists(pkgbase string) (bool, error) {
	configPath := config.GetConfigPath(pkgbase)
	_, err := os.Stat(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
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
	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)

	encoder.SetIndent(2)
	err := encoder.Encode(pconfig)

	if err != nil {
		return err
	}

	return os.WriteFile(configPath, buffer.Bytes(), 0666)
}

func (vcinfo *PackageVcs) IsEqual(newVcinfo *PackageVcs) bool {
	if vcinfo == nil && newVcinfo == nil {
		return true
	}

	if vcinfo == nil || newVcinfo == nil {
		return false
	}

	if vcinfo.Pkgver != newVcinfo.Pkgver {
		return false
	}

	for _, override := range vcinfo.SourceOverrides {
		found := false

		for _, newOverride := range newVcinfo.SourceOverrides {
			if override.From == newOverride.From {
				found = true

				if override.To != newOverride.To {
					return false
				}
			}

			if found {
				break
			}
		}

		if !found {
			return false
		}
	}

	for _, newOverride := range newVcinfo.SourceOverrides {
		found := false

		for _, override := range vcinfo.SourceOverrides {
			if override.From == newOverride.From {
				found = true
			}

			if found {
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
