#!/usr/bin/env bash

HSI_DISK_NAMES=()
HSI_DISK_SIZES=()
HSI_DISK_MODELS=()
HSI_DISK_TRANS=()
HSI_DISK_RM=()
HSI_DISK_ROTA=()
HSI_DISK_SERIALS=()
HSI_LIVE_DISK=""

pair_field() {
    local line="$1" key="$2"
    local regex="${key}=\"([^\"]*)\""
    if [[ "$line" =~ $regex ]]; then
        printf '%s\n' "${BASH_REMATCH[1]}"
    fi
}

parent_disk_for_source() {
    local source="$1" parent type
    [[ "$source" == /dev/* ]] || return 0
    type="$(lsblk -dn -o TYPE "$source" 2>/dev/null || true)"
    if [[ "$type" == "disk" ]]; then
        printf '%s\n' "$source"
        return 0
    fi
    parent="$(lsblk -dn -o PKNAME "$source" 2>/dev/null | head -n1)"
    [[ -n "$parent" ]] && printf '/dev/%s\n' "$parent"
}

find_live_media_disk() {
    local iso_root="$1" source
    source="$(findmnt -n -o SOURCE --target "$iso_root" 2>/dev/null || true)"
    parent_disk_for_source "$source"
}

is_hardware_disk() {
    local dev="$1" base
    base="${dev##*/}"

    # Alpha intentionally uses a conservative whole-disk allowlist. Exotic
    # persistent-memory/network/virtual block devices can be added later only
    # after their erase semantics are tested.
    case "$base" in
        sd[a-z]*|hd[a-z]*|vd[a-z]*|xvd[a-z]*|nvme[0-9]*n[0-9]*|mmcblk[0-9]*) ;;
        *) return 1 ;;
    esac

    [[ -e "/sys/class/block/$base/device" ]]
}

discover_disks() {
    local iso_root="$1"
    local line name size model tran rm rota type serial

    HSI_DISK_NAMES=()
    HSI_DISK_SIZES=()
    HSI_DISK_MODELS=()
    HSI_DISK_TRANS=()
    HSI_DISK_RM=()
    HSI_DISK_ROTA=()
    HSI_DISK_SERIALS=()
    HSI_LIVE_DISK="$(find_live_media_disk "$iso_root")"

    while IFS= read -r line; do
        name="$(pair_field "$line" NAME)"
        size="$(pair_field "$line" SIZE)"
        model="$(pair_field "$line" MODEL)"
        tran="$(pair_field "$line" TRAN)"
        rm="$(pair_field "$line" RM)"
        rota="$(pair_field "$line" ROTA)"
        type="$(pair_field "$line" TYPE)"
        serial="$(pair_field "$line" SERIAL)"

        [[ "$type" == "disk" ]] || continue
        is_hardware_disk "$name" || continue
        [[ -n "$HSI_LIVE_DISK" && "$name" == "$HSI_LIVE_DISK" ]] && continue

        HSI_DISK_NAMES+=("$name")
        HSI_DISK_SIZES+=("$size")
        HSI_DISK_MODELS+=("${model:-Unknown model}")
        HSI_DISK_TRANS+=("${tran:-unknown}")
        HSI_DISK_RM+=("${rm:-0}")
        HSI_DISK_ROTA+=("${rota:-0}")
        HSI_DISK_SERIALS+=("${serial:-unknown}")
    done < <(lsblk -d -n -P -p -b -o NAME,SIZE,MODEL,TRAN,RM,ROTA,TYPE,SERIAL 2>/dev/null)
}

print_disks() {
    local i class warning
    if ((${#HSI_DISK_NAMES[@]} == 0)); then
        printf 'No eligible physical installation disks detected.\n'
        return 0
    fi

    printf 'Detected installation disks:\n\n'
    for i in "${!HSI_DISK_NAMES[@]}"; do
        if [[ "${HSI_DISK_ROTA[$i]}" == "1" ]]; then
            class="HDD"
        else
            class="SSD / solid-state"
        fi
        warning=""
        [[ "${HSI_DISK_RM[$i]}" == "1" ]] && warning="  [removable/external]"
        printf '  %d. %s\n' "$((i + 1))" "${HSI_DISK_NAMES[$i]}"
        printf '     Model:     %s\n' "${HSI_DISK_MODELS[$i]}"
        printf '     Size:      %s\n' "$(human_bytes "${HSI_DISK_SIZES[$i]}")"
        printf '     Transport: %s\n' "${HSI_DISK_TRANS[$i]}"
        printf '     Class:     %s%s\n' "$class" "$warning"
        printf '     Serial:    %s\n\n' "${HSI_DISK_SERIALS[$i]}"
    done

    if [[ -n "$HSI_LIVE_DISK" ]]; then
        printf 'Installation media disk excluded: %s\n\n' "$HSI_LIVE_DISK"
    fi
}

select_disk_interactive() {
    local choice
    ((${#HSI_DISK_NAMES[@]} > 0)) || fatal "No eligible installation disks detected"

    while true; do
        read -r -p "Select installation disk [1-${#HSI_DISK_NAMES[@]}]: " choice
        [[ "$choice" =~ ^[0-9]+$ ]] || { warn "Enter a number."; continue; }
        ((choice >= 1 && choice <= ${#HSI_DISK_NAMES[@]})) || { warn "Selection out of range."; continue; }
        printf '%s\n' "${HSI_DISK_NAMES[$((choice - 1))]}"
        return 0
    done
}

validate_target_disk() {
    local disk="$1" type mounted
    [[ -b "$disk" ]] || fatal "Target is not a block device: $disk"
    type="$(lsblk -dn -o TYPE "$disk" 2>/dev/null || true)"
    [[ "$type" == "disk" ]] || fatal "Target is not a whole disk: $disk"
    is_hardware_disk "$disk" || fatal "Target is not a supported hardware-backed disk: $disk"
    [[ -z "$HSI_LIVE_DISK" || "$disk" != "$HSI_LIVE_DISK" ]] || fatal "Refusing to erase the installation media: $disk"

    mounted="$(lsblk -nr -o MOUNTPOINTS "$disk" | sed '/^[[:space:]]*$/d' || true)"
    [[ -z "$mounted" ]] || fatal "Target disk or one of its partitions is mounted:\n$mounted"

}
