#!/usr/bin/env bash

HSI_ALLOWED_METADATA_KEYS=(
    INSTALLER_FORMAT_VERSION
    INSTALLER_VERSION
    IMAGE_VARIANT
    IMAGE_REF
    IMAGE_DIGEST
    IMAGE_PINNED_REF
    OCI_ARCHIVE
    OCI_ARCHIVE_SHA256
    OCI_LOCAL_REF
    SSH_PUBLIC_KEY
    PARTITION_POLICY
    UPDATE_POLICY
)

metadata_key_allowed() {
    local key="$1" allowed
    for allowed in "${HSI_ALLOWED_METADATA_KEYS[@]}"; do
        [[ "$key" == "$allowed" ]] && return 0
    done
    return 1
}

metadata_set() {
    local key="$1" value="$2"
    printf -v "$key" '%s' "$value"
    export "${key?}"
}

load_metadata_file() {
    local file="$1"
    local line key value lineno=0

    [[ -r "$file" ]] || fatal "Installer metadata not readable: $file"

    while IFS= read -r line || [[ -n "$line" ]]; do
        lineno=$((lineno + 1))
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" == *=* ]] || fatal "Invalid metadata line $lineno in $file"

        key="${line%%=*}"
        value="${line#*=}"

        [[ "$key" =~ ^[A-Z0-9_]+$ ]] || fatal "Invalid metadata key on line $lineno: $key"
        metadata_key_allowed "$key" || fatal "Unsupported metadata key on line $lineno: $key"

        # Metadata is intentionally data, not shell. Quotes, command substitutions,
        # escapes, and variable expansion are never evaluated here.
        metadata_set "$key" "$value"
    done < "$file"
}

require_metadata() {
    local key value
    for key in "$@"; do
        value="${!key:-}"
        [[ -n "$value" ]] || fatal "Required installer metadata is missing: $key"
    done
}

validate_metadata() {
    require_metadata \
        INSTALLER_FORMAT_VERSION \
        IMAGE_REF \
        IMAGE_DIGEST \
        OCI_ARCHIVE \
        OCI_ARCHIVE_SHA256 \
        SSH_PUBLIC_KEY

    [[ "$INSTALLER_FORMAT_VERSION" == "1" ]] || fatal "Unsupported installer format: $INSTALLER_FORMAT_VERSION"
    is_sha256_digest "$IMAGE_DIGEST" || fatal "IMAGE_DIGEST is not a sha256 digest"
    is_sha256_hex "$OCI_ARCHIVE_SHA256" || fatal "OCI_ARCHIVE_SHA256 is not a SHA-256 checksum"

    [[ "$IMAGE_REF" =~ ^[A-Za-z0-9._:/@+-]+$ ]] || fatal "IMAGE_REF contains unsupported characters"
    if [[ -n "${IMAGE_PINNED_REF:-}" ]]; then
        [[ "$IMAGE_PINNED_REF" =~ ^[A-Za-z0-9._:/@+-]+$ ]] || fatal "IMAGE_PINNED_REF contains unsupported characters"
    fi
    if [[ -n "${OCI_LOCAL_REF:-}" ]]; then
        [[ "$OCI_LOCAL_REF" =~ ^[A-Za-z0-9._:/@+-]+$ ]] || fatal "OCI_LOCAL_REF contains unsupported characters"
    fi

    [[ "$OCI_ARCHIVE" == /home-server/* ]] || fatal "OCI_ARCHIVE must be an absolute /home-server/... path"
    [[ "$SSH_PUBLIC_KEY" == /home-server/* ]] || fatal "SSH_PUBLIC_KEY must be an absolute /home-server/... path"

    PARTITION_POLICY="${PARTITION_POLICY:-home-server-default-v1}"
    UPDATE_POLICY="${UPDATE_POLICY:-home-server-stage-v1}"
    export PARTITION_POLICY UPDATE_POLICY
}

locate_iso_root() {
    local candidate

    if [[ -n "${HSI_ISO_ROOT:-}" ]]; then
        [[ -d "$HSI_ISO_ROOT/home-server" ]] || fatal "HSI_ISO_ROOT does not contain /home-server: $HSI_ISO_ROOT"
        printf '%s\n' "$HSI_ISO_ROOT"
        return 0
    fi

    for candidate in /run/media/iso /run/media/* /mnt/iso /media/*; do
        [[ -d "$candidate/home-server" ]] || continue
        [[ -r "$candidate/home-server/config/install.env" ]] || continue
        printf '%s\n' "$candidate"
        return 0
    done

    fatal "Could not locate installation media. Set HSI_ISO_ROOT if needed."
}

resolve_iso_payload_path() {
    local iso_root="$1" payload_path="$2" resolved
    resolved="$(readlink -m "$iso_root$payload_path")"
    path_is_under "$resolved" "$iso_root/home-server" || fatal "Payload path escapes /home-server: $payload_path"
    printf '%s\n' "$resolved"
}
