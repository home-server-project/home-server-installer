# Testing Home Server Installer V1

V1 is validated through end-to-end VM installation and dedicated bare-metal testing. VM testing is still recommended first before moving to physical hardware.

## Go tests

For the current Home Server-specific path, run:

```bash
GOTOOLCHAIN=auto go test ./internal/install ./internal/model ./internal/tui
```

Changed Go files should also be clean under `gofmt`.

## Build the installer binary

Example development build:

```bash
VERSION="$(git describe --tags --always)"
GOTOOLCHAIN=auto \
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -buildvcs=false \
  -ldflags="-s -w -X main.version=${VERSION}" \
  -o bin/knuckle-amd64 \
  ./cmd/knuckle
```

## Build the Fedora CoreOS live ISO

The self-contained development builder is `scripts/build-fcos-iso.sh`.

Example with an optional public SSH key:

```bash
./scripts/build-fcos-iso.sh \
  --stream stable \
  --arch amd64 \
  --binary ./bin/knuckle-amd64 \
  --ssh-key "$(cat ~/.ssh/id_ed25519.pub)"
```

Only a public SSH key belongs in the ISO. Never embed a private SSH key.

## VM validation

A clean VM test should verify the normal interactive path from ISO boot through first boot of the installed system.

Check at minimum:

- installer boots in UEFI mode;
- TUI appears normally;
- all 11 signed LTS installation choices are available through the four image families;
- every picker choice maps to the expected image tag;
- intended target disk is selected;
- 1 GiB or 2 GiB `/boot` layout matches the test plan;
- local password and/or SSH public key can be configured;
- signed image pull succeeds;
- direct installation completes;
- reboot enters the selected Gina/uCore image directly;
- SSH works when configured;
- `sudo bootc status` shows the expected image;
- `sudo systemctl --failed --no-pager` has no unexpected failures.

The four image families are:

- Home Server Gina LTS;
- Universal Blue uCore LTS;
- Universal Blue uCore LTS / NVIDIA Open;
- Universal Blue uCore LTS / NVIDIA LTS.

## Non-target disk test

When validating disk safety, attach a second test disk containing a recognizable sentinel partition/file before installation.

Install only to the selected target disk. After first boot, verify that the second disk has not been repartitioned or modified.

This test has already passed in clean VM validation and should remain part of release-level testing when disk logic changes.

## Image coverage

Current clean VM testing has validated:

- a Home Server Gina HCI LTS target;
- upstream `ghcr.io/ublue-os/ucore-minimal:lts` as a representative standard Universal Blue signing/install path;
- upstream `ghcr.io/ublue-os/ucore-hci:lts-nvidia-lts` as an NVIDIA LTS signing/install path.

The 2026-09-10 acceptance test exercised the complete two-level picker and verified that all 11 visible choices mapped to the intended image tags. A full install of **uCore HCI LTS NVIDIA LTS** completed successfully from signed pull through installation.

The installer exposes 11 signed LTS targets in total. Changes to image selection, signing trust, or installation logic should be tested against the affected trust path.

## SSH validation

If the ISO is built with `--ssh-key`, verify both stages when relevant:

1. SSH access to the live installer environment.
2. SSH access to the installed user after installation and reboot.

A matching private key loaded in `ssh-agent`, available under a normal default identity, or configured in the SSH client allows normal:

```bash
ssh <user>@<IP>
```

If the private key has a custom filename and is not loaded in an agent:

```bash
ssh -i ~/.ssh/<private-key> <user>@<IP>
```

When reinstalling a VM that receives the same IP, a new SSH host identity is expected. If the client reports a changed host key and you know this is the freshly reinstalled VM:

```bash
ssh-keygen -R <IP>
```

Then reconnect and verify the new fingerprint.

## Dedicated bare-metal testing

Bare-metal testing has completed successfully on dedicated test hardware using both direct-write USB media and Ventoy.

Validated examples include:

- Home Server Gina HCI LTS with a 1 GiB `/boot` layout;
- upstream Universal Blue uCore LTS with a 2 GiB `/boot` layout.

Additional hardware coverage remains useful, especially for NVIDIA systems. Use a dedicated test drive or test hardware where the selected installation disk can be erased.

Secure Boot must be disabled during the current V1 installation path. After installation, follow the current uCore documentation if Secure Boot is to be configured or enabled.
