#!/usr/bin/env bash

validate_ssh_public_key() {
    local file="$1"
    local first_line type payload extra

    [[ -r "$file" ]] || fatal "SSH public key not readable: $file"

    if grep -qE 'BEGIN (OPENSSH|RSA|EC|DSA|PRIVATE) PRIVATE KEY|BEGIN PRIVATE KEY' "$file"; then
        fatal "Private-key material detected. Only a .pub public key may be embedded."
    fi

    first_line="$(grep -v '^[[:space:]]*$' "$file" | head -n 1)"
    [[ -n "$first_line" ]] || fatal "SSH public key file is empty"

    read -r type payload extra <<< "$first_line"
    case "$type" in
        ssh-ed25519|ssh-rsa|ecdsa-sha2-nistp256|ecdsa-sha2-nistp384|ecdsa-sha2-nistp521|sk-ssh-ed25519@openssh.com|sk-ecdsa-sha2-nistp256@openssh.com)
            ;;
        *) fatal "Unsupported or invalid SSH public key type: $type" ;;
    esac

    [[ "$payload" =~ ^[A-Za-z0-9+/=]+$ ]] || fatal "SSH public key payload is not valid base64 text"

    if command -v ssh-keygen >/dev/null 2>&1; then
        ssh-keygen -lf "$file" >/dev/null 2>&1 || fatal "ssh-keygen rejected the public key"
    fi
}

ssh_key_fingerprint() {
    local file="$1"
    if command -v ssh-keygen >/dev/null 2>&1; then
        ssh-keygen -lf "$file" -E sha256 | awk '{print $2}'
    else
        printf 'unavailable\n'
    fi
}
