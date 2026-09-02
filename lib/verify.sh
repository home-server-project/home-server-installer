#!/usr/bin/env bash

verify_bls_uuids() {
    local root_uuid="$1" boot_uuid="$2" entries
    entries="$HSI_TARGET_MOUNT/boot/loader/entries"
    [[ -d "$entries" ]] || fatal "BLS entries directory missing: $entries"

    grep -R -q -- "root=UUID=$root_uuid" "$entries" || fatal "BLS does not contain expected root UUID"
    grep -R -q -- "boot=UUID=$boot_uuid" "$entries" || fatal "BLS does not contain expected boot UUID"
}

verify_origin() {
    local origin
    origin="$(cat "$HSI_TARGET_MOUNT"/ostree/deploy/*/deploy/*.origin 2>/dev/null || true)"
    [[ -n "$origin" ]] || fatal "OSTree origin file missing"
    [[ "$origin" == *"$IMAGE_REF"* ]] || fatal "OSTree origin does not contain target image reference $IMAGE_REF"
}

verify_ssh_injection() {
    local deploy="$1"
    local rule="$deploy/etc/tmpfiles.d/bootc-root-ssh.conf"
    [[ -s "$rule" ]] || fatal "bootc SSH tmpfiles rule missing"
    grep -q '/var/roothome/.ssh/authorized_keys' "$rule" || fatal "SSH tmpfiles rule does not target root authorized_keys"
}

verify_boot_artifacts() {
    local disk="$1" p2 p3
    p2="$(partition_path "$disk" 2)"
    p3="$(partition_path "$disk" 3)"
    blkid "$p2" | grep -q 'TYPE="vfat"' || fatal "EFI partition is not FAT"
    blkid "$p3" | grep -q 'TYPE="ext4"' || fatal "/boot partition is not ext4"
    [[ -d "$HSI_TARGET_MOUNT/boot/loader/entries" ]] || fatal "Boot loader entries missing"
}

verify_installation() {
    local disk="$1" deploy="$2"
    info "Verifying installed target"
    verify_bls_uuids "$HSI_ROOT_UUID" "$HSI_BOOT_UUID"
    verify_origin
    verify_ssh_injection "$deploy"
    verify_update_policy_offline "$deploy"
    verify_boot_artifacts "$disk"
}
