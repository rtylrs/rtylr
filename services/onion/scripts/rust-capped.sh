#!/usr/bin/env bash
# scripts/rust-capped.sh: run a command under a hard CPU and memory cap.
#
#   scripts/rust-capped.sh cargo check -p core
#
# Memory and CPU are enforced by a systemd cgroup, so the kernel kills the
# command rather than the machine. If the cgroup cannot be applied the script
# fails closed. Tunables (environment):
#   RM_MEM_MAX         memory cap, systemd syntax (default: 60% of RAM)
#   RM_CPUS            CPUs to use (default: half the cores, at least 1)
#   RM_MIN_DISK_GB     refuse to start below this much free disk (default: 5)
#   RM_TIMEOUT         seconds before the command is killed (default: 3600)
#   RM_ALLOW_UNCAPPED  1 = fall back to nice, ionice and job limits only
#   RM_DRY_RUN         1 = print the final command and exit

set -euo pipefail

if [ "$#" -eq 0 ]; then
  echo "usage: $0 <command> [args...]" >&2
  exit 64
fi

ncpu="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 1)"
cpus="${RM_CPUS:-$(( (ncpu + 1) / 2 ))}"
[ "$cpus" -ge 1 ] || cpus=1

mem="${RM_MEM_MAX:-}"
if [ -z "$mem" ]; then
  if [ -r /proc/meminfo ]; then
    total_kb="$(awk '/^MemTotal:/ {print $2}' /proc/meminfo)"
    mem="$(( total_kb * 60 / 100 ))K"
  else
    echo "rust-capped: cannot read /proc/meminfo; set RM_MEM_MAX (e.g. 2G)" >&2
    exit 78
  fi
fi

min_disk_gb="${RM_MIN_DISK_GB:-5}"
free_kb="$(df -Pk . | awk 'NR==2 {print $4}')"
if [ "$(( free_kb / 1024 / 1024 ))" -lt "$min_disk_gb" ]; then
  echo "rust-capped: less than ${min_disk_gb} GB free disk; clean target/ or raise nothing, ask the human" >&2
  exit 75
fi

export CARGO_BUILD_JOBS="$cpus"
export RUST_TEST_THREADS="$cpus"
export CARGO_INCREMENTAL=0
export CARGO_PROFILE_DEV_DEBUG=0
export CARGO_PROFILE_TEST_DEBUG=0
export GOMAXPROCS="$cpus"
export GOFLAGS="${GOFLAGS:-} -p=$cpus"
export MAKEFLAGS="-j$cpus"

timeout_s="${RM_TIMEOUT:-3600}"
low_priority=(nice -n 10)
command -v ionice >/dev/null 2>&1 && low_priority+=(ionice -c2 -n7)

cmd=("$@")
if command -v systemd-run >/dev/null 2>&1 && systemctl --user show-environment >/dev/null 2>&1; then
  final=(systemd-run --user --scope --quiet --collect
    -p "MemoryMax=$mem" -p "MemorySwapMax=0"
    -p "CPUQuota=$(( cpus * 100 ))%" -p "TasksMax=4096"
    -- timeout "$timeout_s" "${low_priority[@]}" "${cmd[@]}")
elif [ "${RM_ALLOW_UNCAPPED:-0}" = "1" ]; then
  echo "rust-capped: WARNING no cgroup available; only priority and job limits apply" >&2
  final=(timeout "$timeout_s" "${low_priority[@]}" "${cmd[@]}")
else
  cat >&2 <<'EOF'
rust-capped: cannot apply a hard memory cap (systemd user session not available).
Fix on the server, as the user running the agent:
  loginctl enable-linger "$USER"   # as root, once
  systemctl --user show-environment  # must succeed
Or accept a soft cap by exporting RM_ALLOW_UNCAPPED=1 (not recommended).
EOF
  exit 78
fi

if [ "${RM_DRY_RUN:-0}" = "1" ]; then
  printf '%q ' "${final[@]}"; echo
  exit 0
fi

exec "${final[@]}"
