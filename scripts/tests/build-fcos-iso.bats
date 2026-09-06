#!/usr/bin/env bats
# Tests for scripts/build-fcos-iso.sh argument validation and packaging contract.
# Requires: bats-core (https://github.com/bats-core/bats-core)
#
# Run: bats scripts/tests/build-fcos-iso.bats

SCRIPT="$BATS_TEST_DIRNAME/../build-fcos-iso.sh"

run_expect_fail() {
  run bash "$SCRIPT" "$@"
  [ "$status" -ne 0 ]
}

@test "unknown argument exits with error" {
  run_expect_fail --bogus-flag
  [[ "$output" == *"Unknown argument"* ]]
}

@test "invalid --arch value exits with error" {
  run_expect_fail --arch riscv64
  [[ "$output" == *"must be amd64 or arm64"* ]]
}

@test "invalid --stream value exits with error" {
  run_expect_fail --stream nightly
  [[ "$output" == *"must be stable, testing, or next"* ]]
}

@test "valid stream names are accepted" {
  for s in stable testing next; do
    run bash "$SCRIPT" --stream "$s" --arch amd64 2>&1 || true
    [[ "$output" != *"must be stable, testing, or next"* ]]
  done
}

@test "--stream=value (equals form) is parsed" {
  run bash "$SCRIPT" --stream=nightly 2>&1
  [ "$status" -ne 0 ]
  [[ "$output" == *"must be stable, testing, or next"* ]]
}

@test "--arch=value (equals form) is parsed" {
  run bash "$SCRIPT" --arch=sparc 2>&1
  [ "$status" -ne 0 ]
  [[ "$output" == *"must be amd64 or arm64"* ]]
}

@test "bare stream name (positional) is parsed" {
  run bash "$SCRIPT" testing --arch=riscv64 2>&1
  [ "$status" -ne 0 ]
  [[ "$output" == *"must be amd64 or arm64"* ]]
}

@test "--binary value (space-separated) does not trigger unknown-argument error" {
  run bash "$SCRIPT" --binary /tmp/fake-knuckle 2>&1 || true
  [[ "$output" != *"Unknown argument"* ]]
}

@test "--binary=value (equals form) does not trigger unknown-argument error" {
  run bash "$SCRIPT" --binary=/tmp/fake-knuckle 2>&1 || true
  [[ "$output" != *"Unknown argument"* ]]
}

@test "--binary combined with valid flags does not trigger arg-parse error" {
  run bash "$SCRIPT" --binary /tmp/fake-knuckle --stream stable --arch amd64 2>&1 || true
  [[ "$output" != *"Unknown argument"* ]]
  [[ "$output" != *"must be amd64 or arm64"* ]]
  [[ "$output" != *"must be stable, testing, or next"* ]]
}

@test "FCOS builder packages Knuckle outside live Ignition" {
  run grep -F 'KNUCKLE_INITRD=' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F '90-home-server-installer.sh' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F 'initrd  /knuckle.img' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F 'initrd  /ignition.img' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F 'base64.b64encode' "$SCRIPT"
  [ "$status" -ne 0 ]

  run grep -F 'data:;base64' "$SCRIPT"
  [ "$status" -ne 0 ]
}

@test "FCOS builder verifies payload and keeps live Ignition bounded" {
  run grep -F 'cmp -s "$BINARY" "$VERIFY_DIR/opt/home-server-installer/knuckle"' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F 'if (( IGN_SIZE > 65536 )); then' "$SCRIPT"
  [ "$status" -eq 0 ]

  run grep -F 'sha256sum /opt/knuckle' "$SCRIPT"
  [ "$status" -eq 0 ]
}

@test "--ssh-key value (space-separated) does not trigger unknown-argument error" {
  run bash "$SCRIPT" --ssh-key "ssh-ed25519 AAAA..." 2>&1 || true
  [[ "$output" != *"Unknown argument"* ]]
}

@test "--ssh-key=value (equals form) does not trigger unknown-argument error" {
  run bash "$SCRIPT" --ssh-key="ssh-ed25519 AAAA..." 2>&1 || true
  [[ "$output" != *"Unknown argument"* ]]
}

@test "missing coreos-installer prints dependency error" {
  run env PATH=/usr/bin:/bin bash "$SCRIPT" --stream stable --arch amd64 2>&1
  if [[ "$output" == *"required command not found: coreos-installer"* ]]; then
    [ "$status" -ne 0 ]
  else
    skip "coreos-installer is available in PATH"
  fi
}
