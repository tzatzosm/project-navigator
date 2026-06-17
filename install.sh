#!/usr/bin/env bash
#
# Install the `pn` (project-navigator) binary from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/tzatzosm/project-navigator/main/install.sh | bash
#
# Environment overrides:
#   PN_VERSION=v0.2.0   install a specific tag (default: latest)
#   PN_INSTALL_DIR=...  install location (default: /usr/local/bin if writable,
#                       else ~/.local/bin)
#   PN_NO_WRAPPER=1     do not add the pn() shell wrapper to your shell rc
set -euo pipefail

REPO="tzatzosm/project-navigator"
BINARY="pn"
VERSION="${PN_VERSION:-latest}"
INSTALL_DIR="${PN_INSTALL_DIR:-}"

info() { printf '%s\n' "$*" >&2; }
err() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

need() { command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"; }
need curl
need tar
need uname

# --- Detect platform ---------------------------------------------------------
os="$(uname -s)"
case "$os" in
  Linux) OS=linux ;;
  Darwin) OS=darwin ;;
  *) err "unsupported OS: $os — try 'go install github.com/${REPO}/cmd/pn@latest'" ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) ARCH=amd64 ;;
  arm64 | aarch64) ARCH=arm64 ;;
  *) err "unsupported architecture: $arch" ;;
esac

# --- Resolve version ---------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' | cut -d '"' -f4)"
  [ -n "$TAG" ] || err "could not determine the latest release tag"
else
  TAG="$VERSION"
fi
VER="${TAG#v}" # GoReleaser strips the leading v in asset names

ASSET="pn_${VER}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/${REPO}/releases/download/${TAG}"

# --- Download ----------------------------------------------------------------
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

info "Downloading ${ASSET} (${TAG})…"
curl -fsSL "${BASE}/${ASSET}" -o "${tmp}/${ASSET}" \
  || err "download failed: ${BASE}/${ASSET}"

# --- Verify checksum (required) ----------------------------------------------
if command -v sha256sum >/dev/null 2>&1; then
  sum="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sum="shasum -a 256"
else
  err "no SHA-256 tool (sha256sum or shasum) found — cannot verify the download"
fi

curl -fsSL "${BASE}/checksums.txt" -o "${tmp}/checksums.txt" \
  || err "could not download checksums.txt — refusing to install unverified binary"

# Ensure there is actually an entry for our asset (an empty match would make
# `$sum -c` succeed on zero lines and falsely report success).
grep " ${ASSET}\$" "${tmp}/checksums.txt" >/dev/null \
  || err "no checksum entry for ${ASSET} in checksums.txt"

if (cd "$tmp" && grep " ${ASSET}\$" checksums.txt | $sum -c -) >/dev/null 2>&1; then
  info "Checksum verified."
else
  err "checksum verification failed for ${ASSET}"
fi

# --- Extract -----------------------------------------------------------------
tar -xzf "${tmp}/${ASSET}" -C "$tmp"
[ -f "${tmp}/${BINARY}" ] || err "binary '${BINARY}' not found in archive"
chmod +x "${tmp}/${BINARY}"

# --- Install -----------------------------------------------------------------
if [ -z "$INSTALL_DIR" ]; then
  if [ -w /usr/local/bin ] 2>/dev/null; then
    INSTALL_DIR=/usr/local/bin
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi
mkdir -p "$INSTALL_DIR"

if mv "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}" 2>/dev/null; then
  :
elif command -v sudo >/dev/null 2>&1; then
  info "Installing to ${INSTALL_DIR} (requires sudo)…"
  sudo mv "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  err "cannot write to ${INSTALL_DIR}; set PN_INSTALL_DIR to a writable directory"
fi

info "Installed ${BINARY} ${TAG} to ${INSTALL_DIR}/${BINARY}"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) info "Note: ${INSTALL_DIR} is not on your PATH. Add: export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
esac

# --- Shell wrapper (cd integration) ------------------------------------------
# `pn` prints `cd <path>`; a shell function must eval it. We append that wrapper
# to the user's shell rc unless it's already there (set PN_NO_WRAPPER=1 to skip).
# shellcheck disable=SC2016  # the wrapper body must stay literal in the rc file
WRAPPER='
# project-navigator: let `pn` change the shell directory
pn() {
  result=$(command pn "$@")
  if echo "$result" | grep -q "^cd "; then
    eval "$result"
  else
    echo "$result"
  fi
}'

# Marker used for the idempotency check.
WRAPPER_MARK='command pn "$@"'

rc=""
case "$(basename "${SHELL:-}")" in
  zsh) rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
  bash) rc="${HOME}/.bashrc" ;;
esac

if [ "${PN_NO_WRAPPER:-0}" = "1" ]; then
  : # user opted out
elif [ -z "$rc" ]; then
  info ""
  info "Add this wrapper to your shell rc to enable 'cd' into projects:"
  printf '%s\n' "$WRAPPER" >&2
elif [ -f "$rc" ] && grep -qF "$WRAPPER_MARK" "$rc"; then
  info "Shell wrapper already present in ${rc}."
else
  printf '%s\n' "$WRAPPER" >>"$rc"
  info "Added the pn() shell wrapper to ${rc}. Reload your shell: exec \$SHELL"
fi
