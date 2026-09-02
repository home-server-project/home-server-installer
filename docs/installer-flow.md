# Installer Flow

The alpha flow is intentionally linear and inspectable.

```text
start
  -> locate /home-server payload
  -> parse install.env as data (never source/eval)
  -> validate OCI archive SHA-256
  -> validate SSH public key
  -> discover hardware-backed disks
  -> exclude installation media
  -> user selects target
  -> print exact disk + partition proposal
  -> require: ERASE /dev/<target>
  -> wipe/repartition/format target
  -> mount root
  -> mount /boot
  -> create /boot/efi after /boot is mounted
  -> mount ESP
  -> privileged blkid root + boot UUID capture
  -> hard abort if either UUID is empty
  -> create target-backed Podman/TMP workspace
  -> load embedded OCI archive
  -> verify loaded image digest
  -> privileged target-image container
  -> bootc install to-filesystem --skip-finalize
  -> verify BLS UUIDs immediately
  -> discover actual OSTree deployment root
  -> apply selected update policy
  -> verify offline policy
  -> remove temporary Podman/TMP workspace
  -> bootc install finalize
  -> verify BLS/origin/SSH/policy/boot artifacts
  -> write installation report
  -> sync
  -> unmount target
  -> recommend poweroff so installation media can be removed
```

## Why UUID validation occurs twice

A lab failure proved that an unprivileged `blkid` invocation can return an
empty value and that bootc may still accept `--root-mount-spec UUID=` and
`--boot-mount-spec UUID=`.

The installer therefore validates UUIDs before bootc and verifies the generated
BLS entry after bootc.

## Why final poweroff is recommended

On physical hardware the installation USB may still have firmware boot
priority. Powering off gives the user a clean opportunity to remove the media
before first boot of the installed system.
