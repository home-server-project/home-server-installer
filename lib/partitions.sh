#!/usr/bin/env bash

partition_path() {
    local disk="$1" number="$2"
    if [[ "$disk" =~ [0-9]$ ]]; then
        printf '%sp%s\n' "$disk" "$number"
    else
        printf '%s%s\n' "$disk" "$number"
    fi
}

print_partition_plan() {
    local disk="$1"
    cat <<EOF_PLAN
Partition policy: home-server-default-v1

  Target: $disk

  Partition 1   1 MiB       BIOS boot compatibility (EF02)
  Partition 2   512 MiB     EFI System Partition (EF00, FAT32)
  Partition 3   2 GiB       /boot XBOOTLDR (EA00, ext4)
  Partition 4   remainder   root (8304, XFS)
EOF_PLAN
}

partition_and_format() {
    local disk="$1" p2 p3 p4
    [[ "${PARTITION_POLICY:-}" == "home-server-default-v1" ]] || fatal "Unsupported partition policy: ${PARTITION_POLICY:-unset}"

    p2="$(partition_path "$disk" 2)"
    p3="$(partition_path "$disk" 3)"
    p4="$(partition_path "$disk" 4)"

    info "Erasing partition metadata on $disk"
    wipefs -a "$disk"
    sgdisk --zap-all "$disk"

    info "Creating GPT layout"
    sgdisk \
        -n 1:0:+1M   -t 1:EF02 -c 1:"BIOS boot" \
        -n 2:0:+512M -t 2:EF00 -c 2:"EFI System" \
        -n 3:0:+2G   -t 3:EA00 -c 3:"boot" \
        -n 4:0:0     -t 4:8304 -c 4:"root" \
        "$disk"

    udevadm settle

    [[ -b "$p2" && -b "$p3" && -b "$p4" ]] || fatal "Partitions did not appear after partitioning $disk"

    wipefs -a "$p2"
    wipefs -a "$p3"
    wipefs -a "$p4"

    mkfs.vfat -F 32 -n EFI-SYSTEM "$p2"
    mkfs.ext4 -F -L boot "$p3"
    mkfs.xfs -f -L root "$p4"
    udevadm settle
}

mount_target() {
    local disk="$1" p2 p3 p4
    p2="$(partition_path "$disk" 2)"
    p3="$(partition_path "$disk" 3)"
    p4="$(partition_path "$disk" 4)"

    safe_mkdir "$HSI_TARGET_MOUNT"
    mount "$p4" "$HSI_TARGET_MOUNT"

    safe_mkdir "$HSI_TARGET_MOUNT/boot"
    mount "$p3" "$HSI_TARGET_MOUNT/boot"

    # Must be created after /boot is mounted, otherwise the directory is hidden.
    safe_mkdir "$HSI_TARGET_MOUNT/boot/efi"
    mount "$p2" "$HSI_TARGET_MOUNT/boot/efi"
}

capture_target_uuids() {
    local disk="$1" p3 p4 root_uuid boot_uuid
    p3="$(partition_path "$disk" 3)"
    p4="$(partition_path "$disk" 4)"

    root_uuid="$(blkid -s UUID -o value "$p4" || true)"
    boot_uuid="$(blkid -s UUID -o value "$p3" || true)"

    printf 'ROOT_UUID=%s\n' "$root_uuid"
    printf 'BOOT_UUID=%s\n' "$boot_uuid"

    [[ -n "$root_uuid" ]] || fatal "ROOT_UUID is empty; refusing to call bootc"
    [[ -n "$boot_uuid" ]] || fatal "BOOT_UUID is empty; refusing to call bootc"

    HSI_ROOT_UUID="$root_uuid"
    HSI_BOOT_UUID="$boot_uuid"
    export HSI_ROOT_UUID HSI_BOOT_UUID
}

unmount_target() {
    umount "$HSI_TARGET_MOUNT/boot/efi" 2>/dev/null || true
    umount "$HSI_TARGET_MOUNT/boot" 2>/dev/null || true
    umount "$HSI_TARGET_MOUNT" 2>/dev/null || true
}
