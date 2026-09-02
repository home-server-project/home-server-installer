#!/usr/bin/env bash

HSI_SCRATCH_PODMAN=""
HSI_SCRATCH_TMP=""
HSI_LOADED_REF=""

validate_embedded_archive() {
    local archive="$1" expected actual
    expected="${OCI_ARCHIVE_SHA256,,}"
    info "Verifying embedded OCI archive SHA-256"
    actual="$(sha256_file "$archive")"
    [[ "$actual" == "$expected" ]] || fatal "OCI archive checksum mismatch: expected $expected, got $actual"
}

prepare_target_podman_workspace() {
    HSI_SCRATCH_PODMAN="$HSI_TARGET_MOUNT/.installer-podman"
    HSI_SCRATCH_TMP="$HSI_TARGET_MOUNT/.installer-tmp"
    export HSI_SCRATCH_PODMAN HSI_SCRATCH_TMP

    safe_mkdir "$HSI_SCRATCH_PODMAN"
    safe_mkdir "$HSI_SCRATCH_TMP"
    safe_mkdir "$HSI_PODMAN_BIND"

    # bootc's empty-root safety check permits mount points. Self-binding the two
    # temporary directories keeps them explicit mount points while their data
    # remains on the target filesystem.
    mount --bind "$HSI_SCRATCH_PODMAN" "$HSI_SCRATCH_PODMAN"
    mount --bind "$HSI_SCRATCH_TMP" "$HSI_SCRATCH_TMP"
    mount --bind "$HSI_SCRATCH_PODMAN" "$HSI_PODMAN_BIND"

    restorecon -RF "$HSI_PODMAN_BIND" 2>/dev/null || true
}

load_embedded_image() {
    local archive="$1" output detected actual_digest

    info "Loading embedded OCI archive into target-backed Podman storage"
    output="$(env TMPDIR="$HSI_SCRATCH_TMP" podman load --input "$archive" 2>&1)" || {
        printf '%s\n' "$output" >&2
        fatal "podman load failed"
    }
    printf '%s\n' "$output"

    if [[ -n "${OCI_LOCAL_REF:-}" ]]; then
        detected="$OCI_LOCAL_REF"
    else
        detected="$(printf '%s\n' "$output" | sed -n 's/^Loaded image: //p' | tail -n1)"
    fi

    [[ -n "$detected" ]] || fatal "Could not determine the loaded image reference. Builder should set OCI_LOCAL_REF."
    podman image exists "$detected" || fatal "Loaded image reference not found in Podman storage: $detected"

    actual_digest="$(skopeo inspect --format '{{.Digest}}' "containers-storage:$detected" 2>/dev/null || true)"
    [[ -n "$actual_digest" ]] || fatal "Could not inspect digest for loaded image: $detected"
    [[ "$actual_digest" == "$IMAGE_DIGEST" ]] || fatal "Loaded image digest mismatch: expected $IMAGE_DIGEST, got $actual_digest"

    HSI_LOADED_REF="$detected"
    export HSI_LOADED_REF

    info "Loaded image digest verified: $actual_digest"
}

cleanup_target_podman_workspace() {
    if mountpoint -q "$HSI_PODMAN_BIND" 2>/dev/null; then
        umount "$HSI_PODMAN_BIND"
    fi
    if [[ -n "${HSI_SCRATCH_TMP:-}" ]] && mountpoint -q "$HSI_SCRATCH_TMP" 2>/dev/null; then
        umount "$HSI_SCRATCH_TMP"
    fi
    if [[ -n "${HSI_SCRATCH_PODMAN:-}" ]] && mountpoint -q "$HSI_SCRATCH_PODMAN" 2>/dev/null; then
        umount "$HSI_SCRATCH_PODMAN"
    fi

    [[ -z "${HSI_SCRATCH_PODMAN:-}" ]] || rm -rf --one-file-system "$HSI_SCRATCH_PODMAN"
    [[ -z "${HSI_SCRATCH_TMP:-}" ]] || rm -rf --one-file-system "$HSI_SCRATCH_TMP"
}
