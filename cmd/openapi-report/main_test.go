package main

import (
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
