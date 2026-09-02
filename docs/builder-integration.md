# Builder Integration Contract

Builder and Installer are separate repositories and separate releases.

The public Builder template should download a versioned Home Server Installer
release and embed it into the personalized Fedora CoreOS ISO. It should not
copy or fork Installer logic into each user's Builder repository.

## ISO payload layout

```text
/home-server/
├── installer/
│   ├── bin/home-server-installer
│   ├── lib/*.sh
│   ├── config/00-home-server.preset
│   └── VERSION
├── payload/
│   └── <selected-image>.oci.tar
└── config/
    ├── install.env
    └── authorized_keys.pub
```

The Installer currently discovers this tree at `/run/media/iso/home-server`.
`HSI_ISO_ROOT` or `--iso-root` exists for development/testing.

## `install.env` format version 1

Example:

```text
INSTALLER_FORMAT_VERSION=1
INSTALLER_VERSION=0.1.0-alpha.1
IMAGE_VARIANT=home-server-ucore-hci
IMAGE_REF=ghcr.io/home-server-project/home-server-ucore-hci:lts
IMAGE_DIGEST=sha256:<resolved-digest>
IMAGE_PINNED_REF=ghcr.io/home-server-project/home-server-ucore-hci@sha256:<resolved-digest>
OCI_ARCHIVE=/home-server/payload/home-server-ucore-hci.oci.tar
OCI_ARCHIVE_SHA256=<archive-sha256>
OCI_LOCAL_REF=localhost/home-server-ucore-hci:latest
SSH_PUBLIC_KEY=/home-server/config/authorized_keys.pub
PARTITION_POLICY=home-server-default-v1
UPDATE_POLICY=home-server-stage-v1
```

Installer parses this file as data. It never executes the file as shell.

### Required fields

- `INSTALLER_FORMAT_VERSION`
- `IMAGE_REF`
- `IMAGE_DIGEST`
- `OCI_ARCHIVE`
- `OCI_ARCHIVE_SHA256`
- `SSH_PUBLIC_KEY`

### Recommended fields

- `INSTALLER_VERSION`
- `IMAGE_VARIANT`
- `IMAGE_PINNED_REF`
- `OCI_LOCAL_REF`
- `PARTITION_POLICY`
- `UPDATE_POLICY`

`OCI_LOCAL_REF` should match the reference stored in the OCI archive. Installer
can parse Podman's `Loaded image:` output as a compatibility fallback, but
Builder should set it explicitly.

## Moving tag versus frozen payload

Builder resolves a moving reference such as:

```text
ghcr.io/home-server-project/home-server-ucore-hci:lts
```

to an immutable digest during the build. The OCI archive contains that exact
resolved image, while `IMAGE_REF` remains the moving reference bootc will use
for future installed-system updates.

## Signature boundary

Builder should verify the source registry image/signature policy before
creating the offline archive. The offline OCI archive may need signatures
removed for OCI archive compatibility; this does not replace the earlier
source verification.

Installer then checks:

1. archive SHA-256;
2. loaded image digest against `IMAGE_DIGEST`.

## SSH boundary

Builder embeds **public key only**. It must reject obvious private-key material
before creating the ISO; Installer validates again before destructive work.

## Partition-policy boundary

Builder may select a policy identifier. Installer owns:

- actual disk discovery;
- concrete partition sizes on that disk;
- proposal display;
- exact destructive confirmation;
- partition execution.

## Update-policy boundary

Known uCore/Home Server presets may use:

```text
UPDATE_POLICY=home-server-stage-v1
```

This is the policy proven in the disposable VM lab.

Until separately verified, vanilla Fedora CoreOS and arbitrary custom images
should use:

```text
UPDATE_POLICY=preserve
```

The Builder catalog should make that decision explicitly per preset instead of
forcing one policy onto every OCI image.

## Autostart

`systemd/home-server-installer.service` is included as a candidate live-media
unit, but automatic installation/enablement of that unit inside Fedora CoreOS
live media has **not yet been proven**. Builder should not claim autostart until
that integration is tested end-to-end.
