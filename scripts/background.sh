#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
state_dir="${NSK_RUNTIME_DIR:-$root_dir/runtime}/background"
log_dir="${NSK_LOG_DIR:-$root_dir/runtime/logs}"

usage() {
  cat <<'EOF'
Usage:
  background.sh start NAME COMMAND
  background.sh stop NAME [TIMEOUT_SECONDS]
  background.sh status NAME
  background.sh list
  background.sh logs NAME [LINES]
EOF
}

validate_name() {
  [[ "${1:-}" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]] || {
    printf 'invalid task name: %s\n' "${1:-<missing>}" >&2
    exit 2
  }
}

validate_positive_integer() {
  [[ "${1:-}" =~ ^[0-9]+$ ]] || {
    printf 'expected a non-negative integer: %s\n' "${1:-<missing>}" >&2
    exit 2
  }
}

pid_file() { printf '%s/%s.pid' "$state_dir" "$1"; }
command_file() { printf '%s/%s.command' "$state_dir" "$1"; }
mode_file() { printf '%s/%s.mode' "$state_dir" "$1"; }
log_file() { printf '%s/%s.log' "$log_dir" "$1"; }

read_pid() {
  local name="$1" pid_path
  pid_path="$(pid_file "$name")"
  [[ -s "$pid_path" ]] || return 1
  read -r pid <"$pid_path"
  [[ "$pid" =~ ^[1-9][0-9]*$ ]] || return 1
  printf '%s' "$pid"
}

read_mode() {
  local name="$1" mode_path
  mode_path="$(mode_file "$name")"
  if [[ -s "$mode_path" ]]; then
    read -r mode <"$mode_path"
    printf '%s' "$mode"
  else
    printf 'pid'
  fi
}

is_running() {
  local pid="$1" mode="${2:-pid}"
  if [[ "$mode" == group ]]; then
    kill -0 -- "-$pid" 2>/dev/null || kill -0 "$pid" 2>/dev/null
  else
    kill -0 "$pid" 2>/dev/null
  fi
}

start_task() {
  local name="$1" command="$2" pid_path command_path mode_path output_path pid mode
  validate_name "$name"
  [[ -n "$command" ]] || {
    usage >&2
    exit 2
  }
  mkdir -p "$state_dir" "$log_dir"
  pid_path="$(pid_file "$name")"
  command_path="$(command_file "$name")"
  mode_path="$(mode_file "$name")"
  output_path="$(log_file "$name")"

  if pid="$(read_pid "$name")" && is_running "$pid" "$(read_mode "$name")"; then
    printf '%s is already running (pid %s)\n' "$name" "$pid" >&2
    exit 1
  fi

  rm -f "$pid_path"
  printf '%s\n' "$command" >"${command_path}.tmp"
  mv "${command_path}.tmp" "$command_path"
  if command -v setsid >/dev/null 2>&1; then
    nohup setsid bash -lc "$command" >"$output_path" 2>&1 </dev/null &
    mode=group
  else
    nohup bash -lc "$command" >"$output_path" 2>&1 </dev/null &
    mode=pid
  fi
  pid=$!
  printf '%s\n' "$pid" >"${pid_path}.tmp"
  mv "${pid_path}.tmp" "$pid_path"
  printf '%s\n' "$mode" >"${mode_path}.tmp"
  mv "${mode_path}.tmp" "$mode_path"
  printf '%s started (pid %s)\n' "$name" "$pid"
}

stop_task() {
  local name="$1" timeout="$2" pid mode elapsed=0
  validate_name "$name"
  validate_positive_integer "$timeout"
  pid="$(read_pid "$name")" || {
    printf '%s is not running\n' "$name"
    return 0
  }
  mode="$(read_mode "$name")"
  if ! is_running "$pid" "$mode"; then
    rm -f "$(pid_file "$name")" "$(command_file "$name")" "$(mode_file "$name")"
    printf '%s is stale (pid %s)\n' "$name" "$pid"
    return 0
  fi

  if [[ "$mode" == group ]]; then
    kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid"
  else
    kill -TERM "$pid"
  fi
  while is_running "$pid" "$mode" && ((elapsed < timeout * 10)); do
    sleep 0.1
    ((elapsed += 1))
  done
  if is_running "$pid" "$mode"; then
    if [[ "$mode" == group ]]; then
      kill -KILL -- "-$pid" 2>/dev/null || kill -KILL "$pid"
    else
      kill -KILL "$pid"
    fi
    printf '%s killed after %ss (pid %s)\n' "$name" "$timeout" "$pid"
  else
    printf '%s stopped (pid %s)\n' "$name" "$pid"
  fi
  rm -f "$(pid_file "$name")" "$(command_file "$name")" "$(mode_file "$name")"
}

status_task() {
  local name="$1" pid mode command_path
  validate_name "$name"
  pid="$(read_pid "$name")" || {
    printf '%s is not running\n' "$name"
    return 1
  }
  mode="$(read_mode "$name")"
  command_path="$(command_file "$name")"
  if is_running "$pid" "$mode"; then
    printf '%s running (pid %s, %s): %s\n' "$name" "$pid" "$mode" "$(<"$command_path")"
    return 0
  fi
  printf '%s is stale (pid %s)\n' "$name" "$pid"
  return 1
}

list_tasks() {
  local path name
  shopt -s nullglob
  for path in "$state_dir"/*.pid; do
    name="$(basename "$path" .pid)"
    status_task "$name" || true
  done
}

logs_task() {
  local name="$1" lines="$2" output_path
  validate_name "$name"
  validate_positive_integer "$lines"
  output_path="$(log_file "$name")"
  [[ -f "$output_path" ]] || {
    printf 'no log for %s\n' "$name" >&2
    return 1
  }
  tail -n "$lines" "$output_path"
}

main() {
  local action="${1:-}"
  case "$action" in
    start)
      [[ $# -eq 3 ]] || {
        usage >&2
        exit 2
      }
      start_task "$2" "$3"
      ;;
    stop)
      [[ $# -ge 2 && $# -le 3 ]] || {
        usage >&2
        exit 2
      }
      stop_task "$2" "${3:-10}"
      ;;
    status)
      [[ $# -eq 2 ]] || {
        usage >&2
        exit 2
      }
      status_task "$2"
      ;;
    list)
      [[ $# -eq 1 ]] || {
        usage >&2
        exit 2
      }
      list_tasks
      ;;
    logs)
      [[ $# -ge 2 && $# -le 3 ]] || {
        usage >&2
        exit 2
      }
      logs_task "$2" "${3:-80}"
      ;;
    help | --help | -h)
      usage
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
}

main "$@"
