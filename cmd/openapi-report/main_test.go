package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestAddedAndBumpedEndpointsClassifiesVersionBumps(t *testing.T) {
	base := openAPI{
		Paths: map[string]map[string]operation{
			"/health": {
				"get": {},
			},
			"/v1/users": {
				"get": {},
			},
			"/v1/users/{id}": {
				"get": {},
			},
		},
	}
	revision := openAPI{
		Paths: map[string]map[string]operation{
			"/health": {
				"get": {},
			},
			"/v1/projects": {
				"post": {
					Summary: "Create project",
				},
			},
			"/v1/users": {
				"get": {},
			},
			"/v2/users": {
				"get": {
					Summary: "List users v2",
				},
			},
			"/v2/users/{id}": {
				"get": {
					OperationID: "get-user-v2",
				},
			},
		},
	}

	added, bumped := addedAndBumpedEndpoints(base, revision)

	if len(added) != 1 {
		t.Fatalf("expected 1 added endpoint, got %d: %#v", len(added), added)
	}
	if added[0].Method != "POST" || added[0].Path != "/v1/projects" {
		t.Fatalf("unexpected added endpoint: %#v", added[0])
	}

	if len(bumped) != 2 {
		t.Fatalf("expected 2 version bumps, got %d: %#v", len(bumped), bumped)
	}
	if bumped[0].Method != "GET" || bumped[0].FromPath != "/v1/users" || bumped[0].ToPath != "/v2/users" {
		t.Fatalf("unexpected first version bump: %#v", bumped[0])
	}
	if bumped[1].Method != "GET" || bumped[1].FromPath != "/v1/users/{id}" || bumped[1].ToPath != "/v2/users/{id}" {
		t.Fatalf("unexpected second version bump: %#v", bumped[1])
	}
}

func TestReadBreakingChangesAcceptsNumericLevel(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "breaking-*.json")
	if err != nil {
		t.Fatalf("create breaking json: %v", err)
	}

	_, err = file.WriteString(`[{"id":"api-path-removed-without-deprecation","text":"api path removed without deprecation","level":3,"operation":"GET","operationId":"get-v1-users-legacy","path":"/v1/users/legacy","section":"paths"}]`)
	if err != nil {
		t.Fatalf("write breaking json: %v", err)
	}
	if closeErr := file.Close(); closeErr != nil {
		t.Fatalf("close breaking json: %v", closeErr)
	}

	changes, err := readBreakingChanges(file.Name())
	if err != nil {
		t.Fatalf("read breaking changes: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 breaking change, got %d", len(changes))
	}
	if api := changeAPI(changes[0]); api != "GET /v1/users/legacy" {
		t.Fatalf("unexpected breaking API: %s", api)
	}
}

func TestWriteReportBreakingModeOnlyShowsBreakingChanges(t *testing.T) {
	output := captureStdout(t, func() {
		writeReport(
			reportModeBreaking,
			[]endpoint{{Method: "POST", Path: "/v1/projects"}},
			[]versionBump{{Method: "GET", FromPath: "/v1/users", ToPath: "/v2/users"}},
			[]change{{
				ID:        "api-removed",
				Text:      "endpoint removed",
				Operation: "get",
				Path:      "/v1/users",
			}},
		)
	})

	if strings.Contains(output, "### New APIs") {
		t.Fatalf("breaking mode should not include new APIs: %s", output)
	}
	if strings.Contains(output, "### API Version Bumps") {
		t.Fatalf("breaking mode should not include version bumps: %s", output)
	}
	if !strings.Contains(output, "### Breaking Changes") || !strings.Contains(output, "`GET /v1/users`") {
		t.Fatalf("breaking mode did not include breaking API details: %s", output)
	}
}

func TestWriteJSONReport(t *testing.T) {
	path := t.TempDir() + "/openapi-report.json"
	report := newReportData(
		reportModeMain,
		[]endpoint{{
			Method:      "POST",
			Path:        "/v1/projects",
			OperationID: "post-v1-projects",
			Summary:     "Create project",
		}},
		[]versionBump{{
			Method:      "GET",
			FromPath:    "/v1/users",
			ToPath:      "/v2/users",
			OperationID: "get-v2-users",
			Summary:     "List users v2",
		}},
		[]change{{
			ID:        "api-path-removed-without-deprecation",
			Text:      "api path removed without deprecation",
			Level:     float64(3),
			Operation: "GET",
			Path:      "/v1/users/legacy",
		}},
	)

	if err := writeJSONReport(path, report); err != nil {
		t.Fatalf("write JSON report: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}

	var got reportData
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal JSON report: %v", err)
	}
	if got.SchemaVersion != 1 || got.Mode != reportModeMain {
		t.Fatalf("unexpected report metadata: %#v", got)
	}
	if got.Counts.NewAPIs != 1 || got.Counts.APIVersionBumps != 1 || got.Counts.BreakingChanges != 1 {
		t.Fatalf("unexpected report counts: %#v", got.Counts)
	}
	if got.NewAPIs[0].Reason != "Endpoint added: Create project" {
		t.Fatalf("unexpected new API reason: %s", got.NewAPIs[0].Reason)
	}
	if got.APIVersionBumps[0].FromPath != "/v1/users" || got.APIVersionBumps[0].ToPath != "/v2/users" {
		t.Fatalf("unexpected version bump: %#v", got.APIVersionBumps[0])
	}
	if got.BreakingChanges[0].API != "GET /v1/users/legacy" {
		t.Fatalf("unexpected breaking API: %#v", got.BreakingChanges[0])
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}

	os.Stdout = writer
	fn()
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("close stdout writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	return string(output)
}
