#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin_dir="${NSK_TOOL_BIN_DIR:-$root_dir/runtime/tools/bin}"
golangci_version="2.13.1"
goreleaser_version="2.18.0"
shellcheck_version="0.11.0"
tombi_version="1.4.1"
temp_dir=""

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'required command not found: %s\n' "$1" >&2
    exit 1
  }
}

platform_name() {
  case "$(uname -s)" in
    Linux) printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *)
      printf 'unsupported operating system: %s\n' "$(uname -s)" >&2
      exit 1
      ;;
  esac
}

architecture_name() {
  case "$(uname -m)" in
    x86_64 | amd64) printf 'amd64' ;;
    arm64 | aarch64) printf 'arm64' ;;
    *)
      printf 'unsupported architecture: %s\n' "$(uname -m)" >&2
      exit 1
      ;;
  esac
}

golangci_checksum() {
  case "$1-$2" in
    linux-amd64) printf 'b17bfbc9d4aaa48be7f4f1ce3240bc3d8200c870c072bacf15c26219e2cfb9cc' ;;
    linux-arm64) printf '908317c23db18448f924e853b3d8a659fd919614cd438f224810a4053daa2607' ;;
    darwin-amd64) printf '2c373363953e4e0bee2a03b7fe864a5eb6a3822927cb077d9ca33f2ae3cb2da2' ;;
    darwin-arm64) printf '0c9818baf6fb8ad26c6d2ef51b68d5a1e260ef07727036b1431647cc44637c7c' ;;
  esac
}

shellcheck_checksum() {
  case "$1-$2" in
    linux-amd64) printf 'b7af85e41cc99489dcc21d66c6d5f3685138f06d34651e6d34b42ec6d54fe6f6' ;;
    linux-arm64) printf '68a8133197a50beb8803f8d42f9908d1af1c5540d4bb05fdfca8c1fa47decefc' ;;
    darwin-amd64) printf 'c2c15e08df0e8fbc374c335b230a7ee958c313fa5714817a59aa59f1aa594f51' ;;
    darwin-arm64) printf '339b930feb1ea764467013cc1f72d09cd6b869ebf1013296ba9055ab2ffbd26f' ;;
  esac
}

goreleaser_checksum() {
  case "$1-$2" in
    linux-amd64) printf '41cdf49b653784b03a08013dd99e382cd5d463049e915c2d818eaed182ae6197' ;;
    linux-arm64) printf '1975566c9668e6f4247e6bb57656f21da13635c24d948ef47b1232e5c864a35b' ;;
    darwin-amd64) printf 'c115f9ca07163d55885ba2276c5c2efebc95d60f7f7f69fe2dd6a54e97ac6db4' ;;
    darwin-arm64) printf '1c42b87cbce094a60f1a94dab0c71f640dbe4396fa5dc632b5c25bf14b1e88fc' ;;
  esac
}

tombi_checksum() {
  case "$1-$2" in
    linux-amd64) printf '9aa69eb3e75a4a22a961b8a1c8cc44e4f81328ce25ad5b10d151be1a09faa88d' ;;
    linux-arm64) printf '21f51d092597053266e0ed051082743b5956b6de2f0db1cecce78e0eb29165e5' ;;
    darwin-amd64) printf '4a14a0bf18ec0bbbeab3003f6c0dea3ffabb2cb38649ebfd6cafabc4d7eeffe0' ;;
    darwin-arm64) printf '0054ea75a98db2ccdef3b304bf97aeb2b1fed201df13b30b68a19672a275199c' ;;
  esac
}

verify_checksum() {
  local archive="$1" expected="$2" actual
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$archive" | awk '{print $1}')"
  else
    require_command shasum
    actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
  fi
  if [[ "$actual" != "$expected" ]]; then
    printf 'checksum mismatch for %s\nexpected: %s\nactual:   %s\n' "$archive" "$expected" "$actual" >&2
    exit 1
  fi
}

download() {
  local url="$1" destination="$2" checksum="$3"
  curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location --retry 3 \
    --output "$destination" "$url"
  verify_checksum "$destination" "$checksum"
}

install_golangci_lint() {
  local os_name="$1" arch_name="$2" temp_dir="$3"
  local binary="$bin_dir/golangci-lint" archive_name archive_url archive_path extracted_dir
  if [[ -x "$binary" ]] && "$binary" version 2>/dev/null | grep -q "version $golangci_version "; then
    printf 'golangci-lint %s already installed\n' "$golangci_version"
    return
  fi

  archive_name="golangci-lint-${golangci_version}-${os_name}-${arch_name}.tar.gz"
  archive_url="https://github.com/golangci/golangci-lint/releases/download/v${golangci_version}/${archive_name}"
  archive_path="$temp_dir/$archive_name"
  extracted_dir="$temp_dir/${archive_name%.tar.gz}"
  download "$archive_url" "$archive_path" "$(golangci_checksum "$os_name" "$arch_name")"
  tar -xzf "$archive_path" -C "$temp_dir"
  install -m 0755 "$extracted_dir/golangci-lint" "$binary"
  printf 'installed golangci-lint %s\n' "$golangci_version"
}

