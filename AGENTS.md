<!-- Modified by Home Server Project from Project Bluefin Knuckle. -->
# Home Server Installer — Agent Guide

This repository is the Home Server Project installer for signed uCore images. It is derived from Project Bluefin Knuckle, but the active product direction is the Home Server Installer V1 described in `README.md`.

## Core rules

- Treat the selected installation disk as the only disk that may be partitioned or modified.
- Preserve target-disk revalidation immediately before partitioning.
- Do not weaken signed-image verification or broaden trusted repositories without an explicit design decision.
- Home Server Project images and upstream Universal Blue uCore images use different signing trust; preserve that separation.
- Never place private SSH keys in source, logs, GitHub Secrets for the public-key builder flow, or installer ISOs. Only public SSH keys are passed into the installer.
- Keep installer-only files, services, temporary trust configuration, and live SSH handoff state out of the installed system.
- Secure Boot must remain disabled for the current V1 installation path; post-install Secure Boot guidance belongs to uCore.

## Current V1 scope

The interactive installer currently exposes five signed LTS targets:

- Home Server uCore LTS
- Home Server uCore HCI LTS
- upstream uCore Minimal LTS
- upstream uCore LTS
- upstream uCore HCI LTS

The user selects the target disk and a 1 GiB or 2 GiB XBOOTLDR `/boot` layout. All installer-created partitions are placed on that selected drive. The selected image is installed directly with `bootc install to-filesystem`; there is no installed Fedora CoreOS intermediate or first-boot autorebase step.

The installer supports a primary user, local password, and SSH authorized-key provisioning.

## Validation before claiming a change is ready

Run the tests relevant to the changed code. For the current Home Server-specific path, the minimum known-good package set is:

```bash
GOTOOLCHAIN=auto go test ./internal/install ./internal/model ./internal/tui
```

Changed Go files must be `gofmt` clean.

Disk, image-signing, user-provisioning, SSH, ISO-boot, or installation-path changes require VM validation. See `docs/TESTING.md` and `docs/RELEASE.md`.

When disk logic changes, include a non-target sentinel disk in VM testing and verify it remains unchanged.

## Documentation

Keep these files aligned with current behavior:

- `README.md` — product overview and current V1 behavior
- `CHANGELOG.md` — Home Server fork changes and verified milestones
- `docs/TESTING.md` — current testing workflow
- `docs/TROUBLESHOOTING.md` — user/test troubleshooting
- `docs/SECURITY.md` — current trust and safety model
- `docs/RELEASE.md` — release validation checklist

Historical Project Bluefin Knuckle agent workflows and Flatcar-specific documentation are not authoritative for this fork. Use the current code and the Home Server documentation above.