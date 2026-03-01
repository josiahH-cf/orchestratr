#!/usr/bin/env bash
# install.sh — Install orchestratr on Linux/macOS
#
# Usage:
#   curl -sSL <url>/install.sh | bash
#   ./install.sh
#
# This script:
#   1. Checks prerequisites (Go toolchain)
#   2. Builds orchestratr from source (or uses pre-built binary if available)
#   3. Installs the binary to ~/.local/bin (or $INSTALL_DIR)
#   4. Runs 'orchestratr install' to configure autostart
#
# Environment variables:
#   INSTALL_DIR — override install directory (default: ~/.local/bin)
#   SKIP_BUILD  — set to 1 to skip building (use existing binary in PATH)

set -euo pipefail

readonly BINARY_NAME="orchestratr"
readonly DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
INSTALL_DIR="${INSTALL_DIR:-${DEFAULT_INSTALL_DIR}}"

info()  { printf '  \033[1;34m→\033[0m %s\n' "$*"; }
warn()  { printf '  \033[1;33m⚠\033[0m %s\n' "$*"; }
error() { printf '  \033[1;31m✗\033[0m %s\n' "$*" >&2; }
ok()    { printf '  \033[1;32m✓\033[0m %s\n' "$*"; }

check_prereqs() {
    if [[ "${SKIP_BUILD:-0}" == "1" ]]; then
        if ! command -v "${BINARY_NAME}" &>/dev/null; then
            error "SKIP_BUILD=1 but '${BINARY_NAME}' not found in PATH"
            exit 1
        fi
        info "Using existing ${BINARY_NAME} from PATH"
        return
    fi

    if ! command -v go &>/dev/null; then
        error "Go toolchain not found. Install Go from https://go.dev/dl/"
        exit 1
    fi

    local go_version
    go_version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
    info "Found Go ${go_version}"
}

build_binary() {
    if [[ "${SKIP_BUILD:-0}" == "1" ]]; then
        return
    fi

    # Determine source directory.
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    if [[ -f "${script_dir}/go.mod" ]]; then
        info "Building from source in ${script_dir}"
        (cd "${script_dir}" && go build -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/orchestratr)
    else
        info "Building via go install"
        GOBIN="${INSTALL_DIR}" go install github.com/josiahH-cf/orchestratr/cmd/orchestratr@latest
    fi
}

install_binary() {
    mkdir -p "${INSTALL_DIR}"

    if [[ "${SKIP_BUILD:-0}" == "1" ]]; then
        # Binary already in PATH; just verify.
        local bin_path
        bin_path="$(command -v "${BINARY_NAME}")"
        ok "Using ${bin_path}"
        return
    fi

    build_binary

    local bin_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [[ ! -x "${bin_path}" ]]; then
        error "Build failed — ${bin_path} not found"
        exit 1
    fi

    chmod +x "${bin_path}"
    ok "Binary installed to ${bin_path}"

    # Ensure install dir is in PATH.
    if ! echo "${PATH}" | tr ':' '\n' | grep -qx "${INSTALL_DIR}"; then
        warn "${INSTALL_DIR} is not in your PATH"
        warn "Add to your shell profile:  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

run_install() {
    local bin_path
    if [[ "${SKIP_BUILD:-0}" == "1" ]]; then
        bin_path="$(command -v "${BINARY_NAME}")"
    else
        bin_path="${INSTALL_DIR}/${BINARY_NAME}"
    fi

    info "Running '${BINARY_NAME} install'..."
    "${bin_path}" install
}

main() {
    echo ""
    echo "  orchestratr installer"
    echo "  ====================="
    echo ""

    check_prereqs
    install_binary
    run_install

    echo ""
    ok "Installation complete!"
    echo ""
    info "Start the daemon:   ${BINARY_NAME} start"
    info "Check status:       ${BINARY_NAME} status"
    info "View config:        ${BINARY_NAME} configure"
    echo ""
}

main "$@"
