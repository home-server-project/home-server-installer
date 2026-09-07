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
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWU3SeANKBm2Dql6FGZYNu2Bd7nZf
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
REGISTRIES_DIR=/etc/containers/registries.d
REGISTRIES_FILE=""
PODMAN_MOUNTED=0
ROOT_MOUNTED=0
BOOT_MOUNTED=0
EFI_MOUNTED=0

cleanup_mounts() {
    set +e
    if [[ -n "$REGISTRIES_FILE" ]]; then rm -f -- "$REGISTRIES_FILE"; fi
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

for cmd in readlink lsblk wipefs sgdisk blockdev udevadm mkfs.vfat mkfs.ext4 mkfs.xfs mount umount mountpoint blkid podman bootc systemctl find install mktemp useradd usermod awk chown stat chcon matchpathcon; do
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

mkdir -p "$PODMAN_SCRATCH" "$IMAGE_TMP" /var/lib/containers
mount --bind "$PODMAN_SCRATCH" "$PODMAN_SCRATCH"
mount --bind "$IMAGE_TMP" "$IMAGE_TMP"
mount --bind "$PODMAN_SCRATCH" /var/lib/containers
PODMAN_MOUNTED=1

install -d -m0755 "$REGISTRIES_DIR"
REGISTRIES_FILE="$(mktemp "${REGISTRIES_DIR}/00-home-server-installer.XXXXXX.yaml")"
cat > "$REGISTRIES_FILE" <<'EOF_REGISTRIES'
docker:
  ghcr.io/home-server-project/home-server-ucore:
    use-sigstore-attachments: true
  ghcr.io/home-server-project/home-server-ucore-hci:
    use-sigstore-attachments: true
EOF_REGISTRIES
chmod 0644 "$REGISTRIES_FILE"

env TMPDIR="$IMAGE_TMP" podman pull --signature-policy "$POLICY_FILE" "$IMAGE"
rm -f -- "$REGISTRIES_FILE"
REGISTRIES_FILE=""

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

if grep -q "^${USERNAME}:" "${DEPLOY}/etc/passwd"; then
    usermod --root "$DEPLOY" --append --groups wheel "$USERNAME"
else
    useradd --root "$DEPLOY" \
        --no-create-home \
        --user-group \
        --groups wheel \
        --home-dir "/var/home/${USERNAME}" \
        --shell /bin/bash \
        --comment "Home Server Admin" \
        "$USERNAME"
fi

if [[ -n "$PASSWORD_HASH" ]]; then
    usermod --root "$DEPLOY" --password "$PASSWORD_HASH" "$USERNAME"
else
    usermod --root "$DEPLOY" --lock "$USERNAME"
fi

USER_ENTRY="$(awk -F: -v user="$USERNAME" '$1 == user { print; exit }' "${DEPLOY}/etc/passwd")"
[[ -n "$USER_ENTRY" ]] || { echo "selected user was not written to target passwd" >&2; exit 1; }
IFS=: read -r _ _ USER_UID USER_GID _ USER_HOME USER_SHELL <<< "$USER_ENTRY"
[[ "$USER_UID" =~ ^[0-9]+$ && "$USER_GID" =~ ^[0-9]+$ ]] || { echo "selected user has invalid target UID/GID" >&2; exit 1; }
[[ "$USER_HOME" == "/var/home/${USERNAME}" ]] || { echo "selected user has unexpected home directory: ${USER_HOME}" >&2; exit 1; }
[[ "$USER_SHELL" == "/bin/bash" ]] || { echo "selected user has unexpected shell: ${USER_SHELL}" >&2; exit 1; }

WHEEL_MEMBERS="$(awk -F: '$1 == "wheel" { print $4; exit }' "${DEPLOY}/etc/group")"
case ",${WHEEL_MEMBERS}," in
    *",${USERNAME},"*) ;;
    *) echo "selected user was not added to wheel" >&2; exit 1 ;;
esac

SHADOW_HASH="$(awk -F: -v user="$USERNAME" '$1 == user { print $2; exit }' "${DEPLOY}/etc/shadow")"
if [[ -n "$PASSWORD_HASH" ]]; then
    [[ "$SHADOW_HASH" == "$PASSWORD_HASH" ]] || { echo "selected user password hash was not written to target shadow" >&2; exit 1; }
else
    [[ "$SHADOW_HASH" == '!'* || "$SHADOW_HASH" == '*'* ]] || { echo "selected user password is not locked for SSH-only install" >&2; exit 1; }
fi

SUDOERS_FILE="${DEPLOY}/etc/sudoers.d/90-home-server-admin"
mkdir -p "${DEPLOY}/etc/sudoers.d"
if [[ -z "$PASSWORD_HASH" ]]; then
    printf '%s ALL=(ALL) NOPASSWD: ALL\n' "$USERNAME" > "$SUDOERS_FILE"
    chmod 0440 "$SUDOERS_FILE"
else
    rm -f "$SUDOERS_FILE"
