# Changelog

## 0.1.1

- Fix the Homebrew cask, which Homebrew could not read in 0.1.0

## 0.1.0

- Single binary for macOS (Apple Silicon and Intel), distributed with Homebrew, an install script and `go install`
- Interactive scan and clean with size, risk label and explanation for each item
- 38 built-in recipes as YAML files, with a JSON Schema and an Electron preset
- Custom recipes in `~/.config/clean-my-cache/recipes`
- `scan`, `list`, `explain` and `completion` commands
- `--only`, `--exclude`, `--all`, `--trash`, `--yes`, `--json` and `--verbose`
- Path guard that refuses anything outside known cache folders and symbolic links, checked at load, scan and delete time
- Open apps are skipped and listed with the command to clean them later
