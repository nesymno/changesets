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
    bash scripts/release.sh
