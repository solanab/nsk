set shell := ["bash", "-euo", "pipefail", "-c"]

repo := justfile_directory()
tool_bin := repo + "/runtime/tools/bin"
golangci_lint := tool_bin + "/golangci-lint"
goreleaser := tool_bin + "/goreleaser"
shellcheck_bin := tool_bin + "/shellcheck"
tombi_bin := tool_bin + "/tombi"
source_lines_bin := tool_bin + "/source-lines"
source_lines_version := "0.2.0"
version := env_var_or_default("NSK_VERSION", "0.0.0-dev")

default: help

alias h := help
alias i := install
alias l := lint
alias f := fmt
alias c := check
alias co := check-online
alias t := test
alias cov := coverage
alias a := audit
alias b := build
alias bg := bg-start

help:
    @cat scripts/help.txt

install: install-tools install-source-lines

install-tools:
    ./scripts/install-tools.sh

install-source-lines:
    ./scripts/install-source-lines.sh {{ quote(source_lines_version) }}

hooks-install:
    command -v prek >/dev/null 2>&1 || { echo 'prek is optional; install it with `uv tool install prek` or `brew install prek`' >&2; exit 1; }
    prek install

hooks-check:
    command -v prek >/dev/null 2>&1 || { echo 'prek is optional; install it with `uv tool install prek` or `brew install prek`' >&2; exit 1; }
    prek run --all-files

markdown-format:
    command -v prek >/dev/null 2>&1 || { echo 'prek is optional; install it with `uv tool install prek` or `brew install prek`' >&2; exit 1; }
    prek exec dprint-markdown -- dprint fmt --allow-no-files

markdown-lint:
    command -v prek >/dev/null 2>&1 || { echo 'prek is optional; install it with `uv tool install prek` or `brew install prek`' >&2; exit 1; }
    prek exec rumdl-markdown -- rumdl check --deny-config-warnings

justfile-check:
    just --unstable --fmt --check

fmt: go-fmt shell-fmt toml-fmt yaml-fmt

go-fmt:
    test -x {{ quote(golangci_lint) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(golangci_lint) }} fmt

shell-fmt:
    go tool shfmt -w -i 2 -ci scripts/*.sh

toml-fmt:
    test -x {{ quote(tombi_bin) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(tombi_bin) }} format

yaml-fmt:
    go tool yamlfmt -gitignore_excludes .

fmt-check: go-fmt-check shell-fmt-check toml-fmt-check yaml-lint

go-fmt-check:
    test -x {{ quote(golangci_lint) }} || { echo 'run `just install` first' >&2; exit 1; }
    output="$({{ quote(golangci_lint) }} fmt --diff)"; if [[ -n "$output" ]]; then printf '%s\n' "$output"; exit 1; fi

shell-fmt-check:
    output="$(go tool shfmt -d -i 2 -ci scripts/*.sh)"; if [[ -n "$output" ]]; then printf '%s\n' "$output"; exit 1; fi

toml-fmt-check:
    test -x {{ quote(tombi_bin) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(tombi_bin) }} format --check

toml-lint:
    test -x {{ quote(tombi_bin) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(tombi_bin) }} lint --error-on-warnings

yaml-lint:
    go tool yamlfmt -lint -gitignore_excludes .

lint:
    test -x {{ quote(golangci_lint) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(golangci_lint) }} config verify
    {{ quote(golangci_lint) }} run

test *args:
    go test ./... {{ args }}

race *args:
    go test -race ./... {{ args }}

coverage:
    ./scripts/coverage.sh

vet:
    go vet ./...

mod-check:
    go mod tidy -diff

shellcheck:
    test -x {{ quote(shellcheck_bin) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(shellcheck_bin) }} scripts/*.sh

actionlint:
    go tool actionlint

source-lines:
    test -x {{ quote(source_lines_bin) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(source_lines_bin) }} --config {{ quote(repo + "/source-lines.toml") }} {{ quote(repo) }}

check: justfile-check fmt-check mod-check lint coverage vet toml-lint source-lines shellcheck actionlint

check-online: check audit

audit:
    go tool govulncheck ./...

build:
    mkdir -p dist
    go build -trimpath -ldflags "-s -w -X github.com/solanab/nsk/internal/app.version={{ version }}" -o dist/nsk ./cmd/nsk

release-snapshot:
    test -x {{ quote(goreleaser) }} || { echo 'run `just install` first' >&2; exit 1; }
    {{ quote(goreleaser) }} release --snapshot --clean --skip=publish

bg-start name command:
    ./scripts/background.sh start {{ quote(name) }} {{ quote(command) }}

bg-stop name timeout="10":
    ./scripts/background.sh stop {{ quote(name) }} {{ quote(timeout) }}

bg-status name:
    ./scripts/background.sh status {{ quote(name) }}

bg-list:
    ./scripts/background.sh list

bg-logs name lines="80":
    ./scripts/background.sh logs {{ quote(name) }} {{ quote(lines) }}

clean:
    for path in dist build runtime coverage.out; do if [[ -e "$path" ]]; then trash "$path"; fi; done
