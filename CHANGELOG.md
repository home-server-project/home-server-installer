<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
# Changelog

This changelog tracks the Home Server Installer fork and its Home Server-specific work.

The inherited upstream Knuckle release history has been removed here because it describes the original installer rather than the current Home Server product direction. Upstream history remains available in the [Project Bluefin Knuckle repository](https://github.com/projectbluefin/knuckle).

## [Unreleased]

No unreleased changes are currently documented.

## [1.1.0] - 2026-09-10

### Highlights
- Replaced the flat Home Server image list with a two-level picker: select an image family first, then select an edition.
- Expanded the installer from five to **11 signed LTS installation targets** across four families:
  - Home Server Gina LTS.
  - Universal Blue uCore LTS.
  - Universal Blue uCore LTS / NVIDIA Open.
  - Universal Blue uCore LTS / NVIDIA LTS.
- Added signed NVIDIA Open targets:
  - `ghcr.io/ublue-os/ucore-minimal:lts-nvidia`.
  - `ghcr.io/ublue-os/ucore:lts-nvidia`.
  - `ghcr.io/ublue-os/ucore-hci:lts-nvidia`.
- Added signed NVIDIA LTS targets:
  - `ghcr.io/ublue-os/ucore-minimal:lts-nvidia-lts`.
  - `ghcr.io/ublue-os/ucore:lts-nvidia-lts`.
  - `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts`.
- All upstream uCore NVIDIA variants use the Universal Blue uCore Cosign verification key and the same repository-matching signature policy as the standard upstream uCore targets.
- Added installer version display to the TUI.
- The 2026-09-10 VM acceptance test checked all 11 picker choices against the expected image tags and completed a full signed install of **uCore HCI LTS NVIDIA LTS** using `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts`.
- Main CI passed after promotion of the tested picker changes.

### Working now
- The Home Server image picker exposes 11 signed LTS targets:
  - `ghcr.io/home-server-project/home-server-gina:lts` — Home Server Gina LTS.
  - `ghcr.io/home-server-project/home-server-gina-hci:lts` — Home Server Gina HCI LTS.
  - `ghcr.io/ublue-os/ucore-minimal:lts` — upstream Universal Blue uCore Minimal LTS.
  - `ghcr.io/ublue-os/ucore:lts` — upstream Universal Blue uCore LTS.
  - `ghcr.io/ublue-os/ucore-hci:lts` — upstream Universal Blue uCore HCI LTS.
  - `ghcr.io/ublue-os/ucore-minimal:lts-nvidia` — upstream uCore Minimal LTS NVIDIA Open.
  - `ghcr.io/ublue-os/ucore:lts-nvidia` — upstream uCore LTS NVIDIA Open.
  - `ghcr.io/ublue-os/ucore-hci:lts-nvidia` — upstream uCore HCI LTS NVIDIA Open.
  - `ghcr.io/ublue-os/ucore-minimal:lts-nvidia-lts` — upstream uCore Minimal LTS NVIDIA LTS.
  - `ghcr.io/ublue-os/ucore:lts-nvidia-lts` — upstream uCore LTS NVIDIA LTS.
  - `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts` — upstream uCore HCI LTS NVIDIA LTS.
- The installer intentionally exposes LTS targets. Stable and testing remain post-install `bootc switch` destinations.
- Home Server Project images are verified with the Home Server Project cosign public key; upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore cosign public key. Unknown image repositories remain rejected.
- Sigstore attachment discovery is enabled temporarily for all supported repositories during the signed image pull.
- Home Server installs use a dedicated direct `bootc install to-filesystem` path instead of the stock `coreos-installer install` partition layout.
- Home Server storage layout supports two explicit XBOOTLDR presets:
  - 1 GiB `/boot` — standard and recommended for Gina / Gina HCI / upstream uCore.
  - 2 GiB `/boot` — large layout for NVIDIA or custom-image use cases.
- The direct install path creates:
  - 1 MiB BIOS boot partition.
  - 512 MiB EFI System Partition.
  - ext4 XBOOTLDR `/boot` using the selected preset.
  - XFS root using the remaining disk.
- Destructive target selection is revalidated immediately before partitioning, including block-device identity, expected size, serial when available, and mounted-filesystem checks.
- Supported container images are pulled with enforced sigstore verification using the trust key assigned to the selected repository.
- Temporary signature-discovery configuration and installer Podman scratch state are cleaned up and do not persist into the installed system.
- The selected image is installed directly as the first bootable deployment; there is no installed Fedora CoreOS intermediate and no autorebase bootstrap service.
- Primary-user provisioning is written directly into the target deployment before first boot.
- The installer supports setting a local password during installation in addition to SSH authorized-key provisioning.
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
- [Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder) provides the GitHub-based personalized ISO path. Its SSH flow imports only the user's public SSH key; private keys remain on the user's own device.

### Verified in end-to-end testing
Clean VM runs completed through the normal installer path with no Podman wrapper, custom `PATH`, service stop, or signature bypass.

Verified results:
- Signed Home Server Gina HCI image pull and install completed successfully.
- First boot went directly into Home Server Gina HCI.
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
- Builder SSH handoff was verified end to end; the ISO builder received the lab public key only through `--ssh-key`, with no manual SSH-key paste into the TUI.
- The same builder-provided key worked for SSH into the live installer and, after installation and reboot, for SSH into the installed `core` account.
- The installed key was present at `/var/home/core/.ssh/authorized_keys` with mode `0600`, while `/var/home/core/.ssh` had mode `0700`; both were owned by `core:core`.
- The dedicated live-only `/opt/home-server-installer-ssh.pub` handoff file was absent from the installed system, and `/opt/knuckle` was also absent.
- In the password-backed builder-key run, SSH key login worked while `sudo -n true` correctly failed with `a password is required` after `sudo -k`.
- Upstream `ghcr.io/ublue-os/ucore-minimal:lts` completed signed pull/install and first booted directly into `uCore minimal`.
- The 1 GiB `/boot` layout was selected and installed successfully for the uCore Minimal run.
- Reinstalling the VM at the same IP produced the expected SSH host-key-change warning. Removing the stale client entry with `ssh-keygen -R IP` and reconnecting succeeded normally.
- The four-family picker was exercised in the 2026-09-10 acceptance test and all 11 entries mapped to the intended image tags.
- `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts` completed a signed pull and full VM installation successfully.

Bare-metal validation has also completed successfully using dedicated test hardware, including Home Server Gina HCI LTS with a 1 GiB `/boot` layout and upstream uCore LTS with a 2 GiB `/boot` layout.

### Still to validate end-to-end
- Validate the SSH-key-only path with a blank password and confirm the intended `NOPASSWD` sudo rule on the finished machine.
- Continue adding hardware coverage as useful, especially for NVIDIA systems.

### Notes
- Installation progress may remain around 20% for several minutes while the container image is downloaded and deployed. This is a known cosmetic limitation; do not power off or reboot during this stage.
- V1 is working in end-to-end VM and bare-metal testing. A VM test is still recommended before first use on physical hardware.
