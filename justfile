# Run the CLI with optional arguments: just run init, add, next, release
run *ARGS:
    go run . {{ ARGS }}

# Build the binary with version from config.json
build VERSION="dev":
    go build -ldflags "-X main.version={{ VERSION }}" -o changesets .

# Run all tests
test:
    go test ./... -v

test_cover:
    go test ./... -v -coverprofile=coverage.txt -covermode=atomic

next:
    go run . next

release:
    #!/usr/bin/env bash
    set -euo pipefail
    version=$(go run . next)
    git add .
    git commit -m "Release ${version}"
    if [[ "$(git tag -l ${version})" == "" ]]; then
        git tag "${version}"
    fi
    git push origin main --tags
    notes=$(awk '/^## /{if(c++){exit}else{next}} c' CHANGELOG.md)
    gh release create "${version}" --title "${version}" --notes "${notes}" --draft
