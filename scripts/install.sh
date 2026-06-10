#!/bin/bash
set -euo pipefail

# rem installer — downloads the latest release from GitHub
# Usage: curl -fsSL https://rem.sidv.dev/install | bash

REPO="BRO3886/rem"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="rem"

info() { printf "\033[36m%s\033[0m\n" "$*"; }
warn() { printf "\033[33m%s\033[0m\n" "$*"; }
error() { printf "\033[31mError: %s\033[0m\n" "$*" >&2; exit 1; }

# --- Pre-flight checks ---

if [ "$(uname -s)" != "Darwin" ]; then
    error "rem only supports macOS"
fi

ARCH="$(uname -m)"
case "$ARCH" in
    arm64|aarch64) ARCH="arm64" ;;
    x86_64)        ARCH="amd64" ;;
    *)             error "Unsupported architecture: $ARCH" ;;
esac

if ! command -v curl >/dev/null 2>&1; then
    error "curl is required but not found"
fi

# --- Resolve latest version ---

info "Fetching latest release..."
LATEST=$(curl -sSL -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    error "Could not determine latest release"
fi

info "Latest version: $LATEST"

# --- Download and extract ---

ASSET_NAME="rem-darwin-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET_NAME}"

TMPDIR_PATH=$(mktemp -d)
trap 'rm -rf "$TMPDIR_PATH"' EXIT

info "Downloading ${ASSET_NAME}..."
HTTP_CODE=$(curl -sSL -w "%{http_code}" -o "${TMPDIR_PATH}/${ASSET_NAME}" "$DOWNLOAD_URL")

if [ "$HTTP_CODE" != "200" ]; then
    error "Download failed (HTTP $HTTP_CODE). Asset '${ASSET_NAME}' may not exist for ${LATEST}."
fi

tar -xzf "${TMPDIR_PATH}/${ASSET_NAME}" -C "${TMPDIR_PATH}"

# --- Install ---

if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "${TMPDIR_PATH}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    info "Requires sudo to install to ${INSTALL_DIR}"
    sudo mv "${TMPDIR_PATH}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

info "Installed rem ${LATEST} to ${INSTALL_DIR}/${BINARY_NAME}"

# --- Old install cleanup notice ---

if [ "$INSTALL_DIR" != "/usr/local/bin" ] && [ -e "/usr/local/bin/${BINARY_NAME}" ]; then
    warn "An older rem exists at /usr/local/bin/${BINARY_NAME}."
    warn "Remove it to avoid PATH confusion: sudo rm /usr/local/bin/${BINARY_NAME}"
fi

# --- PATH check ---

case ":$PATH:" in
    *":${INSTALL_DIR}:"*)
        info "Run 'rem --help' to get started"
        ;;
    *)
        warn "${INSTALL_DIR} is not in your PATH."
        case "${SHELL:-}" in
            */zsh)
                warn "Add it with: echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
                ;;
            */fish)
                warn "Add it with: fish_add_path ${INSTALL_DIR}"
                ;;
            *)
                warn "Add it with: echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
                ;;
        esac
        ;;
esac

# --- Agent skill installation ---

echo ""
info "rem can install an AI agent skill that teaches Claude Code / Codex / OpenClaw how to use it."
printf "Install agent skill now? [Y/n] "
{ read -r answer < /dev/tty; } 2>/dev/null || answer="n"
if [ "$answer" != "n" ] && [ "$answer" != "N" ]; then
    "${INSTALL_DIR}/${BINARY_NAME}" skills install || true
fi
