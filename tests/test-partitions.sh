#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../lib/common.sh
source "$ROOT/lib/common.sh"
# shellcheck source=../lib/partitions.sh
source "$ROOT/lib/partitions.sh"

[[ "$(partition_path /dev/vda 3)" == "/dev/vda3" ]]
[[ "$(partition_path /dev/sda 4)" == "/dev/sda4" ]]
[[ "$(partition_path /dev/nvme0n1 2)" == "/dev/nvme0n1p2" ]]
[[ "$(partition_path /dev/mmcblk0 1)" == "/dev/mmcblk0p1" ]]

plan="$(print_partition_plan /dev/nvme0n1)"
[[ "$plan" == *"512 MiB"* ]]
[[ "$plan" == *"2 GiB"* ]]
[[ "$plan" == *"remainder"* ]]

printf 'PASS partition path and plan tests\n'
