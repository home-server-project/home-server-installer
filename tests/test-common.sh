#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
HSI_VERSION="test"
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"

is_sha256_digest "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if is_sha256_digest "sha256:not-a-digest"; then
    printf 'FAIL invalid sha256 digest accepted\n' >&2
    exit 1
fi
is_sha256_hex "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
path_is_under /tmp/a/b /tmp/a
if path_is_under /tmp/ab /tmp/a; then
    printf 'FAIL sibling path accepted as child path\n' >&2
    exit 1
fi

printf 'PASS common helper tests\n'
