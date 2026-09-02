#!/usr/bin/env bash

apply_update_policy() {
    local deploy="$1"
    local preset_src="$2"
    local policy="${UPDATE_POLICY:-home-server-stage-v1}"

    case "$policy" in
        preserve)
            info "Update policy is preserve; leaving image policy unchanged"
            return 0
            ;;
        home-server-stage-v1)
            ;;
        *)
            fatal "Unsupported UPDATE_POLICY: $policy"
            ;;
    esac

    [[ -r "$preset_src" ]] || fatal "Home Server preset missing: $preset_src"
    [[ -f "$deploy/etc/rpm-ostreed.conf" ]] || fatal "rpm-ostreed.conf missing from deployment; home-server-stage-v1 is not supported by this image"

    if ! grep -qE '^[[:space:]]*AutomaticUpdatePolicy=stage[[:space:]]*$' "$deploy/etc/rpm-ostreed.conf"; then
        fatal "home-server-stage-v1 requires AutomaticUpdatePolicy=stage in the selected image; installer will not rewrite an unverified rpm-ostree policy"
    fi

    info "Installing persistent Home Server systemd preset"
    install -D -m 0644 "$preset_src" "$deploy/etc/systemd/system-preset/00-home-server.preset"

    # The explicit mask is stronger than preset policy and protects against the
    # Fedora CoreOS vendor preset which enables zincati.service on first boot.
    systemctl --root="$deploy" disable zincati.service >/dev/null 2>&1 || true
    systemctl --root="$deploy" mask zincati.service
    systemctl --root="$deploy" enable rpm-ostreed-automatic.timer
}

verify_update_policy_offline() {
    local deploy="$1" policy="${UPDATE_POLICY:-home-server-stage-v1}"
    [[ "$policy" == "preserve" ]] && return 0

    [[ "$(systemctl --root="$deploy" is-enabled zincati.service 2>&1 || true)" == "masked" ]] || fatal "Zincati is not masked in the deployment"
    [[ "$(systemctl --root="$deploy" is-enabled rpm-ostreed-automatic.timer 2>&1 || true)" == "enabled" ]] || fatal "rpm-ostreed automatic timer is not enabled in the deployment"
    grep -qE '^[[:space:]]*AutomaticUpdatePolicy=stage[[:space:]]*$' "$deploy/etc/rpm-ostreed.conf" || fatal "AutomaticUpdatePolicy is not stage"
    [[ -L "$deploy/etc/systemd/system/zincati.service" ]] || fatal "Zincati mask symlink is missing"
    [[ "$(readlink "$deploy/etc/systemd/system/zincati.service")" == "/dev/null" ]] || fatal "Zincati mask does not point to /dev/null"
    [[ -f "$deploy/etc/systemd/system-preset/00-home-server.preset" ]] || fatal "Home Server preset is missing"
}
