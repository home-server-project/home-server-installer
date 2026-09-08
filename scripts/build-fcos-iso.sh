#!/usr/bin/env bash
# Modified by Home Server Project from Project Bluefin Knuckle.
# Build a self-contained UEFI Fedora CoreOS live ISO containing Knuckle.
#
# The stock FCOS ISO reserves only 256 KiB for embedded live-Ignition data, so
# the Knuckle executable must not be stored there. This builder instead:
#   1. downloads + verifies the official FCOS live ISO with coreos-installer
#   2. extracts the official PXE kernel/initramfs/rootfs from that verified ISO
#   3. creates a separate initrd containing Knuckle plus a dracut pre-pivot hook
#   4. keeps live Ignition small (bootstrap service + optional SSH key only)
#   5. concatenates FCOS initramfs + rootfs + Knuckle + Ignition into one initrd
#   6. assembles a UEFI-only systemd-boot ISO that loads that single initrd
#
# The pre-pivot hook copies Knuckle into the writable live root. The small
# Ignition-created bootstrap service restores its SELinux label, verifies its
# SHA256, and then launches the TUI on tty1.
#
# Requirements: coreos-installer, python3, cpio, gzip, xorriso, mtools,
#               and a systemd-boot EFI binary.
#
# Usage: ./scripts/build-fcos-iso.sh [--stream stable|testing|next] [--arch amd64|arm64] [--binary /path/to/knuckle] [--ssh-key "ssh-ed25519 ..."]
set -euo pipefail

# ── Argument parsing ─────────────────────────────────────────────────────────
STREAM="stable"
ARCH="amd64"
BINARY_OVERRIDE=""
SSH_PUB_KEY=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --stream=*) STREAM="${1#--stream=}"; shift ;;
        --stream)   STREAM="$2"; shift 2 ;;
        --arch=*)   ARCH="${1#--arch=}"; shift ;;
        --arch)     ARCH="$2"; shift 2 ;;
        --binary=*) BINARY_OVERRIDE="${1#--binary=}"; shift ;;
        --binary)   BINARY_OVERRIDE="$2"; shift 2 ;;
        --ssh-key=*) SSH_PUB_KEY="${1#--ssh-key=}"; shift ;;
        --ssh-key)  SSH_PUB_KEY="$2"; shift 2 ;;
        stable|testing|next) STREAM="$1"; shift ;;
        *) echo "Unknown argument: $1" >&2; exit 1 ;;
    esac
done

case "$STREAM" in
    stable|testing|next) ;;
    *) echo "error: --stream must be stable, testing, or next (got '$STREAM')" >&2; exit 1 ;;
esac
case "$ARCH" in
    amd64|arm64) ;;
    *) echo "error: --arch must be amd64 or arm64 (got '$ARCH')" >&2; exit 1 ;;
esac

case "$ARCH" in
    amd64)
        COREOS_ARCH="x86_64"
        SDBOOT_FILENAME="systemd-bootx64.efi"
        EFI_BOOT_NAME="BOOTX64.EFI"
        ;;
    arm64)
        COREOS_ARCH="aarch64"
        SDBOOT_FILENAME="systemd-bootaa64.efi"
        EFI_BOOT_NAME="BOOTAA64.EFI"
        ;;
esac

# ── Dependency checks ────────────────────────────────────────────────────────
for cmd in coreos-installer python3 cpio gzip xorriso mformat mcopy mmd mtype sha256sum; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "error: required command not found: $cmd" >&2
        echo "  Fedora: install coreos-installer xorriso mtools cpio systemd-boot-unsigned" >&2
        echo "  Ubuntu: install xorriso mtools cpio systemd-boot-efi and coreos-installer" >&2
        exit 1
    fi
done

SDBOOT_EFI=""
for candidate in \
    "/usr/lib/systemd/boot/efi/${SDBOOT_FILENAME}" \
    "/lib/systemd/boot/efi/${SDBOOT_FILENAME}" \
    "/usr/share/systemd/boot/efi/${SDBOOT_FILENAME}"; do
    if [[ -f "$candidate" ]]; then
        SDBOOT_EFI="$candidate"
        break
    fi
done
if [[ -z "$SDBOOT_EFI" ]]; then
    echo "error: ${SDBOOT_FILENAME} not found" >&2
    echo "  Fedora: install systemd-boot-unsigned" >&2
    echo "  Ubuntu: install systemd-boot-efi" >&2
    exit 1
fi

