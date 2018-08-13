#!/bin/bash

set -e

# Get a list of modified go files under our source directory. Ignore vendor files. If none, exit.
git_root=$(git rev-parse --show-toplevel)
gofiles=$(find $git_root -type f -name '*.go' | grep -v "$git_root/vendor")
[ -z "$gofiles" ] && exit 0

# Get the subset of unformatted files. If none, exit.
goimports_cmd="goimports -local github.com/allenai/beaker/"
unformatted=$($goimports_cmd -l $gofiles)
[ -z "$unformatted" ] && exit 0

# If VERIFY is set, perform a strict check. Otherwise just format files.
if [ -z "$VERIFY" ]; then
  for file in $unformatted; do
    echo "$file"
    $goimports_cmd -w "$file"
  done
else
  echo >&2 "Go files must be formatted. Please run:"
  for file in $unformatted; do
    echo >&2 "  $goimports_cmd -w $file"
  done
  exit 1
fi
