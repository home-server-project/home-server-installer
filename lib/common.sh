#!/usr/bin/env bash

set -o pipefail

HSI_VERSION="${HSI_VERSION:-0.1-alpha}"
HSI_TARGET_MOUNT="${HSI_TARGET_MOUNT:-/var/mnt/home-server-target}"
HSI_PODMAN_BIND="${HSI_PODMAN_BIND:-/var/lib/containers}"
HSI_LOG_PREFIX="[home-server-installer]"

log() {
    printf '%s %s\n' "$HSI_LOG_PREFIX" "$*"
}

info() {
    printf 'INFO: %s\n' "$*"
}

warn() {
    printf 'WARNING: %s\n' "$*" >&2
}

fatal() {
    printf 'FATAL: %s\n' "$*" >&2
    exit 1
}

require_cmd() {
    local cmd
    for cmd in "$@"; do
        command -v "$cmd" >/dev/null 2>&1 || fatal "Required command not found: $cmd"
    done
}

require_root() {
    if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
        if command -v sudo >/dev/null 2>&1; then
            exec sudo --preserve-env=HSI_ISO_ROOT "$0" "$@"
        fi
        fatal "Run this installer as root."
    fi
}

require_supported_host() {
    local arch
    arch="$(uname -m)"
    [[ "$arch" == "x86_64" ]] || fatal "Alpha supports x86_64 only; detected $arch"
    [[ -d /sys/firmware/efi ]] || fatal "Alpha verified path requires UEFI firmware."
}

confirm_exact() {
    local expected="$1"
    local prompt="$2"
    local answer
    printf '%s\n' "$prompt"
    read -r answer
    [[ "$answer" == "$expected" ]]
}

human_bytes() {
    local bytes="$1"
    if command -v numfmt >/dev/null 2>&1; then
        numfmt --to=iec-i --suffix=B "$bytes"
    else
        printf '%s bytes\n' "$bytes"
    fi
}

sha256_file() {
    sha256sum "$1" | awk '{print $1}'
}

safe_mkdir() {
    install -d -m "${2:-0755}" "$1"
}

is_sha256_digest() {
    [[ "$1" =~ ^sha256:[0-9a-fA-F]{64}$ ]]
}

is_sha256_hex() {
    [[ "$1" =~ ^[0-9a-fA-F]{64}$ ]]
}

path_is_under() {
    local child parent
    child="$(readlink -m "$1")"
    parent="$(readlink -m "$2")"
    [[ "$child" == "$parent" || "$child" == "$parent"/* ]]
}
