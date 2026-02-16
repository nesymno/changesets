# Run the CLI with optional arguments: just run init, add, next, release
run *ARGS:
    go run . {{ARGS}}

# Build the binary with version from config.json
build VERSION="dev":
    go build -ldflags "-X main.version={{VERSION}}" -o changesets .

# Run all tests
test:
    go test ./... -v

next:
    go run . next

release:
    #!/usr/bin/env bash
    set -euo pipefail
    version=$(go run . release)
    git add .
    git commit -m "Release ${version}"
    git tag "${version}"
    git push origin main --tags