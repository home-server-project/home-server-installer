# Disk Layout

## `home-server-default-v1`

The first alpha intentionally exposes one verified layout:

```text
GPT
├─ p1   1 MiB       EF02   BIOS boot compatibility
├─ p2   512 MiB     EF00   FAT32   EFI-SYSTEM
├─ p3   2 GiB       EA00   ext4    boot
└─ p4   remainder   8304   XFS     root
```

The layout is generated on the Installer side because Builder does not know the
actual physical target disk.

Builder may later select a `PARTITION_POLICY` identifier, but Installer always
shows the concrete physical-disk proposal before erasing anything.

## Mount order

Correct order:

```text
mount root
  -> create /boot
  -> mount /boot
  -> create /boot/efi
  -> mount ESP
```

Creating `/boot/efi` before mounting `/boot` is incorrect because the later
`/boot` mount hides the previously created directory.

## Future layouts

Possible later policies include a simpler CoreOS layout, separate data area,
or advanced/custom partitioning. They are intentionally not implemented until
the safety UX and test coverage are ready.
