#!/usr/bin/env bash
set -euo pipefail

declare -A targets=(
  [model]=100 [validate]=99 [ignition]=100 [github]=96
  [bakery]=100 [probe]=100 [runner]=100 [install]=98
  [headless]=99 [wizard]=99 [iso]=100 [tui]=98
  [demo]=100 [fcos]=100
)

fail=0
for pkg in "${!targets[@]}"; do
  pct=$(go test -count=1 -cover ./internal/${pkg}/... 2>/dev/null \
    | awk '/coverage:/ {gsub("%",""); print $(NF-2); exit}')
  pct=${pct%.*}
  if [[ -z "$pct" ]]; then
    echo "FAIL  internal/${pkg}   no coverage reported"
    fail=1
    continue
  fi
  if (( pct < targets[$pkg] )); then
    echo "FAIL  internal/${pkg}  ${pct}%  (target ${targets[$pkg]}%)"
    fail=1
  else
    echo "ok    internal/${pkg}  ${pct}%  (target ${targets[$pkg]}%)"
  fi
done

check_extra_package() {
  local pkg=$1
  local target=$2
  local pct

  pct=$(go test -count=1 -cover ./${pkg}/... 2>/dev/null \
    | awk '/coverage:/ {gsub("%",""); print $(NF-2); exit}')
  pct=${pct%.*}

  if [[ -z "$pct" ]]; then
    echo "FAIL  ${pkg}   no coverage reported"
    fail=1
  elif (( pct < target )); then
    echo "FAIL  ${pkg}  ${pct}%  (target ${target}%)"
    fail=1
  else
    echo "ok    ${pkg}  ${pct}%  (target ${target}%)"
  fi
}

check_extra_package scripts/catalog_check 100
check_extra_package cmd/knuckle 85
check_extra_package cmd/compile-butane-fresh 100
check_extra_package cmd/nvidia-check 95

exit "$fail"
