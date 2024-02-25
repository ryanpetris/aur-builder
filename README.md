# AUR Builder

AUR Builder is a personal repository builder, allowing synchronization and overrides of packages from the AUR, from the official Arch repositories, or from packages local to your repository. This is not a traditional AUR helper in the sense that it will automatically download, compile, and install a package from the AUR, but helps manage your own personal repository.

> [!CAUTION]
> -git, -hg, and other source packages are currently not supported. Only packages that are pinned to a specific version are supported.

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

> [!NOTE]
> The package name is really the `pkgbase`, thus you should use the base package name for any packages that contain multiple packages.

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

Example:

```shell
aur-builder prepare # prepares all packages
aur-builder prepare --package yay # prepares only the yay package
```

### Needs Build

The `needs-build` command checks if any packages need to be built. Note that versions are compared against your local sync DB, and therefore it should be up to date prior to running this. As this tool is intended to be run from a CI environment, this is generally not an issue.

Example:

```shell
aur-builder needs-build
```

## Configuration

### Top-Level

* `source` - The source of the package, either `aur` or `arch`. If the package is local to this repository, omit this option.
* `ignore` - Ignores this package, unless explicitly specified via the `--package` argument.
* `overrides` - Overrides for this package. See the [overrides](#overrides) section.

TODO: Document vcs.

### Overrides

* `bumpEpoch` - If specified, will bump the epoch by the specified amount.
* `bumpPkgrel` - If specified, will bump the pkgrel for the specified package versions by the amount specified. Multiple versions can be specified, and when the `update` command is run, obsolete versions will automatically be removed from the configuration. Example:

```yaml
bumpPkgrel:
    1.0.0: 2
```

* `clearDependsVersions` - Sometimes packages are locked to specific versions unnecessarily; this will remove those depends versions. If you need something more granular, you can try the `modifySection` override below.
* `clearSignatures` - This removes all signature files from the sources list along with clearing the `validpgpkeys` section, allowing the package to be built without signatures. Generally packages also have `sums` for all the relevant files and importing signatures can be problematic. This is an alternative of just blindly importing signatures or disabling signatures via the command line in makepkg.
* `deleteFile` - Array of files to delete. This occurs in the `merged` directory after all files are merged.
* `modifySection` - Modifies a section of the pkgbuild file. The behavior depends on whether the section is an array or a function. For more information see the [Modify Section Overrides](#modify-section-overrides) configuration.
* `removeSource` - Array of source files to remove from the PKGBUILD file. These are used as regular expressions and anything matching will be removed along with any matching sums.
* `renameFile` - Array of files to rename in the `merged` directory.
    * `from` - The old name of the file
    * `to` - The new name of the file
* `renamePackage` - Renames a package (*not* pkgbase). This will also rename any relevant functions such as `package`, `prepare`, `build`, and `check` functions specific to the named package.
    * `from` - The old name of the package. Can be omitted if the PKGBUILD only contains a single package, which is most of them.
    * `to` - The new name of the package.

### Modify Section Overrides

Each `modifySection` array item is processed in the order listed in the configuration, and therefore it's possible for these commands to step on each other. Please ensure that subsequent instructions are compatible with the changes made in previous instructions.

* `type` - The type of section to modify. This is optional, however if there's multiple items that have the same name with different types (such as pkgver, which could be a variable AND a function), this will help choose one or the other. Additionally, if the section does not exist, this will allow creation of the section. Value values are "function", "array", or "variable".
* `section` or `sections` - The sections to modify. `sections` is an array while `section` is a single section. If multiple are specified, they must be of the same type, either an array or function.
* `package` or `packages` - The packages this applies to, and is only applicable for functions. `packages` is an array while `package` is a single item. This will limit the matched functions to only those applicable for the specified packages, for instance `package_<pkgname>`. If not specified, only sections not tied to specific packages will be matched.
* `append` - Append to the section. If an array, each line will be added as a separate array item to the beginning of the array. If a section, will be appended as lines to the function.
* `prepend` - Prepend to the section. If an array, each line will be added as a separate array item to the end of the array. If a section, will be prepended as lines to the function.
* `replace` - Replaces matched lines or array items with the result. Note that only the matched section will be replace, not the whole line. Therefore if you intend to replace or remove the whole line, ensure the regular expression covers the whole line.
    * `from` - A regular expression for the line to find
    * `to` - The replacement value.
* `rename` - Renames the section. Only the section name is updated; the package name remains appended to the function/array/variable if applicable.

### Examples

The following are examples from my personal repository.

```yaml
# bruno-bin package
source: aur
overrides:
  modifySection:
    # Removes the "bruno" and "bruno-*" entries from conflicts and provides.
    - sections:
        - conflicts
        - provides
      replace:
        - from: ^["']?bruno(-.*)?["']?$
  renamePackage:
    # Renames the package to "bruno"
    - to: bruno
```

```yaml
# flightgear package
source: aur
overrides:
  modifySection:
    # Fix for MAKEFLAGS="-j$(nproc)" not working correctly.
    # Appends this command to the bottom of the prepare function.
    - section: prepare
      append: |
        echo 'add_dependencies(fgfs embeddedresources)' >> src/Main/CMakeLists.txt
```

```yaml
# brave-bin package
source: aur
overrides:
  modifySection:
    # Add a line to the prepare function to replace "brave-bin" with "brave" in the "brave.sh" file
    - section: prepare
      prepend: |
        sed -i -E 's|brave-bin|brave|g' "${srcdir}/brave.sh"
    # Replace instances of "brave-bin" with "brave" in the package function.
    - section: package
      replace:
        - from: brave-bin
          to: brave
    # Remove conflicts and provides entries that are using substitution along with "brave" and "brave-*" entries.
    - sections:
        - conflicts
        - provides
      replace:
        - from: ^["']?\$.*["']?$
        - from: ^["']?brave(-.*)?["']?$
  renameFile:
    # Renames "brave-bin.sh" file to "brave.sh"
    - from: brave-bin.sh
      to: brave.sh
  renamePackage:
    # Renames package to "brave"
    - to: brave
```

```yaml
# plib package
source: aur
overrides:
  # Removes lines from the prepare function that are copying the config.guess and config.sub files.
  modifySection:
    - section: prepare
      replace:
        - from: (?m)^\s*cp [./]*config\.(guess|sub).*$
  removeSource:
    # Removes the config.guess and config.sub files from sources.
    - ^config\.(guess|sub)$
```
