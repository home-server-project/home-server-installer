<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
<p align="center">
  <img src="https://raw.githubusercontent.com/home-server-project/.github/main/logo/banner-navy-mid.png" alt="Home Server Project banner">
</p>

# Home Server Installer

**Current release:** [![GitHub Release](https://img.shields.io/github/v/release/home-server-project/home-server-installer?label=latest)](https://github.com/home-server-project/home-server-installer/releases/latest)

Home Server Installer is a friendly Fedora CoreOS-based installer for [Home Server Gina](https://github.com/home-server-project/home-server-gina) and selected upstream [Universal Blue uCore](https://github.com/ublue-os/ucore) LTS images.

Home Server Installer is based on [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle), Project Bluefin's interactive TUI installer project. Home Server Project adapts that foundation for a Home Server-focused, signed, direct-install path while retaining Knuckle's TUI and hardware-discovery approach.

## Get the installer

There are two supported ways to get installation media.

### Official release ISO

[**Open the latest Home Server Installer release**](https://github.com/home-server-project/home-server-installer/releases/latest) and download the published amd64 ISO.

This is the simplest path. The release ISO is generic: boot it, choose the operating-system image, then configure the user, password and/or SSH access during installation.

Each Installer release publishes its own ISO, checksum and signed release artifacts. That ISO stays tied to the Installer release and Fedora CoreOS live environment used when that release was built. The selected Gina/uCore LTS image is still downloaded during installation, so the operating-system image itself is not embedded or frozen inside the ISO.

### Personalized ISO with Gina Builder

[Home Server Gina Builder](https://github.com/home-server-project/home-server-gina-builder) is optional. It creates a personalized ISO with your SSH public key already embedded, so you do not need to type or paste that key during installation.

The Builder uses the latest published Home Server Installer release, then builds fresh installation media from that released Installer code and the current Fedora CoreOS stable live image available when the Builder workflow runs.

<details>
<summary><strong>Which ISO should I use?</strong></summary>

- **Official release ISO** — easiest generic download; configure password or SSH access during installation.
- **Gina Builder ISO** — useful when you want your SSH public key already included and a freshly built Fedora CoreOS live environment.

</details>

## How it fits

- Boot either the official release ISO or a personalized Builder ISO.
- Home Server Installer collects the installation choices and downloads the selected signed Gina/uCore image.
- The selected image is installed directly as the first bootable deployment.

<details>
<summary><strong>Advanced/local build</strong></summary>

Advanced users can build Home Server Installer media locally instead of using the published ISO or Gina Builder.

A local build can use the current development source or a specific released tag, and the ISO can be created with or without an SSH public key already embedded. This is useful for development, testing, custom build workflows, or users who simply want full control over how their installation media is produced.

The local ISO build uses the Fedora CoreOS live image available for the selected stream at build time. The operating-system image is still selected and downloaded later during installation.

</details>

> [!WARNING]
> **Installation erases the selected target disk.**
>
> The current V1 path is UEFI-only. Secure Boot must be disabled during installation. The Installer erases and repartitions only the target disk selected in the TUI, so verify the selected disk before confirming installation and keep backups of anything important.

## Installer choices

Home Server Installer currently supports **11 signed LTS installation targets** across four families.

<details>
<summary><strong>Show all 11 installation targets</strong></summary>

### Home Server Gina LTS

- **Home Server Gina LTS** — `ghcr.io/home-server-project/home-server-gina:lts`
- **Home Server Gina HCI LTS** — `ghcr.io/home-server-project/home-server-gina-hci:lts`

### Universal Blue uCore LTS

- **uCore Minimal LTS** — `ghcr.io/ublue-os/ucore-minimal:lts`
- **uCore LTS** — `ghcr.io/ublue-os/ucore:lts`
- **uCore HCI LTS** — `ghcr.io/ublue-os/ucore-hci:lts`

### Universal Blue uCore LTS / NVIDIA Open

- **uCore Minimal LTS NVIDIA Open** — `ghcr.io/ublue-os/ucore-minimal:lts-nvidia`
- **uCore LTS NVIDIA Open** — `ghcr.io/ublue-os/ucore:lts-nvidia`
- **uCore HCI LTS NVIDIA Open** — `ghcr.io/ublue-os/ucore-hci:lts-nvidia`

### Universal Blue uCore LTS / NVIDIA LTS

- **uCore Minimal LTS NVIDIA LTS** — `ghcr.io/ublue-os/ucore-minimal:lts-nvidia-lts`
- **uCore LTS NVIDIA LTS** — `ghcr.io/ublue-os/ucore:lts-nvidia-lts`
- **uCore HCI LTS NVIDIA LTS** — `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts`

</details>

The TUI first selects one of the four image families, then the edition inside that family. All installer choices use LTS images. After installation, users can switch to `stable` or `testing` with `bootc`.

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

## Tested installation paths

Home Server Installer V1 has been successfully tested in virtual machines and on bare metal.

Tested boot and installation methods:

- Virtual machine
- Dedicated USB installer written directly to the drive, such as with Rufus, `dd`, or similar tools
- Ventoy USB

Bare-metal validation was completed on:

- ASUS VivoBook X412DA-AB31
- AMD Ryzen 3 3200U
- AMD Radeon Vega 3
- 12 GB DDR4
- 128 GB SSD
- USB Ethernet adapter using DHCP

Successful bare-metal installations included:

- **Home Server Gina HCI LTS** with a **1 GiB `/boot`**
- **Universal Blue uCore LTS** with a **2 GiB `/boot`**

Both installations completed successfully, rebooted into the selected image, and produced the expected disk layout.

The 2026-09-10 VM acceptance test also verified the new four-family image picker. All 11 installation choices mapped to the expected image tags, and a full **uCore HCI LTS NVIDIA LTS** installation completed successfully using `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts`.

Startup and installation time can vary depending on the USB drive, USB interface, network connection, and boot method.

## Storage and disk safety

V1 uses a direct Home Server storage layout on the selected target disk:

- 1 MiB BIOS boot partition
- 512 MiB EFI System Partition
- ext4 XBOOTLDR `/boot`
  - **1 GiB** — standard and recommended for Gina, Gina HCI and upstream uCore targets
  - **2 GiB** — larger layout for NVIDIA or custom-image use cases
- XFS `/` using the remaining disk

Both `/boot` layouts have now been successfully validated on bare metal: 1 GiB with Home Server Gina HCI LTS and 2 GiB with upstream uCore LTS.

The selected disk is revalidated immediately before destructive partitioning, including device identity, expected size, serial when available and mounted-filesystem checks.

VM testing has also verified that a separate attached non-target sentinel disk remains untouched during installation.

## User and SSH access

The Installer can create the primary user during installation and supports a local password, SSH authorized keys, or a key-only administration path.

With the **official generic ISO**, configure access during installation. You can set a password, paste an SSH public key, or enter a GitHub username so the Installer can fetch that account's published SSH public keys.

With a **personalized Gina Builder ISO**, the supplied public SSH key is already available to the Installer.

Password-backed users keep normal password-required `sudo` behavior. The SSH-key-only/passwordless administration path is available but remains part of ongoing end-to-end validation.

Only public SSH keys belong in installation media. Private SSH keys must remain on the user's own computer.

> [!NOTE]
> When the ISO was built with an SSH public key attached through Home Server Gina Builder, Fedora CoreOS may still display:
>
> `No SSH authorized keys provided by Ignition or Afterburn`
>
> This message can be ignored for this installation path. The SSH key configured in the personalized installer ISO is handled by Home Server Installer and is not supplied through Fedora CoreOS Ignition or Afterburn.

## Image verification

Home Server Project images are verified with the Home Server Project Cosign public key. Upstream `ublue-os/ucore*` images, including NVIDIA Open and NVIDIA LTS variants, are verified with Universal Blue's uCore Cosign public key.

Only supported repositories are accepted by the Home Server installation path. Unknown image repositories are rejected. Temporary signature-discovery configuration used during installation is cleaned up and does not persist into the installed system.

## Current scope

V1 currently focuses on the signed Gina/uCore LTS installation path for x86_64 / amd64 UEFI systems, with direct bootc installation and the storage, user and SSH configuration described above.

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
