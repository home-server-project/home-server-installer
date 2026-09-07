package install

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/projectbluefin/knuckle/internal/model"
	"github.com/projectbluefin/knuckle/internal/runner"
)

const homeServerInstallerCosignPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIz0DAQcDQgAEWU3SeANKBm2Dql6FGZYNu2Bd7nZf
wSS/hmdv0B25JOSqi0dbyvW8XAJHJ4UOl/GeOSQM4XuDey9yI9I09r9XWw==
-----END PUBLIC KEY-----
`

const homeServerDirectInstallScript = `set -euo pipefail

TARGET="$1"
STABLE_PATH="$2"
EXPECTED_SIZE="$3"
EXPECTED_SERIAL="$4"
IMAGE="$5"
BOOT_MIB="$6"
USERNAME="$7"
PASSWORD_HASH="$8"
HOSTNAME="$9"
TIMEZONE="${10}"
NETWORK_MODE="${11}"
NETWORK_IFACE="${12}"
NETWORK_ADDR="${13}"
NETWORK_GATEWAY="${14}"
NETWORK_DNS="${15}"
SSH_KEYS_FILE="${16}"
POLICY_FILE="${17}"
KEY_FILE="${18}"

TARGET_ROOT=/var/mnt/home-server-target
PODMAN_SCRATCH="${TARGET_ROOT}/.installer-podman"
IMAGE_TMP="${TARGET_ROOT}/.installer-tmp"
PODMAN_MOUNTED=0
ROOT_MOUNTED=0
BOOT_MOUNTED=0
EFI_MOUNTED=0

cleanup_mounts() {
    set +e
    if (( EFI_MOUNTED )); then umount "${TARGET_ROOT}/boot/efi"; fi
    if (( BOOT_MOUNTED )); then umount "${TARGET_ROOT}/boot"; fi
    if (( PODMAN_MOUNTED )); then umount /var/lib/containers; fi
    mountpoint -q "${PODMAN_SCRATCH}" && umount "${PODMAN_SCRATCH}"
    mountpoint -q "${IMAGE_TMP}" && umount "${IMAGE_TMP}"
    if (( ROOT_MOUNTED )); then umount "${TARGET_ROOT}"; fi
}
trap cleanup_mounts EXIT

if [[ $EUID -ne 0 ]]; then
    echo "Home Server direct installation must run as root" >&2
    exit 1
fi

case "$BOOT_MIB" in
    1024|2048) ;;
    *) echo "unsupported /boot size: ${BOOT_MIB} MiB" >&2; exit 1 ;;
esac

for cmd in readlink lsblk wipefs sgdisk blockdev udevadm mkfs.vfat mkfs.ext4 mkfs.xfs mount umount mountpoint blkid podman bootc systemctl find install; do
    command -v "$cmd" >/dev/null 2>&1 || {
        echo "required installer command not found: $cmd" >&2
        exit 1
    }
done

# Revalidate the destructive target immediately before the first write.
TARGET_REAL="$(readlink -f "$TARGET")"
[[ -b "$TARGET_REAL" ]] || { echo "target is not a block device: $TARGET" >&2; exit 1; }
if [[ -n "$STABLE_PATH" ]]; then
    STABLE_REAL="$(readlink -f "$STABLE_PATH")"
    [[ "$STABLE_REAL" == "$TARGET_REAL" ]] || {
        echo "stable disk identity no longer resolves to selected target" >&2
        exit 1
    }
fi

ACTUAL_SIZE="$(lsblk -dnbo SIZE "$TARGET_REAL" | tr -d '[:space:]')"
if [[ -n "$EXPECTED_SIZE" && "$EXPECTED_SIZE" != "0" && "$ACTUAL_SIZE" != "$EXPECTED_SIZE" ]]; then
    echo "target disk size changed: expected $EXPECTED_SIZE bytes, got $ACTUAL_SIZE" >&2
    exit 1
