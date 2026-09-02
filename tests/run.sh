#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

for test_file in "$ROOT"/tests/test-*.sh; do
    printf '==> %s\n' "${test_file##*/}"
    bash "$test_file"
done

printf '\nAll non-destructive unit tests passed.\n'
