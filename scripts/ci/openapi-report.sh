#!/usr/bin/env sh
set -eu

repo_dir="${CI_PROJECT_DIR:-$(pwd)}"
cd "${repo_dir}"

: "${BASE_SHA:?BASE_SHA is required}"

revision_sha="$(git rev-parse "${OPENAPI_REVISION_SHA:-${GITHUB_SHA:-HEAD}}")"
base_sha="$(git rev-parse "${BASE_SHA}")"
original_ref="$(git symbolic-ref --quiet --short HEAD || git rev-parse HEAD)"
report_mode="${OPENAPI_REPORT_MODE:-${REPORT_MODE:-main}}"
temp_dir="${OPENAPI_TEMP_DIR:-${RUNNER_TEMP:-/tmp}}"
base_spec="${OPENAPI_BASE_SPEC:-${temp_dir}/openapi-base.yaml}"
revision_spec="${OPENAPI_REVISION_SPEC:-${temp_dir}/openapi-revision.yaml}"
breaking_report="${OPENAPI_BREAKING_REPORT:-${temp_dir}/openapi-breaking.json}"
breaking_error="${OPENAPI_BREAKING_ERROR:-${temp_dir}/openapi-breaking.err}"
allow_breaking_message="${OPENAPI_ALLOW_BREAKING_MESSAGE:-[allow-breaking-api]}"

restore_checkout() {
	git checkout "${original_ref}" >/dev/null 2>&1 || true
}
trap restore_checkout EXIT INT TERM

generate_openapi() {
	output_file="$1"

	rm -rf gen/grpc
	if grep -q '^inputs:' buf.gen.yaml; then
		buf generate
	else
		buf generate "https://github.com/kitti12911/proto-sandbox.git#tag=v1.9.0" --path common/v1 --path user/v1
	fi
	go run ./cmd/gen-patch
	make gen-openapi >"${output_file}"
}

git checkout "${revision_sha}"
generate_openapi "${revision_spec}"

git checkout "${base_sha}"
generate_openapi "${base_spec}"

git checkout "${revision_sha}"

allow_breaking="${OPENAPI_ALLOW_BREAKING:-}"
if [ -z "${allow_breaking}" ]; then
	allow_breaking=false
	if [ "${GITHUB_EVENT_NAME:-}" = "push" ] &&
		git log -1 --pretty=%B | grep -Fq "${allow_breaking_message}"; then
		allow_breaking=true
	fi
fi

set +e
oasdiff breaking --format json --fail-on ERR \
	"${base_spec}" \
	"${revision_spec}" \
	>"${breaking_report}" \
	2>"${breaking_error}"
breaking_status="$?"
set -e

report_file="$(mktemp)"
{
	echo "## OpenAPI changes"
	echo
	echo "Compared base \`${base_sha}\` to revision \`${revision_sha}\`."
	echo
	echo "Report mode: \`${report_mode}\`."
	echo
	go run ./cmd/openapi-report \
		-base "${base_spec}" \
		-revision "${revision_spec}" \
		-breaking "${breaking_report}" \
		-mode "${report_mode}"
} >"${report_file}"

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
	cat "${report_file}" >>"${GITHUB_STEP_SUMMARY}"
else
	cat "${report_file}"
fi

breaking_json="$(tr -d '[:space:]' <"${breaking_report}")"
breaking_detected=false
if [ "${breaking_json}" != "" ] && [ "${breaking_json}" != "[]" ] && [ "${breaking_json}" != "{\"changes\":[]}" ]; then
	breaking_detected=true
fi

if [ "${breaking_detected}" = "true" ] && [ "${allow_breaking}" = "true" ]; then
	{
		echo
		echo "> Breaking changes were detected, but release is allowed because \`${allow_breaking_message}\` is present in the head commit message."
	} >>"${GITHUB_STEP_SUMMARY:-/dev/stdout}"
	exit 0
fi

if [ "${breaking_detected}" = "true" ]; then
	echo "OpenAPI breaking changes detected." >&2
	cat "${breaking_error}" >&2
	exit 1
fi

if [ "${breaking_status}" -ne 0 ]; then
	cat "${breaking_error}" >&2
	exit "${breaking_status}"
fi
