# Home Server Installer

> **Alpha software. Destructive by design. Test on disposable hardware/VMs first.**

Home Server Installer is the installation engine for the Home Server project.
It runs from a personalized Fedora CoreOS live ISO, validates the embedded
payload, asks the user to choose a physical target disk, shows the exact disk
layout, requires an exact destructive confirmation, and installs the selected
bootc image directly to disk.

The project boundary is intentional:

> **Builder creates the installation media. Installer creates the machine. Deployer creates and manages the server environment.**

Current priority is **Builder + Installer**. Deployer is a later project.

## Current alpha status

The low-level installation path represented by this repository has been proven
manually in the disposable `home-server-installer-lab` VM through:

- embedded/offline OCI payload;
- target-backed Podman storage;
- direct `bootc install to-filesystem`;
- SSH public-key injection;
- mandatory root and boot UUID validation;
- correct BLS root/boot UUIDs;
- OSTree deployment-root discovery;
- Zincati masking;
- persistent Home Server systemd preset;
- `rpm-ostreed-automatic.timer` enabled and active after reboot;
- `AutomaticUpdatePolicy=stage` for the tested uCore image;
- `bootc install finalize`;
- first boot, SSH login, reboot, and zero failed systemd units.

This repository converts that proven manual procedure into an alpha installer
engine. The repository itself still needs end-to-end testing as a packaged
Installer release before it should be considered generally usable.

## Verified platform for v0.1 alpha

- x86_64
- UEFI
- Secure Boot **not yet part of the verified path**
- Fedora CoreOS live environment
- embedded OCI archive
- bootc-capable target image
- default Home Server partition policy

## Default disk layout

| Partition | Size | Type | Filesystem / role |
|---|---:|---|---|
| 1 | 1 MiB | EF02 | BIOS boot compatibility |
| 2 | 512 MiB | EF00 | FAT32 EFI System Partition |
| 3 | 2 GiB | EA00 | ext4 `/boot` / XBOOTLDR |
| 4 | remainder | 8304 | XFS root |

The 2 GiB `/boot` is deliberate. It avoids the too-small boot-partition problem
seen on an older real uCore server during update staging.

## Safety model

The installer does not silently choose a target disk.

It:

1. detects hardware-backed whole disks;
2. excludes zram, loop devices, device-mapper/md pseudo-disks, optical drives,
   and the detected installation-media disk;
3. displays device, model, size, transport, class, and serial;
4. shows the full proposed partition layout;
5. names the exact target again;
6. requires the user to type exactly `ERASE /dev/<target>`;
7. refuses to call bootc if the root or boot filesystem UUID is empty;
8. verifies the generated BLS UUID arguments before finalization.

There is intentionally no `--yes` or unattended erase switch in this alpha.

## Installation-media contract

The Builder embeds this repository/release under `/home-server/installer` and
provides:

```text
/home-server/
├── installer/
│   ├── bin/
│   ├── lib/
│   ├── config/
│   └── VERSION
├── payload/
│   └── <image>.oci.tar
└── config/
    ├── install.env
    └── authorized_keys.pub
```

See [`docs/builder-integration.md`](docs/builder-integration.md) for the exact
metadata contract.

## Manual alpha launch

Until Builder autostart wiring is separately proven, the canonical manual
entry point from the Fedora CoreOS live environment is:

```bash
sudo /run/media/iso/home-server/installer/bin/home-server-installer
```

Development dry-run:

```bash
sudo /run/media/iso/home-server/installer/bin/home-server-installer \
  --dry-run \
  --target /dev/vda
```

Dry-run validates the metadata, SSH public key, OCI archive checksum, disk
discovery, and proposed partition plan without changing the disk.

## Update policy

For the proven uCore/Home Server path, Builder metadata selects:

```text
UPDATE_POLICY=home-server-stage-v1
```

Installer then requires the selected image to already contain:

```text
AutomaticUpdatePolicy=stage
```

and applies the tested Installer-owned policy:

- mask `zincati.service`;
- install `/etc/systemd/system-preset/00-home-server.preset`;
- keep Zincati disabled through first-boot preset processing;
- enable `rpm-ostreed-automatic.timer` through first-boot preset processing.

The preset contains:

```text
disable zincati.service
enable rpm-ostreed-automatic.timer
```

The installer intentionally does **not** rewrite an unknown
`rpm-ostreed.conf` in v0.1. If `AutomaticUpdatePolicy=stage` is not already
present, the verified Home Server policy aborts instead of guessing.

For an image whose update mechanism should not be changed, Builder can set:

```text
UPDATE_POLICY=preserve
```

Custom and vanilla Fedora CoreOS image presets should use `preserve` until
their desired update policy is separately tested.

## SSH key handling

Only a public key is embedded. Private keys never belong in the ISO.

The installer rejects files containing private-key headers and validates the
public key with `ssh-keygen` before destructive work begins.

The future public Builder template is expected to obtain the public key from a
GitHub Actions repository secret or another GitHub-native UX. The Builder
README will provide exact Windows, macOS, and Linux instructions.

## Tests

Run non-destructive unit tests:

```bash
bash tests/run.sh
```

Syntax check:

```bash
bash -n bin/home-server-installer lib/*.sh tests/*.sh
```

GitHub Actions also runs ShellCheck.

## Releases

`VERSION` is the canonical Installer version. Pushing a matching tag such as:

```text
v0.1-alpha
```

runs the release workflow, tests the repository, creates `.tar.gz` and `.zip`
release archives plus `SHA256SUMS`, and publishes a GitHub prerelease.

The future Builder should consume a versioned Installer release rather than
copying Installer logic into each Builder template repository.

## Not yet claimed by this alpha

- Secure Boot support
- BIOS-only firmware support
- advanced/custom partition editor
- RAID/storage-pool creation
- automated live-ISO autostart wiring
- broad custom-image compatibility
- unattended destructive installs
- production support guarantees

Those are separate milestones. The alpha intentionally keeps the verified path
small and explicit.

## License

Apache-2.0.