# ── Paths ────────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$ROOT_DIR/.fcos-iso-build"
OUTPUT_DIR="$ROOT_DIR/output"
ISO_CACHE_DIR="$BUILD_DIR/iso-cache-${STREAM}-${ARCH}"
PXE_DIR="$BUILD_DIR/pxe-${STREAM}-${ARCH}"

mkdir -p "$BUILD_DIR" "$OUTPUT_DIR" "$ISO_CACHE_DIR"

BINARY=""
if [[ -n "$BINARY_OVERRIDE" ]]; then
    BINARY="$BINARY_OVERRIDE"
elif [[ -f "$ROOT_DIR/bin/knuckle-${ARCH}" ]]; then
    BINARY="$ROOT_DIR/bin/knuckle-${ARCH}"
elif [[ -f "$ROOT_DIR/bin/knuckle" ]]; then
    BINARY="$ROOT_DIR/bin/knuckle"
elif [[ -f "$ROOT_DIR/knuckle" ]]; then
    BINARY="$ROOT_DIR/knuckle"
fi

echo "=== Building Home Server FCOS installer ISO (stream: $STREAM, arch: $ARCH) ==="
echo "  systemd-boot: $SDBOOT_EFI"

# ── 1. Build Knuckle ─────────────────────────────────────────────────────────
if [[ ! -f "$BINARY" ]]; then
    echo "[1/7] Building Knuckle ($ARCH)..."
    VERSION="$(git -C "$ROOT_DIR" describe --tags --always 2>/dev/null || echo dev)"
    (cd "$ROOT_DIR" && GOOS=linux GOARCH="$ARCH" CGO_ENABLED=0 \
        go build -buildvcs=false -ldflags="-s -w -X main.version=${VERSION}" \
        -o "bin/knuckle-${ARCH}" ./cmd/knuckle)
    BINARY="$ROOT_DIR/bin/knuckle-${ARCH}"
else
    echo "[1/7] Using existing Knuckle binary: $BINARY"
fi
BINARY_SHA256="$(sha256sum "$BINARY" | awk '{print $1}')"
echo "  binary : $(du -h "$BINARY" | cut -f1)"
echo "  sha256 : $BINARY_SHA256"

# ── 2. Download the official FCOS ISO ────────────────────────────────────────
echo "[2/7] Preparing verified FCOS live ISO..."
EXISTING_ISO="$(find "$ISO_CACHE_DIR" -maxdepth 1 -type f -name '*.iso' -print -quit)"
if [[ -n "$EXISTING_ISO" ]]; then
    LIVE_ISO="$EXISTING_ISO"
    echo "  Using cached FCOS ISO: $(basename "$LIVE_ISO")"
else
    coreos-installer download \
        --stream "$STREAM" \
        --platform metal \
        --format iso \
        --architecture "$COREOS_ARCH" \
        --directory "$ISO_CACHE_DIR"
    LIVE_ISO="$(find "$ISO_CACHE_DIR" -maxdepth 1 -type f -name '*.iso' -print -quit)"
    [[ -n "$LIVE_ISO" ]] || { echo "error: FCOS ISO download produced no ISO" >&2; exit 1; }
    echo "  Downloaded: $(basename "$LIVE_ISO") ($(du -h "$LIVE_ISO" | cut -f1))"
fi

# ── 3. Extract official PXE artifacts from that ISO ──────────────────────────
echo "[3/7] Extracting FCOS kernel/initramfs/rootfs..."
ISO_BASENAME="$(basename "$LIVE_ISO")"
EXTRACT_MARKER="$PXE_DIR/.source-iso"
if [[ ! -f "$EXTRACT_MARKER" ]] || [[ "$(cat "$EXTRACT_MARKER" 2>/dev/null || true)" != "$ISO_BASENAME" ]]; then
    rm -rf "$PXE_DIR"
    mkdir -p "$PXE_DIR"
    coreos-installer iso extract pxe --output-dir "$PXE_DIR" "$LIVE_ISO"
    printf '%s\n' "$ISO_BASENAME" > "$EXTRACT_MARKER"
else
    echo "  Using cached PXE artifacts from $ISO_BASENAME"
fi

# `iso extract pxe` prefixes the ISO stem to the lower-cased IMAGES/PXEBOOT
# filenames, e.g. ...-vmlinuz, ...-initrd.img, and ...-rootfs.img.
KERNEL="$(find "$PXE_DIR" -maxdepth 1 -type f -name '*-vmlinuz' -print -quit)"
FCOS_INITRAMFS="$(find "$PXE_DIR" -maxdepth 1 -type f -name '*-initrd.img' -print -quit)"
FCOS_ROOTFS="$(find "$PXE_DIR" -maxdepth 1 -type f -name '*-rootfs.img' -print -quit)"

