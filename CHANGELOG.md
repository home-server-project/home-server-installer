# Changelog

This changelog tracks the Home Server Installer fork and its Home Server-specific work.

The inherited upstream Knuckle release history has been removed here because it describes the original installer rather than the current Home Server product direction. Upstream history remains available in the Project Bluefin Knuckle repository.

## [Unreleased]

### Working now
- Home Server image selection is limited to the two supported targets:
  - `ghcr.io/home-server-project/home-server-ucore:lts`
  - `ghcr.io/home-server-project/home-server-ucore-hci:lts`
- Home Server installs use a dedicated direct `bootc install to-filesystem` path instead of the stock `coreos-installer install` partition layout.
- Home Server storage layout supports two explicit XBOOTLDR presets:
  - 1 GiB `/boot` — standard and recommended for uCore / uCore HCI.
  - 2 GiB `/boot` — large layout for NVIDIA or custom-image use cases.
- The direct install path creates:
  - 1 MiB BIOS boot partition.
  - 512 MiB EFI System Partition.
  - ext4 XBOOTLDR `/boot` using the selected preset.
  - XFS root using the remaining disk.
- Destructive target selection is revalidated immediately before partitioning, including block-device identity, expected size, serial when available, and mounted-filesystem checks.
- Home Server container images are pulled with enforced sigstore verification using the embedded Home Server cosign public key.
- Sigstore attachment discovery is enabled temporarily for both Home Server image repositories during installation, fixing the previous `A signature was required, but no signature exists` failure.
- Temporary signature-discovery configuration and installer Podman scratch state are cleaned up and do not persist into the installed system.
- The selected image is installed directly as the first bootable deployment; there is no installed Fedora CoreOS intermediate and no autorebase bootstrap service.
- Primary-user provisioning is written directly into the target deployment before first boot.
- Persistent `/var/home/<user>` creation, ownership, permissions, and SELinux labeling survive `bootc install finalize`.
- SSH authorized keys supplied to the installer are written to the selected user's persistent home and verified after finalize.
- Password-backed users retain normal password-required sudo behavior.
- SSH-key-only/passwordless admins receive the intended passwordless sudo rule.
- Home Server update policy is persisted with `/etc/systemd/system-preset/00-home-server.preset`:
  - Zincati disabled/masked.
  - `rpm-ostreed-automatic.timer` enabled.
  - `AutomaticUpdatePolicy=stage` retained.
- Installer residue checks confirm no installed `home-server-installer.service`, no `/opt/knuckle`, and no temporary signature-discovery file remain after installation.
- The self-contained Fedora CoreOS live ISO builder supports an optional public SSH key for access to the live installer environment.
- The builder also places that public key in a dedicated live-only `/opt/home-server-installer-ssh.pub` handoff file. The Home Server TUI reads that explicit file and merges the key into the normal installer SSH-key configuration without depending on `$HOME` or scanning the live `core` account's authorized-key files.

### Verified in clean VM end-to-end testing
The 2026-09-07 clean uCore HCI VM runs completed through the normal installer path with no Podman wrapper, custom `PATH`, service stop, or signature bypass.

Verified results:
- Signed HCI image pull and install completed successfully.
- First boot went directly into Home Server uCore HCI.
- 1 GiB standard `/boot` layout was correct.
- The separate 20 GiB sentinel disk remained untouched.
- Primary user was created as UID/GID 1000 and added to `wheel`.
- Password-backed sudo correctly required a password after clearing the sudo authentication cache.
- Zincati was `masked`.
- `rpm-ostreed-automatic.timer` was `enabled` and `active`.
- `AutomaticUpdatePolicy=stage` existed in both `/etc/rpm-ostreed.conf` and `/usr/etc/rpm-ostreed.conf`.
- `home-server-autorebase.service` was `not-found` as intended.
- `/var/home/core` survived finalize with correct ownership and permissions.
- `00-home-server.preset` survived finalize with mode `0644` and the intended policy contents.
- No passwordless sudoers file existed for the password-backed test user.
- `systemctl --failed` reported zero failed units.
- No installer signature configuration, installer systemd service, or `/opt/knuckle` binary remained in the installed system.
- Builder SSH handoff was verified end-to-end with Knuckle `9a73fa3` and ISO SHA256 `a778ff4bd09ff90e558285132c536ae98ed5c16051a9f3abf5b03aeaf5c22c5b`.
- The ISO builder received the lab public key only through `--ssh-key`; no SSH key was pasted manually into the TUI.
- The same builder-provided key worked for SSH into the live installer and, after installation and reboot, for SSH into the installed `core` account.
- The installed key was present at `/var/home/core/.ssh/authorized_keys` with mode `0600`, while `/var/home/core/.ssh` had mode `0700`; both were owned by `core:core`.
- The installed ED25519 key fingerprint matched the expected `home-server-installer-lab` key.
- The dedicated live-only `/opt/home-server-installer-ssh.pub` handoff file was absent from the installed system, and `/opt/knuckle` was also absent.
- In the password-backed builder-key run, SSH key login worked while `sudo -n true` correctly failed with `a password is required` after `sudo -k`.

### Still to validate end-to-end
- Validate the SSH-key-only path with a blank password and confirm the intended `NOPASSWD` sudo rule on the finished machine.

### Notes
- Installation progress may remain around 20% for several minutes while the container image is downloaded and deployed. This is expected; do not power off or reboot during this stage.
