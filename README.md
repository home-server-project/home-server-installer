<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
<p align="center">
  <img src="https://raw.githubusercontent.com/home-server-project/.github/main/logo/banner-navy-mid.png" alt="Home Server Project banner">
</p>

# Home Server Installer

**Current release:** [![GitHub Release](https://img.shields.io/github/v/release/home-server-project/home-server-installer?label=latest)](https://github.com/home-server-project/home-server-installer/releases/latest)

> [!WARNING]
> **V1 — VM testing is recommended first.**
>
> - For bare-metal testing, use a **dedicated test drive or hardware where the selected installation disk can be safely erased**, and keep backups of anything important.
> - The installer provides **target-disk selection**, **1 GiB / 2 GiB `/boot` layout selection**, **local password setup**, and **SSH key configuration**.
> - The installer erases and repartitions **only the selected target disk**; all installer-created partitions are placed on that selected drive.
> - **UEFI only** for the current V1 path.
> - **Secure Boot must be disabled during installation.** After installation, follow the current [uCore Secure Boot instructions](https://github.com/ublue-os/ucore) if you want to enable Secure Boot.

Home Server Installer is a friendly Fedora CoreOS-based installer for [Home Server Gina](https://github.com/home-server-project/home-server-gina) and selected upstream [Universal Blue uCore](https://github.com/ublue-os/ucore) images.

It is a thin downstream adaptation of [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle). The goal is to keep Knuckle's proven TUI and hardware discovery while providing a Home Server-focused, signed, direct-install path for Gina and uCore.

## V1 image choices

The installer currently exposes five signed LTS targets:

- `ghcr.io/home-server-project/home-server-gina:lts` — Home Server Gina LTS (default/recommended).
- `ghcr.io/home-server-project/home-server-gina-hci:lts` — Home Server Gina HCI LTS.
- `ghcr.io/ublue-os/ucore-minimal:lts` — upstream Universal Blue uCore Minimal LTS.
- `ghcr.io/ublue-os/ucore:lts` — upstream Universal Blue uCore LTS.
- `ghcr.io/ublue-os/ucore-hci:lts` — upstream Universal Blue uCore HCI LTS.

The installer intentionally keeps the menu to LTS targets. Stable, NVIDIA and custom images can remain post-install `bootc switch` destinations instead of turning the installer into a large image matrix.

## V1 flow

```text
Fedora CoreOS live ISO + Home Server Installer
                 |
                 v
     choose signed Gina/uCore image
      disk / boot layout / user / SSH
                 |
                 v
    verify and pull selected image
                 |
                 v
       partition selected disk
                 |
                 v
   direct bootc install to filesystem
                 |
                 v
               reboot
                 |
                 v
       selected image boots
```

The selected Gina or upstream uCore image is installed directly as the first bootable deployment. There is no installed Fedora CoreOS intermediate and no first-boot autorebase step.

## What to expect when booting and installing

> [!NOTE]
> **The installer may take a few minutes to appear after booting.** Fedora CoreOS is starting in the background before the Home Server Installer UI launches, so a short wait is normal.
>
> During installation, the progress bar may remain around **20% for several minutes** while the selected container image is downloaded, verified, and deployed.
>
> **This is expected. Do not power off or reboot the machine while installation is in progress.**
>
> Once that stage completes, installation normally advances quickly to completion.

## Signed image installation

Home Server Project images are verified with the Home Server Project Cosign public key. Upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore Cosign public key.

Only supported repositories are accepted by the Home Server path. Unknown image repositories are rejected. Temporary signature-discovery configuration used during installation is cleaned up and does not persist into the installed system.

## Storage layout

V1 uses a direct Home Server storage layout on the user-selected target disk:

- 1 MiB BIOS boot partition.
- 512 MiB EFI System Partition.
- ext4 XBOOTLDR `/boot` with one of two presets:
  - **1 GiB** — standard and recommended for Gina / Gina HCI and upstream uCore targets.
  - **2 GiB** — large layout for NVIDIA or custom-image use cases.
- XFS `/` using the remaining disk.

The selected target disk is revalidated immediately before destructive partitioning, including device identity, expected size, serial when available, and mounted-filesystem checks.

Clean VM testing has also verified that a separate attached non-target sentinel disk remains untouched during installation. Bare-metal testing can follow on a dedicated test drive or test hardware where the selected installation disk can be safely erased.

## User and SSH access

The installer can create the primary user during installation and supports a **local password**, SSH authorized keys, or a key-only administration path.

Password-backed users keep normal password-required `sudo` behavior. The SSH-key-only/passwordless administration path is implemented but remains part of the remaining end-to-end validation work.

A public SSH key can also be supplied to the self-contained local ISO build used for development/testing. Only the public key is embedded; private SSH keys never belong in an installer ISO.

For SSH after installation:

- `ssh user@IP` works when the matching private key is available through `ssh-agent`, a normal default SSH identity, or SSH client configuration.
- If the private key has a custom filename and is not loaded into an agent, use `ssh -i /path/to/private-key user@IP`.
- After reinstalling a machine at the same IP, the client may need `ssh-keygen -R IP` because a fresh installation generates a new SSH host identity.

## Builder template

[Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder) is a separate GitHub template project for creating a personalized Home Server Installer ISO in the user's own GitHub account with GitHub Actions.

The user creates a repository from the template, adds an `SSH_PUBLIC_KEY` repository Actions secret, and runs the build workflow. The Builder automatically resolves the **latest published Home Server Installer release** and builds the ISO from that exact release.

Only the user's **public SSH key** is embedded. The private key stays on the user's own computer and is never uploaded to GitHub or embedded in the ISO.

The Builder does not embed one selected Gina/uCore image. The personalized installer keeps all five V1 image choices, and the selected image is downloaded during installation. **An internet connection is required during the normal installation path.**

## Upstream

This repository is derived from [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle) and retains the upstream Apache-2.0 license and project history.

Related projects:

- [Home Server Gina](https://github.com/home-server-project/home-server-gina)
- [Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder)
- [Universal Blue uCore](https://github.com/ublue-os/ucore)
- [Fedora CoreOS](https://fedoraproject.org/coreos/)

## Status

**V1 has a published release and is working in end-to-end VM testing.**

See the [latest published release](https://github.com/home-server-project/home-server-installer/releases/latest) for the current version and release assets.

The current direct-install path has successfully completed installation and first boot with both Home Server Project and upstream Universal Blue uCore targets. The personalized ISO path generated from a fresh Home Server Gina Builder template repository has also been validated.

VM testing is recommended first. Bare-metal testing can be done on a **dedicated test drive or test hardware where the selected installation disk can be safely erased**. Keep backups of any important data before testing.