for artifact in "$KERNEL" "$FCOS_INITRAMFS" "$FCOS_ROOTFS"; do
    [[ -n "$artifact" && -f "$artifact" ]] || { echo "error: FCOS PXE artifact missing after extraction" >&2; exit 1; }
done

echo "  kernel    : $(du -h "$KERNEL" | cut -f1)"
echo "  initramfs : $(du -h "$FCOS_INITRAMFS" | cut -f1)"
echo "  rootfs    : $(du -h "$FCOS_ROOTFS" | cut -f1)"

# ── 4. Build the separate Knuckle initrd ─────────────────────────────────────
# The executable lives in its own initrd, outside the tiny ISO Ignition embed
# area. A dracut pre-pivot hook copies it into the writable live root before
# switch_root. SELinux labeling is finalized by the Ignition-created bootstrap
# service after the pivot and before execution.
echo "[4/7] Building separate Knuckle initrd payload..."
KNUCKLE_ROOT="$BUILD_DIR/knuckle-initrd-root"
KNUCKLE_INITRD="$BUILD_DIR/knuckle-${BINARY_SHA256:0:16}.img"
rm -rf "$KNUCKLE_ROOT"
mkdir -p \
    "$KNUCKLE_ROOT/opt/home-server-installer" \
    "$KNUCKLE_ROOT/usr/lib/dracut/hooks/pre-pivot"

cp "$BINARY" "$KNUCKLE_ROOT/opt/home-server-installer/knuckle"
chmod 0755 "$KNUCKLE_ROOT/opt/home-server-installer/knuckle"

cat > "$KNUCKLE_ROOT/usr/lib/dracut/hooks/pre-pivot/90-home-server-installer.sh" <<'HOOK'
#!/bin/sh
# This file is sourced by dracut. Use return, not exit.
_home_server_root="${NEWROOT:-/sysroot}"
_home_server_src="/opt/home-server-installer/knuckle"
_home_server_dst="${_home_server_root}/opt/knuckle"

if [ ! -f "${_home_server_src}" ]; then
    echo "Home Server Installer: Knuckle payload missing from initramfs" >&2
    unset _home_server_root _home_server_src _home_server_dst
    return 1
fi
if [ ! -d "${_home_server_root}" ]; then
    echo "Home Server Installer: live root is not mounted at ${_home_server_root}" >&2
    unset _home_server_root _home_server_src _home_server_dst
    return 1
fi

mkdir -p "${_home_server_root}/opt"
cp "${_home_server_src}" "${_home_server_dst}"
chmod 0755 "${_home_server_dst}"
echo "Home Server Installer: Knuckle payload copied into live root"

unset _home_server_root _home_server_src _home_server_dst
return 0
HOOK
chmod 0755 "$KNUCKLE_ROOT/usr/lib/dracut/hooks/pre-pivot/90-home-server-installer.sh"

rm -f "$KNUCKLE_INITRD"
(
    cd "$KNUCKLE_ROOT"
    find . -print0 | cpio --null -o -H newc --quiet --owner=0:0 | gzip -9 > "$KNUCKLE_INITRD"
)
echo "  payload initrd: $(du -h "$KNUCKLE_INITRD" | cut -f1)"

# Prove that the compressed payload contains the exact binary we were given.
VERIFY_DIR="$(mktemp -d)"
trap 'rm -rf "$VERIFY_DIR"' EXIT
gzip -dc "$KNUCKLE_INITRD" | (cd "$VERIFY_DIR" && cpio -id --quiet 'opt/home-server-installer/knuckle')
cmp -s "$BINARY" "$VERIFY_DIR/opt/home-server-installer/knuckle" || {
    echo "error: Knuckle initrd payload does not match input binary" >&2
    exit 1
}
rm -rf "$VERIFY_DIR"
trap - EXIT

# ── 5. Generate a SMALL live Ignition bootstrap ──────────────────────────────
# Ignition carries only configuration: no executable payload. It creates the
# launcher with the correct FCOS SELinux labeling, enables sshd, and optionally
# adds the supplied SSH public key.
echo "[5/7] Generating small live Ignition bootstrap..."
IGN_FILE="$BUILD_DIR/live-bootstrap.ign"
IGN_INITRD="$BUILD_DIR/live-bootstrap.img"

python3 - "$IGN_FILE" "$BINARY_SHA256" "$SSH_PUB_KEY" <<'PYEOF'
import base64
import json
import sys

ign_file = sys.argv[1]
expected_sha256 = sys.argv[2]
ssh_key = sys.argv[3].strip()

