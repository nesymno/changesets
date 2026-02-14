# Run the CLI with optional arguments: just run init, add, version, release
run *ARGS:
    go run main.go {{ARGS}}

# Run all tests
test:
    go test ./... -v
