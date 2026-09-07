# Release Checklist

> **V1 note:** Home Server Installer V1 is currently validated through end-to-end VM installation. Before publishing a V1 release intended for broader testing, verify the current branch in a clean VM and keep bare-metal validation limited to a dedicated test drive or test hardware where the selected installation disk can be safely erased.

## Pre-release checks

- [ ] `go test ./internal/install ./internal/model ./internal/tui` passes.
- [ ] `gofmt` is clean for changed Go files.
- [ ] The installer ISO builds successfully from the current branch.
- [ ] The ISO boots in UEFI mode and the Home Server Installer TUI appears.
- [ ] All five signed LTS image choices are displayed correctly.
- [ ] At least one Home Server Project target completes signed pull, direct install, reboot, and first boot.
- [ ] At least one upstream Universal Blue uCore target completes signed pull, direct install, reboot, and first boot.
- [ ] The selected target disk and `/boot` layout shown in the review screen match the intended test configuration.
- [ ] A separate attached non-target test disk remains unchanged.
- [ ] Local-password provisioning works when selected.
- [ ] SSH public-key provisioning works after installation.
- [ ] `sudo bootc status` reports the expected installed image.
- [ ] `sudo systemctl --failed --no-pager` reports no unexpected failed units.
- [ ] Installer-only files/services do not remain in the installed system.
- [ ] `README.md`, `CHANGELOG.md`, and this release checklist describe the current behavior accurately.

## V1 installation expectations

The current V1 path:

1. Boots a Fedora CoreOS live installer environment.
2. Lets the user select a supported signed uCore image.
3. Lets the user select the target disk and 1 GiB or 2 GiB `/boot` layout.
4. Collects primary-user, local-password, and SSH-key configuration.
5. Verifies and pulls the selected signed image.
6. Partitions only the selected target disk.
7. Installs the selected image directly with `bootc install to-filesystem`.
8. Reboots directly into the selected uCore image.

There is no installed Fedora CoreOS intermediate and no first-boot autorebase step.

## Secure Boot

Secure Boot must be disabled during the current V1 installation path. After installation, follow the current [uCore documentation](https://github.com/ublue-os/ucore) if Secure Boot is to be configured or enabled.

## Release notes

For V1 releases, clearly state:

- which Home Server Installer commit/tag was built;
- Fedora CoreOS live ISO version used by the builder;
- which uCore targets were tested end to end;
- whether the release has been tested only in VMs or also on dedicated bare-metal test hardware;
- any known installation limitations.

The inherited Project Bluefin Knuckle release process is not the release process for this fork. Upstream release history remains available at [Project Bluefin Knuckle](https://github.com/projectbluefin/knuckle).