def data_source(text):
    return (
        "data:text/plain;charset=utf-8;base64,"
        + base64.b64encode(text.encode("utf-8")).decode("ascii")
    )

bootstrap = f"""#!/usr/bin/bash
set -euo pipefail
expected={expected_sha256}
actual=$(sha256sum /opt/knuckle | awk '{{print $1}}')
if [[ \"$actual\" != \"$expected\" ]]; then
    echo \"Home Server Installer: Knuckle SHA256 mismatch\" >&2
    exit 1
fi
restorecon -F /opt/knuckle
exec /opt/knuckle
"""
bootstrap_source = data_source(bootstrap)

unit = """\
[Unit]
Description=Home Server Installer
After=network-online.target
Wants=network-online.target
Conflicts=getty@tty1.service
ConditionPathExists=/opt/knuckle

[Service]
Type=idle
ExecStart=/opt/home-server-installer-bootstrap
StandardInput=tty-force
StandardOutput=tty
StandardError=journal+console
TTYPath=/dev/tty1
TTYReset=yes
TTYVHangup=yes
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target"""

config = {
    "ignition": {"version": "3.3.0"},
    "storage": {
        "files": [
            {
                "path": "/opt/home-server-installer-bootstrap",
                "mode": 0o755,
                "contents": {"source": bootstrap_source},
            }
        ]
    },
    "systemd": {
        "units": [
            {"name": "sshd.service", "enabled": True},
            {"name": "home-server-installer.service", "enabled": True, "contents": unit},
        ]
    },
}
if ssh_key:
    config["passwd"] = {"users": [{"name": "core", "sshAuthorizedKeys": [ssh_key]}]}
    config["storage"]["files"].append(
        {
            "path": "/opt/home-server-installer-ssh.pub",
            "mode": 0o600,
            "contents": {"source": data_source(ssh_key + "\n")},
        }
    )

with open(ign_file, "w", encoding="utf-8") as f:
    json.dump(config, f, separators=(",", ":"))
    f.write("\n")
PYEOF

IGN_SIZE="$(stat -c%s "$IGN_FILE")"
if (( IGN_SIZE > 65536 )); then
    echo "error: live Ignition unexpectedly large: ${IGN_SIZE} bytes" >&2
    exit 1
fi
rm -f "$IGN_INITRD"
coreos-installer pxe ignition wrap -i "$IGN_FILE" -o "$IGN_INITRD"
echo "  Ignition JSON : ${IGN_SIZE} bytes"
echo "  Ignition initrd: $(du -h "$IGN_INITRD" | cut -f1)"

# FCOS supports an initramfs+rootfs combined initrd. Keep the Home Server
# additions after the stock FCOS payload, preserving this exact order:
#   FCOS initramfs -> FCOS rootfs -> Knuckle -> Ignition
COMBINED_INITRD="$BUILD_DIR/home-server-initrd-${BINARY_SHA256:0:16}.img"
cat "$FCOS_INITRAMFS" "$FCOS_ROOTFS" "$KNUCKLE_INITRD" "$IGN_INITRD" > "$COMBINED_INITRD"

EXPECTED_COMBINED_SIZE=$(( \
    $(stat -c%s "$FCOS_INITRAMFS") + \
    $(stat -c%s "$FCOS_ROOTFS") + \
    $(stat -c%s "$KNUCKLE_INITRD") + \
    $(stat -c%s "$IGN_INITRD") ))
ACTUAL_COMBINED_SIZE="$(stat -c%s "$COMBINED_INITRD")"
if (( ACTUAL_COMBINED_SIZE != EXPECTED_COMBINED_SIZE )); then
    echo "error: combined initrd size mismatch: ${ACTUAL_COMBINED_SIZE} != ${EXPECTED_COMBINED_SIZE}" >&2
    exit 1
fi
echo "  combined initrd: $(du -h "$COMBINED_INITRD" | cut -f1)"

# ── 6. Build a UEFI ESP that loads one combined FCOS initrd ──────────────────
echo "[6/7] Building UEFI System Partition..."
ISO_DIR="$BUILD_DIR/iso-root"
EFI_IMG="$BUILD_DIR/efi.img"
rm -rf "$ISO_DIR"
mkdir -p "$ISO_DIR"

TOTAL_BYTES=$(( \
    $(stat -c%s "$KERNEL") + \
    $(stat -c%s "$COMBINED_INITRD") + \
    64 * 1024 * 1024 ))
ESP_SIZE_MB=$(( (TOTAL_BYTES + 1024 * 1024 - 1) / 1024 / 1024 ))

