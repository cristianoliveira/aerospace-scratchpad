#!/usr/bin/env bash

NIX_FILE=$1
FLAKE_TASK=$2

VERSION="$(git rev-parse --short HEAD)"

echo "Bumping version in $VERSION"
# replace  version = "local-2024-09-16"; within "$NIX_FILE"
# sed -i "s/version = \".*\";/version = \"$VERSION\";/" "$NIX_FILE"
sed -i 's/sha256-.*=//g' "$NIX_FILE"
## nix build and pipe the error to a build.log file
rm -f build.log

nix build $FLAKE_TASK 2> build.log
echo "aaaaa"

SHA256=$(grep "got:" build.log | grep -o "sha256-.*=" | cut -d'-' -f2)
echo "nix hash SHA256: $SHA256"

echo "git hash SHA256: $SHA256"
sed -i "s# sha256 = \".*\";# sha256 = \"sha256-$SHA256\";#" "$NIX_FILE"
nix build $FLAKE_TASK 2> build.log

SHA256=$(grep "got:" build.log | grep -o "sha256-.*=" | cut -d'-' -f2)
echo "vendorHash hash SHA256: $SHA256"
sed -i "s#vendorHash = \".*\";#vendorHash = \"sha256-$SHA256\";#" "$NIX_FILE"

echo "Building nix derivation"
nix build $FLAKE_TASK

rm -f build.log

if [ -z "$CI" ]; then
  echo "Not in CI, committing changes"
else
  echo "In CI, skipping commit"
  exit 0
fi

git add "$NIX_FILE"
git commit -m "chore(nix): bump nightly ($VERSION)"
