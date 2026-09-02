# Safety Model

The Installer is allowed to destroy a disk, so disk identity is treated as a
first-class safety boundary.

## Disk discovery

The alpha starts from `lsblk` whole-disk records and additionally requires a
hardware-backed sysfs `device` node. This filters pseudo-disks such as zram
which may still report `TYPE=disk`.

Explicitly rejected classes include:

- loop devices;
- zram/ram devices;
- optical drives;
- device-mapper and md pseudo-disks;
- the physical parent of the mounted installation media.

Removable/external disks are not automatically rejected because a user may
intentionally install to an external SSD. They are visibly marked.

## Mounted-disk guard

A requested target with active mount points is rejected before destructive
work.

## Exact erase confirmation

The final prompt repeats the full target path and requires:

```text
ERASE /dev/<exact-target>
```

No `--yes`, wildcard, or shortened confirmation exists in the alpha.

## Payload safety

- `install.env` is parsed as key/value data and is never sourced as shell.
- unknown metadata keys are rejected;
- archive SHA-256 is verified before target modification;
- public SSH key is validated before target modification;
- private-key headers are rejected;
- the loaded OCI image digest must match `IMAGE_DIGEST`.

Source-image signature verification belongs to Builder before it creates the
offline OCI archive. OCI archive format cannot carry every registry signature
format, so Installer verifies the frozen artifact checksum and expected image
digest.

## UUID gate

Root and boot UUIDs are mandatory non-empty values before bootc runs. The BLS
entry is then checked for both UUIDs.

## Failure state

On an error after target mounting, the installer cleans its transient Podman
mounts but intentionally leaves the target filesystem mounted for diagnosis.
It does not reboot or pretend the install succeeded.
