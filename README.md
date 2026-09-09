<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
<p align="center">
  <img src="https://raw.githubusercontent.com/home-server-project/.github/main/logo/banner-navy-mid.png" alt="Home Server Project banner">
</p>

# Home Server Installer

**Current release:** [![GitHub Release](https://img.shields.io/github/v/release/home-server-project/home-server-installer?label=latest)](https://github.com/home-server-project/home-server-installer/releases/latest)

Home Server Installer is a friendly Fedora CoreOS-based installer for [Home Server Gina](https://github.com/home-server-project/home-server-gina) and selected upstream [Universal Blue uCore](https://github.com/ublue-os/ucore) LTS images.

Home Server Installer is based on [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle), Project Bluefin's interactive TUI installer project. Home Server Project adapts that foundation for a Home Server-focused, signed, direct-install path while retaining Knuckle's TUI and hardware-discovery approach.

## How it fits

- **Home Server Gina Builder** creates the personalized bootable ISO for the normal user path.
- **Home Server Installer** runs from that ISO and handles image selection, disk selection, partitioning, user setup and installation.
- **Home Server Gina or upstream uCore** is the operating-system image selected and installed by the user.

For most users, the recommended starting point is [Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder).

<details>
<summary><strong>Advanced/local use</strong></summary>

The Installer can also be built and used directly from this repository for local development, testing and advanced workflows without using the Builder template.

That path is intended for users who want to work with the Installer source itself rather than simply create a personalized ISO.

</details>

> [!WARNING]
> **VM testing is recommended first.**
>
> The current V1 path is UEFI-only. Secure Boot must be disabled during installation. The Installer erases and repartitions only the target disk selected in the TUI, so use a dedicated test drive or hardware where that disk can be safely erased and keep backups of anything important.

## Installer choices

The Installer exposes five signed LTS targets:

- **Home Server Gina LTS** — `ghcr.io/home-server-project/home-server-gina:lts`
- **Home Server Gina HCI LTS** — `ghcr.io/home-server-project/home-server-gina-hci:lts`
- **uCore Minimal LTS** — `ghcr.io/ublue-os/ucore-minimal:lts`
- **uCore LTS** — `ghcr.io/ublue-os/ucore:lts`
- **uCore HCI LTS** — `ghcr.io/ublue-os/ucore-hci:lts`

The menu intentionally stays limited to LTS targets. Other images can remain post-install `bootc switch` destinations instead of turning the Installer into a large image matrix.

## Installation flow

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

## What to expect when installing

> [!NOTE]
> **The Installer may take a few minutes to appear after booting.** Fedora CoreOS starts in the background before the Home Server Installer UI launches.
>
> During installation, the progress bar may remain around **20% for several minutes** while the selected image is downloaded, verified and deployed.
>
> **This is expected. Do not power off or reboot the machine while installation is in progress.**

## Storage and disk safety

V1 uses a direct Home Server storage layout on the selected target disk:

- 1 MiB BIOS boot partition
- 512 MiB EFI System Partition
- ext4 XBOOTLDR `/boot`
  - **1 GiB** — standard and recommended for Gina, Gina HCI and upstream uCore targets
  - **2 GiB** — larger layout for NVIDIA or custom-image use cases
- XFS `/` using the remaining disk

The selected disk is revalidated immediately before destructive partitioning, including device identity, expected size, serial when available and mounted-filesystem checks.

VM testing has also verified that a separate attached non-target sentinel disk remains untouched during installation.

## User and SSH access

The Installer can create the primary user during installation and supports a local password, SSH authorized keys, or a key-only administration path.

Password-backed users keep normal password-required `sudo` behavior. The SSH-key-only/passwordless administration path is available but remains part of ongoing end-to-end validation.

A public SSH key can be supplied when building a personalized ISO. Only the public key belongs in the ISO; private SSH keys must remain on the user's own computer.

For the normal personalized-ISO workflow, use [Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder), which handles the public-key handoff without requiring changes to Installer source.

## Image verification

Home Server Project images are verified with the Home Server Project Cosign public key. Upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore Cosign public key.

Only supported repositories are accepted by the Home Server installation path. Unknown image repositories are rejected. Temporary signature-discovery configuration used during installation is cleaned up and does not persist into the installed system.

## Current scope

The current V1 path includes:

- target-disk selection
- 1 GiB / 2 GiB `/boot` layout selection
- local password setup
- SSH key configuration
- signed Gina and upstream uCore LTS image selection
- direct bootc installation to the selected disk

VM testing remains the recommended first step before controlled bare-metal use.

## Upstream and references

<details>
<summary><strong>Project and upstream links</strong></summary>

- [Home Server Gina](https://github.com/home-server-project/home-server-gina)
- [Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder)
- [Home Server Project](https://github.com/home-server-project)
- [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle)
- [Universal Blue uCore](https://github.com/ublue-os/ucore)
- [Fedora CoreOS](https://fedoraproject.org/coreos/)

</details>

## License

Apache-2.0. Home Server Installer is a modified derivative of Project Bluefin Knuckle and retains the applicable upstream license, project history, attribution and modification notices. Third-party software and upstream components retain their own licenses.
