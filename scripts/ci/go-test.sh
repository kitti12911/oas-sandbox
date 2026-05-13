#!/usr/bin/env sh
set -eu

coverage_profile="${GO_COVERAGE_PROFILE:-coverage.out}"
race="${GO_TEST_RACE:-true}"
cgo="${GO_TEST_CGO:-${race}}"

if [ "${cgo}" = "true" ] && ! command -v gcc >/dev/null 2>&1; then
	echo "CGO-enabled tests require a C compiler in the toolchain image." >&2
	exit 2
fi

packages="$(go list -f '{{.Dir}}' ./... | grep -v '/gen/' | sed "s#^${PWD}#.#")"
if [ -z "${packages}" ]; then
	echo "no Go packages to test"
	exit 0
fi

set -- go test -buildvcs=false -coverprofile="${coverage_profile}" -covermode=atomic

if [ "${race}" = "true" ]; then
	set -- "$@" -race
fi

if [ "${cgo}" = "true" ]; then
	cgo_enabled=1
else
	cgo_enabled=0
fi

# Package paths are emitted by go list and do not contain whitespace in this repository.
# shellcheck disable=SC2086
env CGO_ENABLED="${cgo_enabled}" "$@" ${packages}

tmp_profile="$(mktemp)"
awk 'NR == 1 || ($1 !~ /\/gen\// && $1 !~ /(_gen|_generated)\.go:/)' "${coverage_profile}" >"${tmp_profile}"
mv "${tmp_profile}" "${coverage_profile}"

go tool cover -func="${coverage_profile}" | tail -n 1