fi

PRESET_FILE="${DEPLOY}/etc/systemd/system-preset/00-home-server.preset"
mkdir -p "${DEPLOY}/etc/systemd/system-preset"
cat > "$PRESET_FILE" <<'EOF_PRESET'
disable zincati.service
enable rpm-ostreed-automatic.timer
EOF_PRESET
chmod 0644 "$PRESET_FILE"

PERSISTENT_VAR_ROOT="${TARGET_ROOT}/ostree/deploy/fedora-coreos/var"
[[ -d "$PERSISTENT_VAR_ROOT" ]] || { echo "persistent target stateroot /var is missing" >&2; exit 1; }
PERSISTENT_HOME_ROOT="${PERSISTENT_VAR_ROOT}/home"
PERSISTENT_HOME="${PERSISTENT_HOME_ROOT}/${USERNAME}"
install -d -m0755 "$PERSISTENT_HOME_ROOT"
HOME_ROOT_CONTEXT="$(matchpathcon -n "/var/home")"
[[ -n "$HOME_ROOT_CONTEXT" && "$HOME_ROOT_CONTEXT" != "<<none>>" ]] || { echo "could not resolve SELinux context for /var/home" >&2; exit 1; }
chcon "$HOME_ROOT_CONTEXT" "$PERSISTENT_HOME_ROOT"

install -d -m0700 "$PERSISTENT_HOME"
chown "${USER_UID}:${USER_GID}" "$PERSISTENT_HOME"
HOME_CONTEXT="$(matchpathcon -n "/var/home/${USERNAME}")"
[[ -n "$HOME_CONTEXT" && "$HOME_CONTEXT" != "<<none>>" ]] || { echo "could not resolve SELinux context for user home" >&2; exit 1; }
chcon "$HOME_CONTEXT" "$PERSISTENT_HOME"

if [[ -s "$SSH_KEYS_FILE" ]]; then
    install -d -m0700 "$PERSISTENT_HOME/.ssh"
    install -m0600 "$SSH_KEYS_FILE" "$PERSISTENT_HOME/.ssh/authorized_keys"
    chown -R "${USER_UID}:${USER_GID}" "$PERSISTENT_HOME/.ssh"
    SSH_CONTEXT="$(matchpathcon -n "/var/home/${USERNAME}/.ssh")"
    AUTH_KEYS_CONTEXT="$(matchpathcon -n "/var/home/${USERNAME}/.ssh/authorized_keys")"
    [[ -n "$SSH_CONTEXT" && "$SSH_CONTEXT" != "<<none>>" ]] || { echo "could not resolve SELinux context for user SSH directory" >&2; exit 1; }
    [[ -n "$AUTH_KEYS_CONTEXT" && "$AUTH_KEYS_CONTEXT" != "<<none>>" ]] || { echo "could not resolve SELinux context for authorized_keys" >&2; exit 1; }
    chcon "$SSH_CONTEXT" "$PERSISTENT_HOME/.ssh"
    chcon "$AUTH_KEYS_CONTEXT" "$PERSISTENT_HOME/.ssh/authorized_keys"
fi

[[ "$(stat -c %a "$PERSISTENT_HOME_ROOT")" == "755" ]] || { echo "persistent /var/home has wrong permissions" >&2; exit 1; }
[[ "$(stat -c %C "$PERSISTENT_HOME_ROOT")" == "$HOME_ROOT_CONTEXT" ]] || { echo "persistent /var/home has wrong SELinux context" >&2; exit 1; }
[[ "$(stat -c %u "$PERSISTENT_HOME")" == "$USER_UID" ]] || { echo "persistent user home has wrong owner" >&2; exit 1; }
[[ "$(stat -c %g "$PERSISTENT_HOME")" == "$USER_GID" ]] || { echo "persistent user home has wrong group" >&2; exit 1; }
[[ "$(stat -c %a "$PERSISTENT_HOME")" == "700" ]] || { echo "persistent user home has wrong permissions" >&2; exit 1; }
[[ "$(stat -c %C "$PERSISTENT_HOME")" == "$HOME_CONTEXT" ]] || { echo "persistent user home has wrong SELinux context" >&2; exit 1; }

if [[ -s "$SSH_KEYS_FILE" ]]; then
    [[ -s "$PERSISTENT_HOME/.ssh/authorized_keys" ]] || { echo "authorized_keys was not written to persistent user home" >&2; exit 1; }
    [[ "$(stat -c %u "$PERSISTENT_HOME/.ssh/authorized_keys")" == "$USER_UID" ]] || { echo "authorized_keys has wrong owner" >&2; exit 1; }
    [[ "$(stat -c %g "$PERSISTENT_HOME/.ssh/authorized_keys")" == "$USER_GID" ]] || { echo "authorized_keys has wrong group" >&2; exit 1; }
    [[ "$(stat -c %a "$PERSISTENT_HOME/.ssh/authorized_keys")" == "600" ]] || { echo "authorized_keys has wrong permissions" >&2; exit 1; }
    [[ "$(stat -c %C "$PERSISTENT_HOME/.ssh")" == "$SSH_CONTEXT" ]] || { echo "user SSH directory has wrong SELinux context" >&2; exit 1; }
    [[ "$(stat -c %C "$PERSISTENT_HOME/.ssh/authorized_keys")" == "$AUTH_KEYS_CONTEXT" ]] || { echo "authorized_keys has wrong SELinux context" >&2; exit 1; }
