# Installer Architecture

## Project boundary

```text
Builder -> personalized ISO -> Installer -> installed machine -> Deployer
```

Builder and Installer are separate release units.

Builder owns acquisition, resolution, verification, and packaging of input
artifacts. Installer owns the destructive transition from a booted live ISO to
an installed machine. Deployer later owns applications and ongoing server
configuration.

## Installer layers

```text
bin/home-server-installer
        |
        +-- common.sh
        +-- metadata.sh
        +-- ssh.sh
        +-- disks.sh
        +-- partitions.sh
        +-- image.sh
        +-- bootc-install.sh
        +-- update-policy.sh
        +-- verify.sh
        `-- report.sh
```

The shell modules deliberately keep presentation, disk handling, image staging,
bootc deployment, policy, and verification separate. This makes it possible to
replace the interactive frontend later without rewriting the proven install
engine.

## Why Fedora CoreOS live media

Fedora CoreOS is the temporary live installer environment. It is not installed
as an intermediate operating system and then rebased.

The selected bootc image is installed directly to the target disk.

## Why target-backed Podman storage

The OCI archive is multi-gigabyte and expands further when loaded. The verified
VM test kept live `/var` near 167 MiB by placing both Podman storage and TMPDIR
on the target root filesystem.

The temporary directories are self-bind-mounted because bootc's empty-root
safety check accepts mount points but must not see ordinary installer data in
the target filesystem.

The temporary Podman storage is removed before `bootc install finalize`.

## OSTree deployment root

After `bootc install to-filesystem --skip-finalize`, the mount point is an
OSTree sysroot. The deployed system root is discovered under:

```text
<target>/ostree/deploy/<os>/deploy/<deployment>.0
```

Installer-owned `/etc` and systemd policy changes must target that deployed
root.

## Failure philosophy

Installer favors a hard stop over guessing.

Examples:

- no eligible physical disk -> stop;
- target is live media -> stop;
- mounted target -> stop;
- invalid SSH key -> stop;
- payload checksum mismatch -> stop;
- loaded OCI digest mismatch -> stop;
- empty filesystem UUID -> stop before bootc;
- unexpected update-policy prerequisites -> stop;
- BLS UUID mismatch -> stop before declaring success.
