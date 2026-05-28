#!/bin/bash

set -euo pipefail

version=$(go run . release)

git add .
git commit -m "Release ${version}"
if [[ "$(git tag -l ${version})" == "" ]]; then
    git tag "${version}"
fi
git push origin main --tags

notes=$(awk '/^## /{c++; if(c==2){exit}} c>=1' CHANGELOG.md)

gh release create "${version}" --title "${version}" --notes "${notes}" --draft
