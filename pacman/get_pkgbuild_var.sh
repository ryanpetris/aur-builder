#!/bin/bash

set -e
set -o pipefail

PKGBUILD_VAR="${PKGBUILD_VAR}[*]"

pushd "$(dirname "$PKGBUILD_PATH")" >/dev/null
source "$(basename "$PKGBUILD_PATH")"
export IFS=$'\n'
echo "${!PKGBUILD_VAR}"
