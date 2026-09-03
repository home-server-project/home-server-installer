#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"
# shellcheck source=../lib/bootc-install.sh
source "$ROOT/lib/bootc-install.sh"

TEST_ROOT="$(mktemp -d)"
trap 'rm -rf "$TEST_ROOT"' EXIT

HASH='312cf33a8023d80ef4509608026ebda3129604989e81d111fb947263143461e9'
HSI_TARGET_MOUNT="$TEST_ROOT/target"
REAL_DEPLOY="$HSI_TARGET_MOUNT/ostree/deploy/fedora-coreos/deploy/$HASH.0"
BACKING_DEPLOY="$HSI_TARGET_MOUNT/ostree/deploy/fedora-coreos/backing/$HASH.0"

mkdir -p "$REAL_DEPLOY/usr/lib64/gdk-pixbuf-2.0/2.10.0"
mkdir -p "$REAL_DEPLOY/usr/lib/firmware/ath10k/QCA4019/hw1.0"
mkdir -p "$BACKING_DEPLOY"

actual="$(find_deployment_root)"
[[ "$actual" == "$REAL_DEPLOY" ]] || {
    printf 'FAIL deployment root mismatch: %s\n' "$actual" >&2
    exit 1
}

SECOND_HASH='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
mkdir -p "$HSI_TARGET_MOUNT/ostree/deploy/fedora-coreos/deploy/$SECOND_HASH.0"

if find_deployment_root >/dev/null 2>&1; then
    printf 'FAIL multiple deployment roots were accepted\n' >&2
    exit 1
fi

printf 'PASS bootc deployment root discovery tests\n'
