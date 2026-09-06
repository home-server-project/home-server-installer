# Home Server Installer

> [!WARNING]
> **In development. Destructive installer. Use only in disposable VMs/test hardware for now.**

Home Server Installer is a friendly Fedora CoreOS-based installer for [Home Server uCore](https://github.com/home-server-project/home-server-ucore).

It is a thin downstream adaptation of [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle). The goal is to keep Knuckle's proven TUI, hardware discovery and Fedora CoreOS installer backend, while adding the Home Server-specific image selection and first-boot uCore transition.

## V1 targets

- `ghcr.io/home-server-project/home-server-ucore:lts`
- `ghcr.io/home-server-project/home-server-ucore-hci:lts`

Upstream uCore image choices can be added later without redesigning the installer.

## V1 flow

```text
Fedora CoreOS live ISO + Home Server Installer
                 |
                 v
      choose uCore / uCore HCI
      network / disk / user / SSH
                 |
                 v
        coreos-installer installs FCOS
                 |
                 v
  first boot verifies the Home Server signature
  and rebases directly to the selected signed image
                 |
                 v
               reboot
                 |
                 v
          Home Server uCore
```

The Home Server path intentionally keeps the temporary Fedora CoreOS configuration minimal. Generic Knuckle Flatcar/FCOS functionality remains in the codebase so upstream changes can continue to be merged, but it is not exposed as the primary Home Server V1 flow.

## Known V1 limitation: disk layout

V1 deliberately uses the standard Fedora CoreOS disk layout so the new installer/rebase mechanics can be proven first.

The stock FCOS `/boot` size is not the final Home Server design. A later milestone will add a larger `/boot` option (1 GiB / 2 GiB) required for the Home Server/uCore update workflow.

## Builder

The future Builder template will create the personalized installer ISO and can add small per-user settings such as an SSH public key. Private SSH keys never belong in an ISO.

## Upstream

This repository is derived from [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle) and retains the upstream Apache-2.0 license and project history.

Related projects:

- [Home Server uCore](https://github.com/home-server-project/home-server-ucore)
- [Universal Blue uCore](https://github.com/ublue-os/ucore)
- [Fedora CoreOS](https://fedoraproject.org/coreos/)

## Status

**V1 is development software and is not ready for installation on a real home server.**

The next milestone is a complete end-to-end disposable-VM installation test of the V1 flow.