# Run the CLI with optional arguments: just run init, add, version, release
run *ARGS:
    go run main.go {{ARGS}}

# Run all tests
test:
    go test ./... -v

release:
    go run main.go release
    git add .
    git commit -m "Release $(changesets next)"
    git tag "$(changesets next)"
    git push origin main --tags