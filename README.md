# Home Server Installer

> [!WARNING]
> **In development. This installer is destructive and testing-first.**
>
> - **VM testing is preferred** while V1 is still being validated.
> - Bare-metal testing should be limited to **dedicated test hardware with no important data**.
> - **Do not use it on a real home server or production server yet.** The selected target disk is erased.
> - **UEFI only** for the current V1 path.
> - V1 installs to a **single target disk**.
> - **Secure Boot is not a supported V1 real-hardware path** yet.

Home Server Installer is a friendly Fedora CoreOS-based installer for [Home Server uCore](https://github.com/home-server-project/home-server-ucore) and selected upstream [Universal Blue uCore](https://github.com/ublue-os/ucore) images.

It is a thin downstream adaptation of [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle). The goal is to keep Knuckle's proven TUI and hardware discovery while providing a Home Server-focused, signed, direct-install path for uCore.

## V1 image choices

The installer currently exposes five signed LTS targets:

- `ghcr.io/home-server-project/home-server-ucore:lts` — Home Server uCore LTS (default/recommended).
- `ghcr.io/home-server-project/home-server-ucore-hci:lts` — Home Server uCore HCI LTS.
- `ghcr.io/ublue-os/ucore-minimal:lts` — upstream Universal Blue uCore Minimal LTS.
- `ghcr.io/ublue-os/ucore:lts` — upstream Universal Blue uCore LTS.
- `ghcr.io/ublue-os/ucore-hci:lts` — upstream Universal Blue uCore HCI LTS.

The installer intentionally keeps the menu to LTS targets. Stable, NVIDIA and custom images can remain post-install `bootc switch` destinations instead of turning the installer into a large image matrix.

## V1 flow

```text
Fedora CoreOS live ISO + Home Server Installer
                 |
                 v
      choose signed uCore image
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
       selected uCore image boots
```

The selected uCore image is installed directly as the first bootable deployment. There is no installed Fedora CoreOS intermediate and no first-boot autorebase step.

## Signed image installation

Home Server Project images are verified with the Home Server Project Cosign public key. Upstream `ublue-os/ucore*` images are verified with Universal Blue's uCore Cosign public key.

Only supported repositories are accepted by the Home Server path. Unknown image repositories are rejected. Temporary signature-discovery configuration used during installation is cleaned up and does not persist into the installed system.

## Storage layout

V1 uses a direct Home Server storage layout:

- 1 MiB BIOS boot partition.
- 512 MiB EFI System Partition.
- ext4 XBOOTLDR `/boot` with one of two presets:
  - **1 GiB** — standard and recommended for uCore / uCore HCI.
  - **2 GiB** — large layout for NVIDIA or custom-image use cases.
- XFS `/` using the remaining disk.

The selected target disk is revalidated immediately before destructive partitioning, including device identity, expected size, serial when available, and mounted-filesystem checks.

Clean VM testing has also verified that a separate attached non-target sentinel disk remains untouched during installation. Real-hardware testing is still intentionally limited to dedicated test systems.

## User and SSH access

The installer can create the primary user during installation and supports a **local password**, SSH authorized keys, or a key-only administration path.

Password-backed users keep normal password-required `sudo` behavior. The SSH-key-only/passwordless administration path is implemented but remains part of the remaining end-to-end validation work.

A public SSH key can also be supplied to the self-contained local ISO build used for development/testing. Only the public key is embedded; private SSH keys never belong in an installer ISO.

For SSH after installation:

- `ssh user@IP` works when the matching private key is available through `ssh-agent`, a normal default SSH identity, or SSH client configuration.
- If the private key has a custom filename and is not loaded into an agent, use `ssh -i /path/to/private-key user@IP`.
- After reinstalling a machine at the same IP, the client may need `ssh-keygen -R IP` because a fresh installation generates a new SSH host identity.

## Builder template

The planned [Home Server uCore Builder](https://github.com/home-server-project/home-server-ucore-builder) is a separate GitHub template project. Its purpose is to let a user create a personalized installer ISO in their own GitHub account using GitHub Actions.

The intended Builder flow is to import the user's **public SSH key** through GitHub Secrets and inject that public key into the generated ISO. The private key stays on the user's own computer and is never uploaded to GitHub or embedded in the ISO.

## Upstream

This repository is derived from [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle) and retains the upstream Apache-2.0 license and project history.

Related projects:

- [Home Server uCore](https://github.com/home-server-project/home-server-ucore)
- [Home Server uCore Builder](https://github.com/home-server-project/home-server-ucore-builder)
- [Universal Blue uCore](https://github.com/ublue-os/ucore)
- [Fedora CoreOS](https://fedoraproject.org/coreos/)

## Status

**V1 is working in disposable VM testing but is not ready for a real home server or production server.**

The current direct-install path has completed end-to-end VM installation and first boot with both a Home Server Project target and an upstream Universal Blue uCore target. Dedicated test-hardware validation can follow after the VM path is considered stable enough; production/home-server use remains out of scope for now.
