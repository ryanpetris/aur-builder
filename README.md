# AUR Builder

AUR Builder is a personal repository builder, allowing synchronization and overrides of packages from the AUR, from the official Arch repositories, or from packages local to your repository. This is not a traditional AUR helper in the sense that it will automatically download, compile, and install a package from the AUR, but helps manage your own personal repository.

## FAQ

### What does this tool do?

* Tracks changes to AUR and official Arch repository packages that you want to build and create a personal repository for.
* Creates pull requests for you to manually verify changes made to packages.
* Can be used for private personal packages in addition to AUR and official Arch packages.

### What does this tool NOT do?

* Build packages
* Creates/manages a package repository

You will need to use a CI environment as well as other tools to do the actual package building and repository creation/management.

## Directory Structure

In the directory aur-builder is run from, it is expecting the following directory structure:

```
packages
    package1
        config.yaml
        local
        merged
        scripts
        upstream
    package2
    ...
```

If you're starting from scratch, you'll need to manually create the `packages` directory and commit it to your repository. Use a `.gitkeep` file in the `packages` directory to commit an otherwise empty directory.

### config.yaml

This is the configuration file for this package. If this package does not come from the AUR or the official Aarch repositories, then this file may be missing or empty.

See the [configuration](#configuration) section for configuration options.

### local

This contains any local files to be added to the package. Note that this folder is eventually overlaid on the upstream repository, and therefore any file with the same name will replace the upstream file.

### upstream

This contains a copy of the upstream package, containing all of the files that would normally be downloaded with the package sources. You should not make any manual changes to this repository as any changes will be overwritten when the package is updated.

### scripts

This can contain the following scripts:

* onprepare.sh - Runs before any file copying happens. The working directory will be the `merged` directory, of which nothing will exist yet.
* onmerge.sh - Runs after merging and overrides have been applied. The working directory is also the `merged` directory.

## Commands

### Import

The `import` command imports a package from the aur or the official Arch repositories. This will automatically create a `config.yaml` and populate the `upstream` folder with the contents of the package from the AUR or the official Arch repositories. If this command is run from a supported CI environment, a pull request will automatically be created.

Example:

```shell
aur-builder import --source aur --package yay # imports the yay package from the aur
aur-builder import --source arch --package tailscale # imports the yay package from the official arch repository
```

Note that the package name is really the `pkgbase`, thus you should use the base package name for any packages that contain multiple packages.

### Update

The `update` command looks for updates in the AUR or official Arch repositories and updates the `upstream` folder with the updates. If this command is run from a supported CI environment, a pull request will automatically be created.

Example:

```shell
aur-builder update --source aur # checks for updates for aur packages
aur-builder update --source arch # checks for updates for official arch packages
```

### Prepare

The `prepare` command generates the `merged` folder for all packages in the repository with the following steps:

1. Empties merged directory if it exists, or creates it if it does not exist.
2. Runs onprepare.sh script, if present
3. Copies upstream contents to merged directory
4. Copies local contents to merged directory (which will overwrite any files with the same name)
5. Processes overrides from the config.yaml file
6. Runs onmerge.sh script, if present
7. Generates the .SRCINFO file in the merged directory

Example:

```shell
aur-builder prepare # prepares all packages
aur-builder prepare --package yay # prepares only the yay package
```

### Needs Build

The `needs-build` command checks if any packages need to be built. Note that versions are compared against your local sync DB, and therefore it should be up to date prior to running this. As this tool is intended to be run from a CI environment, this is generally not an issue.

```shell
aur-builder needs-build
```

## Configuration

TODO