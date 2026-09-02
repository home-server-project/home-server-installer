#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
HSI_VERSION="test"
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"
# shellcheck source=../lib/metadata.sh
source "$ROOT/lib/metadata.sh"

rm -f /tmp/HOME_SERVER_INSTALLER_SHOULD_NOT_EXIST
load_metadata_file "$ROOT/tests/fixtures/install.env.valid"
validate_metadata

[[ "$IMAGE_REF" == "ghcr.io/home-server-project/home-server-ucore-hci:lts" ]]
[[ "$PARTITION_POLICY" == "home-server-default-v1" ]]
[[ "$UPDATE_POLICY" == "home-server-stage-v1" ]]

unset IMAGE_REF IMAGE_DIGEST OCI_ARCHIVE OCI_ARCHIVE_SHA256 SSH_PUBLIC_KEY PARTITION_POLICY UPDATE_POLICY INSTALLER_FORMAT_VERSION || true
load_metadata_file "$ROOT/tests/fixtures/install.env.injection"
[[ ! -e /tmp/HOME_SERVER_INSTALLER_SHOULD_NOT_EXIST ]]
# shellcheck disable=SC2016 -- literal command substitution is the injection-test payload.
[[ "$IMAGE_REF" == '$(touch /tmp/HOME_SERVER_INSTALLER_SHOULD_NOT_EXIST)' ]]
if (validate_metadata) >/dev/null 2>&1; then
    printf 'FAIL unsafe IMAGE_REF accepted by metadata validation\n' >&2
    exit 1
fi

printf 'PASS metadata parser does not evaluate shell input and rejects unsafe refs\n'