dd if=/dev/zero of="$EFI_IMG" bs=1M count="$ESP_SIZE_MB" status=none
mformat -i "$EFI_IMG" -F ::
mmd -i "$EFI_IMG" ::/EFI ::/EFI/BOOT ::/loader ::/loader/entries
mcopy -i "$EFI_IMG" "$SDBOOT_EFI" "::/EFI/BOOT/${EFI_BOOT_NAME}"

printf 'default home-server-installer\ntimeout 5\neditor no\n' \
    | mcopy -i "$EFI_IMG" - ::/loader/loader.conf

# Primary interactive entry: tty0 is last, so the TUI is attached to the VGA
# console. Kernel/systemd diagnostics are still mirrored to ttyS0.
cat > "$BUILD_DIR/home-server-installer.conf" <<'ENTRY'
title   Home Server Installer - Fedora CoreOS
linux   /vmlinuz
initrd  /home-server-initrd.img
options ignition.firstboot ignition.platform.id=metal console=ttyS0,115200n8 console=tty0
ENTRY
mcopy -i "$EFI_IMG" "$BUILD_DIR/home-server-installer.conf" ::/loader/entries/home-server-installer.conf

# Alternate serial entry for headless diagnostics.
cat > "$BUILD_DIR/home-server-installer-serial.conf" <<'ENTRY'
title   Home Server Installer - Fedora CoreOS (serial)
linux   /vmlinuz
initrd  /home-server-initrd.img
options ignition.firstboot ignition.platform.id=metal console=ttyS0,115200n8
ENTRY
mcopy -i "$EFI_IMG" "$BUILD_DIR/home-server-installer-serial.conf" ::/loader/entries/home-server-installer-serial.conf

mcopy -i "$EFI_IMG" "$KERNEL"          ::/vmlinuz
mcopy -i "$EFI_IMG" "$COMBINED_INITRD" ::/home-server-initrd.img

# Verify the UEFI boot entry has exactly one initrd and points at the combined
# payload, avoiding the multi-initrd boot path that failed in the VM.
ENTRY_TEXT="$(mtype -i "$EFI_IMG" ::/loader/entries/home-server-installer.conf)"
INITRD_COUNT="$(grep -c '^initrd[[:space:]]' <<<"$ENTRY_TEXT")"
if [[ "$INITRD_COUNT" -ne 1 ]]; then
    echo "error: UEFI loader entry must contain exactly one initrd line (got $INITRD_COUNT)" >&2
    exit 1
fi
grep -Fq 'initrd  /home-server-initrd.img' <<<"$ENTRY_TEXT" || {
    echo "error: UEFI loader entry missing combined initrd" >&2
    exit 1
}

echo "  ESP: $(du -h "$EFI_IMG" | cut -f1)"

# ── 7. Assemble the final UEFI ISO ───────────────────────────────────────────
echo "[7/7] Assembling self-contained FCOS installer ISO..."
cp "$EFI_IMG" "$ISO_DIR/efi.img"
ISO_OUT="$OUTPUT_DIR/knuckle-fcos-installer-${STREAM}-${ARCH}.iso"
rm -f "$ISO_OUT"

xorriso -as mkisofs \
    -o "$ISO_OUT" \
    -R -J -joliet-long \
    -V "HOME_SERVER_INSTALL" \
    -eltorito-alt-boot \
    -e efi.img \
    -no-emul-boot \
    --efi-boot-part --efi-boot-image \
    "$ISO_DIR" >/dev/null 2>&1

echo ""
echo "ISO built: $ISO_OUT ($(du -h "$ISO_OUT" | cut -f1))"
echo "  FCOS source : $ISO_BASENAME"
echo "  Knuckle SHA : $BINARY_SHA256"
echo "  live Ignition: ${IGN_SIZE} bytes (binary is NOT embedded there)"
echo "  boot initrd : one combined FCOS+Knuckle+Ignition payload"
echo ""
if [[ "$ARCH" == "arm64" ]]; then
    echo "Test with QEMU (UEFI, arm64):"
    echo "  OVMF=/usr/share/AAVMF/AAVMF_CODE.fd"
    echo "  qemu-system-aarch64 -m 8192 -M virt -cpu cortex-a57 \\"
else
    echo "Test with QEMU (UEFI, amd64):"
    echo "  OVMF=/usr/share/OVMF/OVMF_CODE.fd"
    echo "  qemu-system-x86_64 -m 8192 -enable-kvm \\"
fi
echo "    -drive if=pflash,format=raw,readonly=on,file=\$OVMF \\"
echo "    -cdrom $ISO_OUT \\"
echo "    -drive if=virtio,file=target.qcow2,format=qcow2"
