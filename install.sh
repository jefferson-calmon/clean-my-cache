#!/bin/sh
set -eu

REPO="jefferson-calmon/clean-my-cache"
BINARY="clean-my-cache"
INSTALL_DIR="${CLEAN_MY_CACHE_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
  printf 'x %s\n' "$1" >&2
  exit 1
}

[ "$(uname -s)" = "Darwin" ] || fail "clean-my-cache supports macOS only for now."

case "$(uname -m)" in
  arm64) arch="arm64" ;;
  x86_64) arch="amd64" ;;
  *) fail "Unsupported processor: $(uname -m)" ;;
esac

archive="${BINARY}_darwin_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/latest/download"
workdir="$(mktemp -d)"
trap 'rm -r "$workdir"' EXIT

printf '~ Downloading %s\n' "$archive"
curl -fsSL "$base/$archive" -o "$workdir/$archive" || fail "Download failed."
curl -fsSL "$base/checksums.txt" -o "$workdir/checksums.txt" || fail "Could not download checksums."

expected="$(grep " $archive\$" "$workdir/checksums.txt" | cut -d ' ' -f 1)"
actual="$(shasum -a 256 "$workdir/$archive" | cut -d ' ' -f 1)"
[ -n "$expected" ] && [ "$expected" = "$actual" ] || fail "Checksum mismatch. Nothing was installed."

tar -xzf "$workdir/$archive" -C "$workdir" "$BINARY"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$workdir/$BINARY" "$INSTALL_DIR/$BINARY"
xattr -d com.apple.quarantine "$INSTALL_DIR/$BINARY" 2>/dev/null || true

printf '+ Installed %s to %s\n' "$BINARY" "$INSTALL_DIR"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) printf '\nRun: %s\n' "$BINARY" ;;
  *) printf '\n! %s is not in your PATH. Add this line to ~/.zshrc:\n  export PATH="%s:$PATH"\n' "$INSTALL_DIR" "$INSTALL_DIR" ;;
esac
