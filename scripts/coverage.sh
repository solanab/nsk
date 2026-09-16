#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

main_packages=()
covered_packages=()
while IFS=$'\t' read -r name package; do
  if [[ "$name" == "main" ]]; then
    main_packages+=("$package")
  else
    covered_packages+=("$package")
  fi
done < <(go list -f '{{.Name}}{{"\t"}}{{.ImportPath}}' ./...)

if [[ ${#covered_packages[@]} -eq 0 ]]; then
  printf 'no non-main Go packages found for coverage\n' >&2
  exit 1
fi

if [[ ${#main_packages[@]} -gt 0 ]]; then
  go test "${main_packages[@]}"
fi
go test -covermode=atomic -coverpkg=./... -coverprofile=coverage.out "${covered_packages[@]}"
go tool go-test-coverage --config=.testcoverage.yml
