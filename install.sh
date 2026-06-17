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

# --- Verify checksum (best effort) -------------------------------------------
if curl -fsSL "${BASE}/checksums.txt" -o "${tmp}/checksums.txt" 2>/dev/null; then
  sum=""
  if command -v sha256sum >/dev/null 2>&1; then
    sum="sha256sum"
  elif command -v shasum >/dev/null 2>&1; then
    sum="shasum -a 256"
  fi
  if [ -n "$sum" ]; then
    if (cd "$tmp" && grep " ${ASSET}\$" checksums.txt | $sum -c -) >/dev/null 2>&1; then
      info "Checksum verified."
    else
      err "checksum verification failed for ${ASSET}"
    fi
  fi
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

cat >&2 <<'EOS'

To enable `cd` into projects, add this wrapper to your ~/.zshrc or ~/.bashrc:

  pn() {
    result=$(command pn "$@")
    if echo "$result" | grep -q "^cd "; then
      eval "$result"
    else
      echo "$result"
    fi
  }
EOS