fi

rm -rf "${DEPLOY}/etc/home-server-installer"
rm -f \
    "${DEPLOY}/etc/systemd/system/home-server-provision-user.service" \
    "${DEPLOY}/etc/systemd/system/multi-user.target.wants/home-server-provision-user.service"

if [[ "$NETWORK_MODE" == static ]]; then
    [[ -n "$NETWORK_IFACE" && -n "$NETWORK_ADDR" && -n "$NETWORK_GATEWAY" ]] || { echo "static network configuration is incomplete" >&2; exit 1; }
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

umount /var/lib/containers
PODMAN_MOUNTED=0
umount "$PODMAN_SCRATCH"
umount "$IMAGE_TMP"
rm -rf "$PODMAN_SCRATCH" "$IMAGE_TMP"

bootc install finalize "$TARGET_ROOT"

systemctl --root="$DEPLOY" disable zincati.service || true
systemctl --root="$DEPLOY" mask zincati.service
systemctl --root="$DEPLOY" preset rpm-ostreed-automatic.timer

[[ "$(systemctl --root="$DEPLOY" is-enabled zincati.service 2>/dev/null || true)" == "masked" ]] || { echo "zincati is not masked in finalized target" >&2; exit 1; }
[[ "$(systemctl --root="$DEPLOY" is-enabled rpm-ostreed-automatic.timer 2>/dev/null || true)" == "enabled" ]] || { echo "rpm-ostreed-automatic.timer is not enabled in finalized target" >&2; exit 1; }
[[ -f "$PRESET_FILE" ]] || { echo "Home Server update preset did not survive bootc finalize" >&2; exit 1; }
[[ "$(stat -c %a "$PRESET_FILE")" == "644" ]] || { echo "Home Server update preset has wrong permissions" >&2; exit 1; }
grep -Fxq "disable zincati.service" "$PRESET_FILE" || { echo "Home Server update preset is missing Zincati policy" >&2; exit 1; }
grep -Fxq "enable rpm-ostreed-automatic.timer" "$PRESET_FILE" || { echo "Home Server update preset is missing rpm-ostree timer policy" >&2; exit 1; }

grep -q "^${USERNAME}:" "${DEPLOY}/etc/passwd" || { echo "selected user did not survive bootc finalize" >&2; exit 1; }
[[ -d "$PERSISTENT_HOME" ]] || { echo "persistent user home did not survive bootc finalize" >&2; exit 1; }
if [[ -s "$SSH_KEYS_FILE" ]]; then
    [[ -s "$PERSISTENT_HOME/.ssh/authorized_keys" ]] || { echo "authorized_keys did not survive bootc finalize" >&2; exit 1; }
fi
if [[ -z "$PASSWORD_HASH" ]]; then
    [[ -f "$SUDOERS_FILE" ]] || { echo "SSH-only admin sudoers file did not survive bootc finalize" >&2; exit 1; }
    [[ "$(stat -c %a "$SUDOERS_FILE")" == "440" ]] || { echo "SSH-only admin sudoers file has wrong permissions" >&2; exit 1; }
    grep -Fxq "${USERNAME} ALL=(ALL) NOPASSWD: ALL" "$SUDOERS_FILE" || { echo "SSH-only admin sudoers rule is incorrect" >&2; exit 1; }
else
    [[ ! -e "$SUDOERS_FILE" ]] || { echo "password-backed admin unexpectedly has passwordless sudo rule" >&2; exit 1; }
fi
[[ ! -e "${DEPLOY}/etc/systemd/system/home-server-provision-user.service" ]] || { echo "obsolete first-boot provisioning service remains in target" >&2; exit 1; }
sync

umount "${TARGET_ROOT}/boot/efi"; EFI_MOUNTED=0
umount "${TARGET_ROOT}/boot"; BOOT_MOUNTED=0
umount "$TARGET_ROOT"; ROOT_MOUNTED=0
trap - EXIT

echo "Home Server direct installation complete"
`

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

	sshKeyContent := ""
	if len(cfg.SSHKeys) > 0 {
		sshKeyContent = strings.Join(cfg.SSHKeys, "\n") + "\n"
	}
	sshFile, err := writePrivateTemp("knuckle-home-server-keys-*", sshKeyContent)
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
	i.Logger.Info("executing Home Server direct bootc install", "disk", cfg.Disk.DevPath, "image", cfg.HomeServerImage, "boot_mib", bootMiB)

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
