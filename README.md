# changesets

[![Build](https://github.com/nesymno/changesets/actions/workflows/ci.yml/badge.svg)](https://github.com/nesymno/changesets/actions/workflows/ci.yml)
[![Tests](https://github.com/nesymno/changesets/actions/workflows/ci.yml/badge.svg?event=push)](https://github.com/nesymno/changesets/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/nesymno/changesets/branch/main/graph/badge.svg)](https://codecov.io/gh/nesymno/changesets)
[![Go Report Card](https://goreportcard.com/badge/github.com/nesymno/changesets)](https://goreportcard.com/report/github.com/nesymno/changesets)
[![License](https://img.shields.io/github/license/nesymno/changesets)](LICENSE)
[![Release](https://img.shields.io/github/v/release/nesymno/changesets)](https://github.com/nesymno/changesets/releases/latest)

A simple, lightweight CLI tool for managing changelogs with [semantic versioning](https://semver.org/), inspired by [@changesets/changesets](https://github.com/changesets/changesets) - but built for Go projects.

Instead of manually editing a `CHANGELOG.md` and figuring out version bumps, you describe each change as a small markdown file (a "changeset"). When you're ready to release, the tool determines the correct version bump, generates a structured changelog, and cleans up after itself.

## Features

- **Interactive changeset creation** - prompted bump type selection (`patch` / `minor` / `major`) and summary input
- **Automatic version calculation** - determines the next semver version based on pending changesets
- **Structured changelog generation** - groups changes by bump type, includes git commit SHAs
- **Human-readable changeset filenames** - randomly generated slugs like `brave-orange-fox.md`
- **Minimal footprint** - single binary, no config files outside your repo, only one external dependency ([semver](https://github.com/Masterminds/semver))

## Installation

### Using `go install` (recommended)

```bash
go install github.com/nesymno/changesets@latest
```

This installs the `changesets` binary to your `$GOPATH/bin` (or `$GOBIN`).

### Building from source

```bash
git clone https://github.com/nesymno/changesets.git
cd changesets
go build -o changesets .
```

## Quick Start

```bash
# 1. Initialize changesets in your project
changesets init

# 2. Make some changes to your code, then describe them
changesets add

# 3. Preview the next version
changesets next

# 4. Release: bump version, generate changelog, clean up
changesets release
```

## Commands

### `changesets init`

Initializes the `.changesets` directory structure in your project root (detected by walking up to find `go.mod`).

```bash
changesets init
```

Creates the following structure:

```
.changesets/
├── config.json          # Tracks the current version (e.g. {"version": "v0.0.0"})
├── README.md            # Short guide for contributors
└── changes/
    └── .gitkeep
```

If `.changesets/` already exists, you will be prompted to confirm before it is recreated.

### `changesets add`

Interactively creates a new changeset file describing your change.

```bash
changesets add
```

You will be prompted to:

1. Select a bump type (`patch`, `minor`, or `major`)
2. Enter a summary of the change
3. Preview and confirm the changeset

A changeset file is created in `.changesets/changes/` with a random human-readable name:

```
.changesets/changes/brave-orange-fox.md
```

The file uses a simple frontmatter format:

```markdown
---
changesets: minor
---

Added support for custom changelog templates
```

### `changesets next`

Calculates and prints the next version based on all pending changesets. The highest bump type wins: if any changeset is `major`, the next version is a major bump; if any is `minor` (and none are `major`), it's a minor bump; otherwise it's a patch.

```bash
changesets next
# => v1.2.0
```

### `changesets release`

Performs the full release process:

1. Computes the next version from all pending changesets
2. Prepends a new section to `CHANGELOG.md`, grouped by change type, with git commit SHAs
3. Updates the version in `.changesets/config.json`
4. Removes all processed changeset files from `.changesets/changes/`

```bash
changesets release
# => v1.2.0
```

The generated changelog entry looks like this:

```markdown
## v1.2.0 - 2026-02-14

### Minor Changes

- a1b2c3d: Added support for custom changelog templates

### Patch Changes

- e4f5g6h: Fixed typo in error message
```

## Recommended Workflow

### During development

Each time you make a meaningful change, add a changeset to describe it:

```bash
changesets add
git add .changesets/changes/
git commit -m "Add changeset for feature X"
```

Changeset files should be committed alongside the code they describe. This way, pull requests carry their own release metadata.

### When releasing

```bash
# Generate changelog and bump version
version=$(changesets release)

# Commit the release artifacts
git add .
git commit -m "Release ${version}"

# Tag and push
git tag "${version}"
git push origin main --tags
```

### CI integration

You can use `changesets next` in CI pipelines to determine the upcoming version, or check for pending changesets to gate releases:

```bash
# Fail if there are no changesets (useful as a PR check)
changesets next || exit 1
```

## Requirements

- **Go 1.25+** (for building / installing)
- **Git** (optional) - used to resolve commit SHAs in changelog entries. If git is not installed or the project is not a git repository, the tool still works - changelog entries will simply omit commit SHAs

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b my-feature`)
3. Make your changes
4. Add a changeset describing your change (`go run . add`)
5. Run tests (`just test`)
6. Commit and push your branch
7. Open a Pull Request

Please include a changeset with every PR that affects user-facing behavior.

## License

See [LICENSE](LICENSE).