fi
if [[ -n "$EXPECTED_SERIAL" ]]; then
    ACTUAL_SERIAL="$(lsblk -dn -o SERIAL "$TARGET_REAL" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    [[ "$ACTUAL_SERIAL" == "$EXPECTED_SERIAL" ]] || {
        echo "target disk serial changed" >&2
        exit 1
    }
fi
if lsblk -nrpo MOUNTPOINT "$TARGET_REAL" | grep -q '[^[:space:]]'; then
    echo "target disk contains mounted filesystems; refusing destructive install" >&2
    exit 1
fi

if [[ "$TARGET_REAL" =~ [0-9]$ ]]; then
    P1="${TARGET_REAL}p1"; P2="${TARGET_REAL}p2"; P3="${TARGET_REAL}p3"; P4="${TARGET_REAL}p4"
else
    P1="${TARGET_REAL}1"; P2="${TARGET_REAL}2"; P3="${TARGET_REAL}3"; P4="${TARGET_REAL}4"
fi

wipefs --all --force "$TARGET_REAL"
sgdisk --zap-all "$TARGET_REAL"
sgdisk \
    --new=1:2048:+1M --typecode=1:EF02 --change-name=1:'BIOS boot' \
    --new=2:0:+512M --typecode=2:EF00 --change-name=2:'EFI System' \
    --new=3:0:+${BOOT_MIB}M --typecode=3:EA00 --change-name=3:'boot' \
    --new=4:0:0 --typecode=4:8304 --change-name=4:'root' \
    "$TARGET_REAL"

# partprobe is intentionally not required; it was absent in the proven FCOS live environment.
blockdev --rereadpt "$TARGET_REAL" || true
udevadm settle
for _ in $(seq 1 20); do
    [[ -b "$P2" && -b "$P3" && -b "$P4" ]] && break
    sleep 0.25
    udevadm settle
 done
[[ -b "$P2" && -b "$P3" && -b "$P4" ]] || {
    echo "kernel did not expose new target partitions" >&2
    exit 1
}

mkfs.vfat -F 32 -n EFI-SYSTEM "$P2"
mkfs.ext4 -F -L boot "$P3"
mkfs.xfs -f -L root "$P4"

mkdir -p "$TARGET_ROOT"
mount "$P4" "$TARGET_ROOT"
ROOT_MOUNTED=1
mkdir -p "$TARGET_ROOT/boot"
mount "$P3" "$TARGET_ROOT/boot"
BOOT_MOUNTED=1
mkdir -p "$TARGET_ROOT/boot/efi"
mount "$P2" "$TARGET_ROOT/boot/efi"
EFI_MOUNTED=1

ROOT_UUID="$(blkid -s UUID -o value "$P4")"
BOOT_UUID="$(blkid -s UUID -o value "$P3")"
[[ -n "$ROOT_UUID" ]] || { echo "root filesystem UUID is empty; refusing bootc install" >&2; exit 1; }
[[ -n "$BOOT_UUID" ]] || { echo "boot filesystem UUID is empty; refusing bootc install" >&2; exit 1; }

echo "Home Server target UUIDs: root=${ROOT_UUID} boot=${BOOT_UUID}"

# Keep the multi-gigabyte OCI working set on the selected target disk rather
# than the RAM-backed live /var filesystem.
mkdir -p "$PODMAN_SCRATCH" "$IMAGE_TMP" /var/lib/containers
mount --bind "$PODMAN_SCRATCH" "$PODMAN_SCRATCH"
mount --bind "$IMAGE_TMP" "$IMAGE_TMP"
mount --bind "$PODMAN_SCRATCH" /var/lib/containers
PODMAN_MOUNTED=1

# The registry source must pass the Home Server sigstore policy before it is
# admitted into the local target-backed store.
env TMPDIR="$IMAGE_TMP" podman pull --signature-policy "$POLICY_FILE" "$IMAGE"

BOOTC_ARGS=(
    bootc install to-filesystem
    --source-imgref "containers-storage:${IMAGE}"
    --target-imgref "$IMAGE"
    --root-mount-spec "UUID=${ROOT_UUID}"
    --boot-mount-spec "UUID=${BOOT_UUID}"
    --karg=console=tty0
    --karg=console=ttyS0,115200n8
    --generic-image
    --skip-fetch-check
    --skip-finalize
    /target
)

