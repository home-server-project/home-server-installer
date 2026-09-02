#!/usr/bin/env bash

run_bootc_install() {
    local pubkey="$1"
    local source_imgref

    [[ -n "${HSI_ROOT_UUID:-}" ]] || fatal "ROOT_UUID is empty before bootc"
    [[ -n "${HSI_BOOT_UUID:-}" ]] || fatal "BOOT_UUID is empty before bootc"
    [[ -n "${HSI_LOADED_REF:-}" ]] || fatal "No loaded OCI image reference"
    [[ -r "$pubkey" ]] || fatal "SSH public key disappeared before bootc: $pubkey"

    source_imgref="containers-storage:$HSI_LOADED_REF"

    info "Starting bootc installation"
    info "Root UUID: $HSI_ROOT_UUID"
    info "Boot UUID: $HSI_BOOT_UUID"

    # shellcheck disable=SC2016 -- variables intentionally expand inside the target container.
    env TMPDIR="$HSI_SCRATCH_TMP" \
        podman run --rm \
        --pull=never \
        --privileged \
        --pid=host \
        --ipc=host \
        -v "$HSI_PODMAN_BIND:$HSI_PODMAN_BIND" \
        -v /dev:/dev \
        -v "$HSI_TARGET_MOUNT:/target:rslave" \
        -v "$pubkey:/run/home-server-installer.pub:ro" \
        -e HSI_ROOT_UUID="$HSI_ROOT_UUID" \
        -e HSI_BOOT_UUID="$HSI_BOOT_UUID" \
        -e HSI_SOURCE_IMGREF="$source_imgref" \
        -e HSI_TARGET_IMGREF="$IMAGE_REF" \
        --security-opt label=type:unconfined_t \
        "$HSI_LOADED_REF" \
        sh -eu -c '
            test -n "$HSI_ROOT_UUID"
            test -n "$HSI_BOOT_UUID"
            test -n "$HSI_SOURCE_IMGREF"
            test -n "$HSI_TARGET_IMGREF"

            bootc install to-filesystem \
                --source-imgref "$HSI_SOURCE_IMGREF" \
                --target-imgref "$HSI_TARGET_IMGREF" \
                --root-mount-spec "UUID=$HSI_ROOT_UUID" \
                --boot-mount-spec "UUID=$HSI_BOOT_UUID" \
                --root-ssh-authorized-keys /run/home-server-installer.pub \
                --karg=console=tty0 \
                --karg=console=ttyS0,115200n8 \
                --generic-image \
                --skip-fetch-check \
                --skip-finalize \
                /target
        '
}

find_deployment_root() {
    local -a deployments=()
    mapfile -t deployments < <(find "$HSI_TARGET_MOUNT/ostree/deploy" -type d -path '*/deploy/*.0' -print 2>/dev/null)

    ((${#deployments[@]} == 1)) || fatal "Expected exactly one OSTree deployment root; found ${#deployments[@]}"
    printf '%s\n' "${deployments[0]}"
}

finalize_bootc_install() {
    info "Finalizing bootc installation"
    bootc install finalize "$HSI_TARGET_MOUNT"
}
