<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
# Release Checklist

> **V1 note:** Home Server Installer V1 is validated through end-to-end VM installation and dedicated bare-metal testing. Before publishing a release, verify the current branch in a clean VM. Physical testing should remain limited to a dedicated test drive or test hardware where the selected installation disk can be safely erased.

## Pre-release checks

- [ ] `go test ./internal/install ./internal/model ./internal/tui` passes.
- [ ] `gofmt` is clean for changed Go files.
- [ ] The installer ISO builds successfully from the current branch.
- [ ] The ISO boots in UEFI mode and the Home Server Installer TUI appears.
- [ ] All 11 signed LTS installation choices are displayed correctly through the four image families.
- [ ] Each picker choice maps to the expected image tag.
- [ ] At least one Home Server Project Gina target completes signed pull, direct install, reboot, and first boot.
- [ ] At least one standard upstream Universal Blue uCore target completes signed pull, direct install, reboot, and first boot.
- [ ] At least one affected NVIDIA trust path is validated when NVIDIA image support changes.
- [ ] The selected target disk and `/boot` layout shown in the review screen match the intended test configuration.
- [ ] A separate attached non-target test disk remains unchanged.
- [ ] Local-password provisioning works when selected.
- [ ] SSH public-key provisioning works after installation.
- [ ] `sudo bootc status` reports the expected installed image.
- [ ] `sudo systemctl --failed --no-pager` reports no unexpected failed units.
- [ ] Installer-only files/services do not remain in the installed system.
- [ ] `README.md`, `CHANGELOG.md`, `docs/TESTING.md`, and this release checklist describe the current behavior accurately.

## V1 installation expectations

The current V1 path:

1. Boots a Fedora CoreOS live installer environment.
2. Lets the user select one of four signed LTS image families and then an edition within that family.
3. Supports 11 signed installation targets across Home Server Gina, standard upstream uCore, NVIDIA Open, and NVIDIA LTS families.
4. Lets the user select the target disk and 1 GiB or 2 GiB `/boot` layout.
5. Collects primary-user, local-password, and SSH-key configuration.
6. Verifies and pulls the selected signed image.
7. Partitions only the selected target disk.
8. Installs the selected image directly with `bootc install to-filesystem`.
9. Reboots directly into the selected Gina/uCore image.

There is no installed Fedora CoreOS intermediate and no first-boot autorebase step.

## Current validation status

The 2026-09-10 VM acceptance test verified the complete two-level image picker and confirmed that all 11 visible choices map to the intended image tags.

A full signed installation of **uCore HCI LTS NVIDIA LTS** using `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts` completed successfully.

Existing dedicated bare-metal validation includes:

- Home Server Gina HCI LTS with a 1 GiB `/boot` layout;
- upstream Universal Blue uCore LTS with a 2 GiB `/boot` layout;
- direct-write USB installation media;
- Ventoy installation media.

## Secure Boot

Secure Boot must be disabled during the current V1 installation path. After installation, follow the current [uCore documentation](https://github.com/ublue-os/ucore) if Secure Boot is to be configured or enabled.

## Release notes

For V1 releases, clearly state:

- which Home Server Installer commit/tag was built;
- Fedora CoreOS live ISO version used by the builder;
- which Gina, standard uCore, and NVIDIA targets were tested end to end;
- whether the release has been tested in VMs, on dedicated bare-metal test hardware, or both;
- any known installation limitations, including the current cosmetic progress-bar behavior when relevant.

The inherited Project Bluefin Knuckle release process is not the release process for this fork. Upstream release history remains available at [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle).
