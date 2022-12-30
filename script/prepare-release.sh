#!/bin/bash

set -e

# Script to prepare a new rmapi release

SED="sed"

set_sed(){
  if $(which gsed 2>&1 >/dev/null); then
    SED="gsed"
  fi
}

update_app_version(){
  local version=$1
  $SED -i "s/var Version = \".*\"/var Version = \"$version\"/" version/version.go
}

update_changelog(){
  local version=$1
  $SED -i "1c## rmapi $version ($(date "+%B %d, %Y"))" CHANGELOG.md
}

update_macosx_tutorial(){
  local version=$1
  $SED -i "s/v.*\/rmapi-macosx.zip/v${version}\/rmapi-macosx.zip -o rmapi.zip/" docs/tutorial-print-macosx.md
}

create_tag(){
  local version=$1
  git tag v${version}
}

show_git_push(){
  local version=$1
  git diff
  echo 
  echo
  echo "Commit and push current changes with:"
  echo "  git commit version/version.go CHANGELOG.md docs/tutorial-print-macosx.md -m 'Release $version'"
  echo "  git push origin HEAD:master HEAD:refs/tags/v$version"
}

if [ -z "$1" ]; then
  echo "Missing version argument" >&2
  echo "Usage: $0 version" >&2
  echo "Example: $0 0.0.10" >&2
  exit 1
fi

version=$1
set_sed
update_app_version $version
update_changelog $version
update_macosx_tutorial $version
create_tag $version
show_git_push $version
