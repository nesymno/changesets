# Run the CLI with optional arguments: just run init, add, version, release
run *ARGS:
    go run main.go {{ARGS}}

# Run all tests
test:
    go test ./... -v

next:
    go run main.go next

release:
    go run main.go release
    git add .
    git commit -m "Release $(just next)"
    git tag "$(just next)"
    git push origin main --tags