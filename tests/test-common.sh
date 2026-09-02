#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
HSI_VERSION=test
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"

is_sha256_digest "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
! is_sha256_digest "sha256:not-a-digest"
is_sha256_hex "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
path_is_under /tmp/a/b /tmp/a
! path_is_under /tmp/ab /tmp/a

printf 'PASS common helper tests\n'
