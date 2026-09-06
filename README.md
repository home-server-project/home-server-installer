# Home Server Installer

> [!WARNING]
> **In development. Destructive installer. Use only in disposable VMs/test hardware for now.**
>
> - **UEFI only** for the current V1 path.
> - **The selected target disk is erased.**
> - V1 installs to a **single target disk**. Multi-disk/NAS hardware has not been cleared for real-machine use yet.
> - **Secure Boot is not a supported V1 real-hardware path.** Keep it disabled until the MOK flow is implemented and tested.

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

### Signed first rebase

V1 injects the Home Server Cosign public key and a temporary container signature policy into the installed Fedora CoreOS system. That policy allows only the two signed Home Server image repositories for the first rebase.

The rebase service retries transient network/registry failures with backoff and leaves a visible MOTD/journal breadcrumb if all attempts fail. After a successful signed rebase is staged, the temporary FCOS trust policy is restored to the original FCOS default before reboot. A one-shot cleanup on the first Home Server uCore boot then restores the image-provided container policy and removes the bootstrap trust files/services.

## Known V1 limitation: disk layout

V1 deliberately uses the standard Fedora CoreOS disk layout so the new installer/rebase mechanics can be proven first.

The stock FCOS `/boot` size is not the final Home Server design. A later milestone will add a larger `/boot` option (1 GiB / 2 GiB) required for the Home Server/uCore update workflow.

**Changing to that larger `/boot` layout will require reinstalling the machine. A normal uCore/bootc update cannot resize the existing V1 partition layout.**

## Disk safety scope

The installer passes only the explicitly selected target disk to `coreos-installer`, and the review screen shows the selected disk before the destructive confirmation.

Before real hardware is in scope, the Home Server VM test suite must also prove that an attached non-target data disk remains unchanged during installation.

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
