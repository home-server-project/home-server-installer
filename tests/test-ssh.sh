#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
HSI_VERSION=test
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"
# shellcheck source=../lib/ssh.sh
source "$ROOT/lib/ssh.sh"

validate_ssh_public_key "$ROOT/tests/fixtures/authorized_keys.pub"

if (validate_ssh_public_key "$ROOT/tests/fixtures/not-a-public-key") >/dev/null 2>&1; then
    printf 'FAIL private key accepted\n' >&2
    exit 1
fi

printf 'PASS SSH public/private key validation\n'