# Root key injection is retained as a recovery path during installation. The
# normal configured user is provisioned below before the first installed boot.
if [[ -s "$SSH_KEYS_FILE" ]]; then
    BOOTC_ARGS=(
        "${BOOTC_ARGS[@]:0:${#BOOTC_ARGS[@]}-1}"
        --root-ssh-authorized-keys /run/home-server-installer.pub
        /target
    )
fi

env TMPDIR="$IMAGE_TMP" podman run --rm --pull=never \
    --privileged \
    --pid=host \
    --ipc=host \
    -v /var/lib/containers:/var/lib/containers \
    -v /dev:/dev \
    -v "${TARGET_ROOT}":/target:rslave \
    -v "${SSH_KEYS_FILE}":/run/home-server-installer.pub:ro \
    --security-opt label=type:unconfined_t \
    "$IMAGE" \
    "${BOOTC_ARGS[@]}"

DEPLOY_BASE="${TARGET_ROOT}/ostree/deploy/fedora-coreos/deploy"
mapfile -t DEPLOYS < <(find "$DEPLOY_BASE" -mindepth 1 -maxdepth 1 -type d -name '*.0' -print)
[[ ${#DEPLOYS[@]} -eq 1 ]] || {
    echo "expected exactly one fresh OSTree deployment, found ${#DEPLOYS[@]}" >&2
    exit 1
}
DEPLOY="${DEPLOYS[0]}"

# Machine identity and SSH policy are written into the actual deployment root.
printf '%s\n' "$HOSTNAME" > "${DEPLOY}/etc/hostname"
if [[ -n "$TIMEZONE" && -e "${DEPLOY}/usr/share/zoneinfo/${TIMEZONE}" ]]; then
    ln -sfn "/usr/share/zoneinfo/${TIMEZONE}" "${DEPLOY}/etc/localtime"
fi
mkdir -p "${DEPLOY}/etc/ssh/sshd_config.d"
if [[ -n "$PASSWORD_HASH" ]]; then
    PASSWORD_AUTH=yes
else
    PASSWORD_AUTH=no
fi
cat > "${DEPLOY}/etc/ssh/sshd_config.d/99-home-server-installer.conf" <<EOF_SSH
PasswordAuthentication ${PASSWORD_AUTH}
PermitRootLogin no
PubkeyAuthentication yes
EOF_SSH
chmod 0600 "${DEPLOY}/etc/ssh/sshd_config.d/99-home-server-installer.conf"

# Provision the selected normal user before sshd starts on first boot. This
# works for the existing core user and for a validated custom username.
mkdir -p "${DEPLOY}/etc/home-server-installer"
printf 'USERNAME=%q\nPASSWORD_HASH=%q\n' "$USERNAME" "$PASSWORD_HASH" > "${DEPLOY}/etc/home-server-installer/user.env"
chmod 0600 "${DEPLOY}/etc/home-server-installer/user.env"
if [[ -s "$SSH_KEYS_FILE" ]]; then
    install -m0600 "$SSH_KEYS_FILE" "${DEPLOY}/etc/home-server-installer/authorized_keys"
else
    : > "${DEPLOY}/etc/home-server-installer/authorized_keys"
    chmod 0600 "${DEPLOY}/etc/home-server-installer/authorized_keys"
fi
cat > "${DEPLOY}/etc/home-server-installer/provision-user.sh" <<'EOF_USER'
#!/usr/bin/bash
set -euo pipefail
source /etc/home-server-installer/user.env
if ! id "$USERNAME" >/dev/null 2>&1; then
    useradd --create-home --user-group --groups wheel "$USERNAME"
fi
if [[ -n "$PASSWORD_HASH" ]]; then
    usermod --password "$PASSWORD_HASH" "$USERNAME"
fi
HOME_DIR="$(getent passwd "$USERNAME" | cut -d: -f6)"
[[ -n "$HOME_DIR" ]] || { echo "could not resolve home for $USERNAME" >&2; exit 1; }
install -d -m0700 -o "$USERNAME" -g "$USERNAME" "$HOME_DIR/.ssh"
if [[ -s /etc/home-server-installer/authorized_keys ]]; then
    install -m0600 -o "$USERNAME" -g "$USERNAME" /etc/home-server-installer/authorized_keys "$HOME_DIR/.ssh/authorized_keys"
fi
rm -f /etc/home-server-installer/authorized_keys /etc/home-server-installer/user.env
systemctl disable home-server-provision-user.service || true
rm -f /etc/systemd/system/multi-user.target.wants/home-server-provision-user.service
rm -f /etc/systemd/system/home-server-provision-user.service
rm -f /etc/home-server-installer/provision-user.sh
rmdir /etc/home-server-installer 2>/dev/null || true
EOF_USER
chmod 0755 "${DEPLOY}/etc/home-server-installer/provision-user.sh"
cat > "${DEPLOY}/etc/systemd/system/home-server-provision-user.service" <<'EOF_UNIT'
[Unit]
Description=Provision Home Server primary user
Before=sshd.service
After=local-fs.target

[Service]
Type=oneshot
ExecStart=/etc/home-server-installer/provision-user.sh

[Install]
WantedBy=multi-user.target
EOF_UNIT
systemctl --root="$DEPLOY" enable home-server-provision-user.service
[[ "$(systemctl --root="$DEPLOY" is-enabled home-server-provision-user.service)" == "enabled" ]] || {
    echo "failed to enable Home Server user provisioning service" >&2
    exit 1
}

if [[ "$NETWORK_MODE" == static ]]; then
    [[ -n "$NETWORK_IFACE" && -n "$NETWORK_ADDR" && -n "$NETWORK_GATEWAY" ]] || {
        echo "static network configuration is incomplete" >&2
        exit 1
    }
    mkdir -p "${DEPLOY}/etc/NetworkManager/system-connections"
    cat > "${DEPLOY}/etc/NetworkManager/system-connections/home-server-static.nmconnection" <<EOF_NET
[connection]
id=home-server-static
interface-name=${NETWORK_IFACE}
type=ethernet
autoconnect=true

[ipv4]
method=manual
address1=${NETWORK_ADDR},${NETWORK_GATEWAY}
dns=${NETWORK_DNS}

[ipv6]
method=auto
EOF_NET
    chmod 0600 "${DEPLOY}/etc/NetworkManager/system-connections/home-server-static.nmconnection"
fi

# Enforce the update policy that is already proven for uCore: Zincati stays
# masked and rpm-ostree stages updates automatically.
systemctl --root="$DEPLOY" disable zincati.service || true
systemctl --root="$DEPLOY" mask zincati.service
systemctl --root="$DEPLOY" enable rpm-ostreed-automatic.timer

# The target-image container is gone; temporary image storage can now be
# removed before bootc finalize and before the first installed boot.
umount /var/lib/containers
PODMAN_MOUNTED=0
umount "$PODMAN_SCRATCH"
umount "$IMAGE_TMP"
rm -rf "$PODMAN_SCRATCH" "$IMAGE_TMP"

bootc install finalize "$TARGET_ROOT"

# Finalization must not drop the first-boot provisioning enablement. Re-enable
# defensively, then fail the install if systemd still does not see it enabled.
systemctl --root="$DEPLOY" enable home-server-provision-user.service
[[ "$(systemctl --root="$DEPLOY" is-enabled home-server-provision-user.service)" == "enabled" ]] || {
    echo "Home Server user provisioning service did not survive bootc finalize" >&2
    exit 1
}
sync

umount "${TARGET_ROOT}/boot/efi"; EFI_MOUNTED=0
umount "${TARGET_ROOT}/boot"; BOOT_MOUNTED=0
umount "$TARGET_ROOT"; ROOT_MOUNTED=0
trap - EXIT

echo "Home Server direct installation complete"
`

// HomeServerInstaller installs the selected signed uCore image directly from
// the Fedora CoreOS live environment using the already-proven bootc layout.
type HomeServerInstaller struct {
	Runner runner.Runner
	Logger *slog.Logger
}

func NewHomeServerInstaller(r runner.Runner, logger *slog.Logger) *HomeServerInstaller {
	return &HomeServerInstaller{Runner: r, Logger: logger}
}

func (i *HomeServerInstaller) Install(ctx context.Context, cfg *model.InstallConfig, progress func(step string)) error {
	if cfg == nil {
		return fmt.Errorf("install config cannot be nil")
	}
	if cfg.HomeServerImage != model.HomeServerUCoreImage && cfg.HomeServerImage != model.HomeServerUCoreHCIImage {
		return fmt.Errorf("unsupported Home Server image %q", cfg.HomeServerImage)
	}
	bootMiB := cfg.HomeServerBootSizeMiB
	if bootMiB == 0 {
		bootMiB = model.HomeServerBootStandardMiB
	}
	if bootMiB != model.HomeServerBootStandardMiB && bootMiB != model.HomeServerBootLargeMiB {
		return fmt.Errorf("unsupported Home Server /boot size %d MiB", bootMiB)
	}
	if len(cfg.Users) == 0 || strings.TrimSpace(cfg.Users[0].Username) == "" {
		return fmt.Errorf("Home Server install requires a primary user")
	}

	sshFile, err := writePrivateTemp("knuckle-home-server-keys-*", strings.Join(cfg.SSHKeys, "\n")+"\n")
	if err != nil {
		return fmt.Errorf("writing temporary SSH keys: %w", err)
	}
	defer os.Remove(sshFile)

	keyFile, err := writePrivateTemp("knuckle-home-server-cosign-*", homeServerInstallerCosignPublicKey)
	if err != nil {
		return fmt.Errorf("writing temporary image verification key: %w", err)
	}
	defer os.Remove(keyFile)

	policy := fmt.Sprintf(`{
  "default": [{"type":"reject"}],
  "transports": {
    "docker": {
      %q: [{
        "type":"sigstoreSigned",
        "keyPath": %q,
        "signedIdentity":{"type":"matchRepository"}
      }]
    }
  }
}
`, strings.TrimSuffix(cfg.HomeServerImage, ":lts"), keyFile)
	policyFile, err := writePrivateTemp("knuckle-home-server-policy-*", policy)
	if err != nil {
		return fmt.Errorf("writing temporary signature policy: %w", err)
	}
	defer os.Remove(policyFile)

	passwordHash := ""
	if len(cfg.Users) > 0 {
		passwordHash = cfg.Users[0].PasswordHash
	}
	stablePath := cfg.Disk.Path
	if stablePath == cfg.Disk.DevPath {
		stablePath = ""
	}

	progress("Preparing Home Server disk layout...")
	i.Logger.Info("executing Home Server direct bootc install",
		"disk", cfg.Disk.DevPath,
		"image", cfg.HomeServerImage,
		"boot_mib", bootMiB)

	args := []string{
		"-s", "--",
		cfg.Disk.DevPath,
		stablePath,
		fmt.Sprintf("%d", cfg.Disk.Size),
		cfg.Disk.Serial,
		cfg.HomeServerImage,
		fmt.Sprintf("%d", bootMiB),
		cfg.Users[0].Username,
		passwordHash,
		cfg.Hostname,
		cfg.Timezone,
		cfg.Network.Mode.String(),
		cfg.Network.Interface,
		cfg.Network.Address,
		cfg.Network.Gateway,
		strings.Join(cfg.Network.DNS, ";"),
		sshFile,
		policyFile,
		keyFile,
	}
	result, runErr := i.Runner.RunWithInput(ctx, homeServerDirectInstallScript, "bash", args...)
	if runErr != nil || (result != nil && result.ExitCode != 0) {
		return formatCommandError("Home Server direct installation failed", result, runErr)
	}

	progress("Installation complete!")
	return nil
}

func writePrivateTemp(pattern, content string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := f.Name()
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if err := f.Chmod(0600); err != nil {
		return "", err
	}
	if _, err := f.WriteString(content); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	ok = true
	return path, nil
}