install_shellcheck() {
  local os_name="$1" arch_name="$2" temp_dir="$3" shell_arch
  local binary="$bin_dir/shellcheck" archive_name archive_url archive_path extracted_dir
  if [[ -x "$binary" ]] && "$binary" --version 2>/dev/null | grep -q "version: $shellcheck_version"; then
    printf 'ShellCheck %s already installed\n' "$shellcheck_version"
    return
  fi

  case "$arch_name" in
    amd64) shell_arch="x86_64" ;;
    arm64) shell_arch="aarch64" ;;
  esac
  archive_name="shellcheck-v${shellcheck_version}.${os_name}.${shell_arch}.tar.gz"
  archive_url="https://github.com/koalaman/shellcheck/releases/download/v${shellcheck_version}/${archive_name}"
  archive_path="$temp_dir/$archive_name"
  extracted_dir="$temp_dir/shellcheck-v${shellcheck_version}"
  download "$archive_url" "$archive_path" "$(shellcheck_checksum "$os_name" "$arch_name")"
  tar -xzf "$archive_path" -C "$temp_dir"
  install -m 0755 "$extracted_dir/shellcheck" "$binary"
  printf 'installed ShellCheck %s\n' "$shellcheck_version"
}

install_goreleaser() {
  local os_name="$1" arch_name="$2" temp_path="$3" release_os release_arch
  local binary="$bin_dir/goreleaser" archive_name archive_url archive_path
  if [[ -x "$binary" ]] && "$binary" --version 2>/dev/null | grep -Eq "^GitVersion:[[:space:]]+$goreleaser_version$"; then
    printf 'GoReleaser %s already installed\n' "$goreleaser_version"
    return
  fi

  case "$os_name" in
    linux) release_os="Linux" ;;
    darwin) release_os="Darwin" ;;
  esac
  case "$arch_name" in
    amd64) release_arch="x86_64" ;;
    arm64) release_arch="arm64" ;;
  esac
  archive_name="goreleaser_${release_os}_${release_arch}.tar.gz"
  archive_url="https://github.com/goreleaser/goreleaser/releases/download/v${goreleaser_version}/${archive_name}"
  archive_path="$temp_path/$archive_name"
  download "$archive_url" "$archive_path" "$(goreleaser_checksum "$os_name" "$arch_name")"
  tar -xzf "$archive_path" -C "$temp_path" goreleaser
  install -m 0755 "$temp_path/goreleaser" "$binary"
  printf 'installed GoReleaser %s\n' "$goreleaser_version"
}

install_tombi() {
  local os_name="$1" arch_name="$2" temp_path="$3" target
  local binary="$bin_dir/tombi" archive_name archive_url archive_path extracted_dir
  if [[ -x "$binary" ]] && "$binary" --version 2>/dev/null | grep -q "^tombi $tombi_version "; then
    printf 'Tombi %s already installed\n' "$tombi_version"
    return
  fi

  case "$os_name-$arch_name" in
    linux-amd64) target="x86_64-unknown-linux-musl" ;;
    linux-arm64) target="aarch64-unknown-linux-musl" ;;
    darwin-amd64) target="x86_64-apple-darwin" ;;
    darwin-arm64) target="aarch64-apple-darwin" ;;
  esac
  archive_name="tombi-cli-${tombi_version}-${target}.tar.gz"
  archive_url="https://github.com/tombi-toml/tombi/releases/download/v${tombi_version}/${archive_name}"
  archive_path="$temp_path/$archive_name"
  extracted_dir="$temp_path/${archive_name%.tar.gz}"
  download "$archive_url" "$archive_path" "$(tombi_checksum "$os_name" "$arch_name")"
  tar -xzf "$archive_path" -C "$temp_path"
  install -m 0755 "$extracted_dir/tombi" "$binary"
  printf 'installed Tombi %s\n' "$tombi_version"
}

main() {
  local os_name arch_name temp_root
  require_command awk
  require_command curl
  require_command grep
  require_command install
  require_command tar

  os_name="$(platform_name)"
  arch_name="$(architecture_name)"
  temp_root="${TMPDIR:-/tmp}"
  temp_dir="$(mktemp -d "${temp_root%/}/nsk-tools.XXXXXX")"
  trap 'rm -rf -- "$temp_dir"' EXIT
  mkdir -p "$bin_dir"

  install_golangci_lint "$os_name" "$arch_name" "$temp_dir"
  install_shellcheck "$os_name" "$arch_name" "$temp_dir"
  install_goreleaser "$os_name" "$arch_name" "$temp_dir"
  install_tombi "$os_name" "$arch_name" "$temp_dir"
}

main "$@"
