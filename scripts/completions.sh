#!/bin/sh
set -eu

rm -rf completions
mkdir completions
for shell in bash zsh fish; do
  go run ./cmd/clean-my-cache completion "$shell" > "completions/clean-my-cache.$shell"
done
