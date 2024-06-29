package config

import (
	"fmt"
	"path/filepath"
)

func (config *Config) GetBasePath() string {
	basePath := config.BasePath

	if basePath == "" {
		basePath = "packages"
	}

	result, _ := filepath.Abs(basePath)

	return result
}

func (config *Config) GetPackagePath(pkgbase string) string {
	return filepath.Join(config.GetBasePath(), pkgbase)
}

func (config *Config) GetConfigPath(pkgbase string) string {
	configPath := config.ConfigPath

	if configPath == "" {
		configPath = "config.yaml"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), configPath)
}

func (config *Config) GetLocalPath(pkgbase string) string {
	localPath := config.LocalPath

	if localPath == "" {
		localPath = "local"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), localPath)
}

func (config *Config) GetMergedPath(pkgbase string) string {
	mergedPath := config.MergedPath

	if mergedPath == "" {
		mergedPath = "merged"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), mergedPath)
}

func (config *Config) GetScriptsPath(pkgbase string) string {
	scriptsPath := config.ScriptsPath

	if scriptsPath == "" {
		scriptsPath = "scripts"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), scriptsPath)
}

func (config *Config) GetScriptOverridePath(pkgbase string) string {
	scriptsPath := config.ScriptOverridePath

	if scriptsPath == "" {
		scriptsPath = "script-override"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), scriptsPath)
}

func (config *Config) GetUpstreamPath(pkgbase string) string {
	upstreamPath := config.UpstreamPath

	if upstreamPath == "" {
		upstreamPath = "upstream"
	}

	return filepath.Join(config.GetPackagePath(pkgbase), upstreamPath)
}

func (config *Config) GetAurBaseUrl() string {
	baseUrl := config.AurBaseUrl

	if baseUrl == "" {
		baseUrl = "https://aur.archlinux.org"
	}

	return baseUrl
}

func (config *Config) GetAurPackagesUrl() string {
	baseUrl := config.GetAurBaseUrl()
	packagesPath := config.AurPackagesPath

	if packagesPath == "" {
		packagesPath = "pkgbase.gz"
	}

	return fmt.Sprintf("%s/%s", baseUrl, packagesPath)
}

func (config *Config) GetAurPackageGitUrl(pkgbase string) string {
	baseUrl := config.GetAurBaseUrl()

	return fmt.Sprintf("%s/%s.git", baseUrl, pkgbase)
}

func (config *Config) GetArchBaseGitUrl() string {
	baseUrl := config.ArchBaseGitUrl

	if baseUrl == "" {
		baseUrl = "https://gitlab.archlinux.org/archlinux"
	}

	return baseUrl
}

func (config *Config) GetArchPackageGitUrl(pkgbase string) string {
	baseUrl := config.GetArchBaseGitUrl()

	return fmt.Sprintf("%s/packaging/packages/%s.git", baseUrl, pkgbase)
}
