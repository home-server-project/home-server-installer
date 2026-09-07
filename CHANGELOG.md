# Changelog

This changelog tracks the Home Server Installer fork and its Home Server-specific work.

The inherited upstream Knuckle release history has been removed here because it describes the original installer rather than the current Home Server product direction. Upstream history remains available in the Project Bluefin Knuckle repository.

## [Unreleased]

### Working now
- The Home Server image picker exposes five signed LTS targets:
  - `ghcr.io/home-server-project/home-server-ucore:lts` — Home Server uCore LTS (default/recommended).
  - `ghcr.io/home-server-project/home-server-ucore-hci:lts` — Home Server uCore HCI LTS.
  - `ghcr.io/ublue-os/ucore-minimal:lts` — upstream Universal Blue uCore Minimal LTS.
  - `ghcr.io/ublue-os/ucore:lts` — upstream Universal Blue uCore LTS.
  - `ghcr.io/ublue-os/ucore-hci:lts` — upstream Universal Blue uCore HCI LTS.
- The installer intentionally exposes only LTS targets. Stable and NVIDIA variants remain available as post-install `bootc switch` destinations instead of expanding the installer into a large image matrix.
- Home Server Project images are verified with the Home Server Project cosign public key; upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore cosign public key. Unknown image repositories remain rejected.
- Sigstore attachment discovery is enabled temporarily for all five supported repositories during the signed image pull.
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
- A separate future [Home Server uCore Builder](https://github.com/home-server-project/home-server-ucore-builder) template is planned for GitHub-based personalized ISO creation. Its intended SSH flow is to import only the user's public SSH key through GitHub Secrets; private keys remain on the user's own device.

### Verified in clean VM end-to-end testing
The 2026-09-07 clean VM runs completed through the normal installer path with no Podman wrapper, custom `PATH`, service stop, or signature bypass.

Verified results:
- Signed Home Server HCI image pull and install completed successfully.
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
- An upstream Universal Blue representative target was also validated end to end using `ghcr.io/ublue-os/ucore-minimal:lts`.
- The uCore Minimal VM completed signed pull/install and first booted directly into `uCore minimal`.
- The 1 GiB `/boot` layout was selected and installed successfully for the uCore Minimal run.
- A second independent ISO build used the laptop's `id_ed25519_lab.pub` key through `--ssh-key`, with no manual SSH-key paste in the TUI.
- After installation, plain `ssh core@IP` succeeded from the laptop because the matching private key was already loaded in `ssh-agent`; no `-i` option was required.
- The test confirms that the local `.pub` filename is not part of server-side SSH authorization; only the public-key contents matter.
- Reinstalling the VM at the same IP produced the expected SSH host-key-change warning. Removing the stale client entry with `ssh-keygen -R IP` and reconnecting succeeded normally.

### Still to validate end-to-end
- Validate the SSH-key-only path with a blank password and confirm the intended `NOPASSWD` sudo rule on the finished machine.
- Continue bare-metal validation only on dedicated test hardware before considering any real home-server or production use.

### Notes
- Installation progress may remain around 20% for several minutes while the container image is downloaded and deployed. This is expected; do not power off or reboot during this stage.
- V1 is working in disposable VM testing, but testing on dedicated hardware should be treated as the next validation stage. Avoid real home servers or production systems until that testing is complete.
