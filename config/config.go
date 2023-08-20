package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	BasePath     string `yaml:"basePath,omitempty"`
	ConfigPath   string `yaml:"configPath,omitempty"`
	LocalPath    string `yaml:"localPath,omitempty"`
	MergedPath   string `yaml:"mergedPath,omitempty"`
	ScriptsPath  string `yaml:"scriptsPath,omitempty"`
	UpstreamPath string `yaml:"upstreamPath,omitempty"`

	AurBaseUrl      string `yaml:"aurBaseUrl,omitempty"`
	AurPackagesPath string `yaml:"aurPackagesUrl,omitempty"`

	ArchBaseGitUrl string `yaml:"archBaseGitUrl,omitempty"`
}

func (config *Config) Load(cfgpath string) error {
	data, err := os.ReadFile(cfgpath)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return err
}
