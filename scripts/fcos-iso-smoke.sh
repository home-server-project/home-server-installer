#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage: fcos-iso-smoke.sh <iso-path> <ovmf-path> [timeout-seconds]

Boot a Home Server FCOS installer ISO headlessly with QEMU and verify that:
- UEFI/systemd-boot reaches the FCOS live environment
- the separate Knuckle initrd pre-pivot hook copies the payload into the live root
- home-server-installer.service is started after pivot
- no payload/hash bootstrap failure is logged
EOF
}

ISO_PATH=${1:-}
OVMF_PATH=${2:-}
TIMEOUT_SECONDS=${3:-180}
QEMU_BIN=${QEMU_BIN:-qemu-system-x86_64}
TARGET_DISK=.vm/fcos-iso-smoke-target.qcow2
SERIAL_LOG=.vm/fcos-iso-smoke-serial.log
QEMU_PID=

if [[ -z "$ISO_PATH" || -z "$OVMF_PATH" ]]; then
  usage >&2
  exit 1
fi
[[ -f "$ISO_PATH" ]] || { echo "ISO not found: $ISO_PATH" >&2; exit 1; }
[[ -f "$OVMF_PATH" ]] || { echo "OVMF not found: $OVMF_PATH" >&2; exit 1; }
[[ "$TIMEOUT_SECONDS" =~ ^[0-9]+$ ]] || { echo "Timeout must be an integer number of seconds: $TIMEOUT_SECONDS" >&2; exit 1; }

cleanup() {
  if [[ -n "$QEMU_PID" ]] && kill -0 "$QEMU_PID" 2>/dev/null; then
    kill "$QEMU_PID" 2>/dev/null || true
    wait "$QEMU_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

log_has() {
  grep -aEq -- "$1" "$SERIAL_LOG"
}

mkdir -p .vm
rm -f "$TARGET_DISK" "$SERIAL_LOG"
qemu-img create -f qcow2 "$TARGET_DISK" 20G >/dev/null

printf '=== fcos-iso-smoke: headless UEFI ISO boot ===\n'
printf 'ISO: %s\n' "$ISO_PATH"
printf 'OVMF: %s\n' "$OVMF_PATH"
printf 'Log: %s\n\n' "$SERIAL_LOG"

"$QEMU_BIN" \
  -machine q35 \
  -m 4096 \
  -smp 2 \
  -cpu host \
  -enable-kvm \
  -drive if=pflash,format=raw,readonly=on,file="$OVMF_PATH" \
  -cdrom "$ISO_PATH" \
  -drive if=virtio,file="$TARGET_DISK",format=qcow2 \
  -nographic \
  >"$SERIAL_LOG" 2>&1 &
QEMU_PID=$!

hook_seen=0
service_seen=0

deadline=$((SECONDS + TIMEOUT_SECONDS))
while (( SECONDS < deadline )); do
  if [[ -f "$SERIAL_LOG" ]]; then
    if log_has 'Knuckle payload missing from initramfs|live root is not mounted|Knuckle SHA256 mismatch'; then
      echo 'FAIL: FCOS installer bootstrap reported a payload failure'
      tail -80 "$SERIAL_LOG" || true
      exit 1
    fi

    if (( hook_seen == 0 )) && log_has 'Home Server Installer: Knuckle payload copied into live root'; then
      echo '  OK: Knuckle initrd pre-pivot hook executed'
      hook_seen=1
    fi

    if (( service_seen == 0 )) && log_has 'Started .*Home Server Installer|Started home-server-installer\.service'; then
      echo '  OK: home-server-installer.service started'
      service_seen=1
    fi

    if (( hook_seen == 1 && service_seen == 1 )); then
      echo
      echo 'PASS: FCOS installer ISO boot smoke'
      exit 0
    fi
  fi

  if ! kill -0 "$QEMU_PID" 2>/dev/null; then
    wait "$QEMU_PID" 2>/dev/null || true
    break
  fi
  sleep 1
done

echo
echo "FAIL: FCOS ISO smoke did not observe all required boot markers"
echo "  hook_seen=$hook_seen service_seen=$service_seen"
echo '--- serial log tail ---'
tail -100 "$SERIAL_LOG" || true
exit 1
