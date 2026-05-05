package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var httpMethods = map[string]struct{}{
	"delete":  {},
	"get":     {},
	"head":    {},
	"options": {},
	"patch":   {},
	"post":    {},
	"put":     {},
	"trace":   {},
}

type openAPI struct {
	Paths map[string]map[string]operation `yaml:"paths"`
}

type operation struct {
	OperationID string `yaml:"operationId"`
	Summary     string `yaml:"summary"`
}

type endpoint struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
}

type change struct {
	ID          string         `json:"id"`
	Text        string         `json:"text"`
	Comment     string         `json:"comment"`
	Level       string         `json:"level"`
	Operation   string         `json:"operation"`
	OperationID string         `json:"operationId"`
	Path        string         `json:"path"`
	Section     string         `json:"section"`
	Attributes  map[string]any `json:"attributes"`
}

func main() {
	os.Exit(run())
}

func run() int {
	basePath := flag.String("base", "", "base OpenAPI file")
	revisionPath := flag.String("revision", "", "revision OpenAPI file")
	breakingPath := flag.String("breaking", "", "oasdiff breaking JSON file")
	flag.Parse()

	if *basePath == "" || *revisionPath == "" || *breakingPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "base, revision, and breaking are required")
		return 2
	}

	base, err := readOpenAPI(*basePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read base OpenAPI: %v\n", err)
		return 1
	}

	revision, err := readOpenAPI(*revisionPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read revision OpenAPI: %v\n", err)
		return 1
	}

	breaking, err := readBreakingChanges(*breakingPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read breaking changes: %v\n", err)
		return 1
	}

	writeReport(newEndpoints(base, revision), breaking)
	return 0
}

func readOpenAPI(path string) (openAPI, error) {
	// #nosec G304 -- CI passes generated OpenAPI file paths explicitly.
	body, err := os.ReadFile(path)
	if err != nil {
		return openAPI{}, fmt.Errorf("read file: %w", err)
	}

	var spec openAPI
	if err := yaml.Unmarshal(body, &spec); err != nil {
		return openAPI{}, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return spec, nil
}

func readBreakingChanges(path string) ([]change, error) {
	// #nosec G304 -- CI passes the oasdiff JSON report path explicitly.
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 {
		return nil, nil
	}

	var wrapped struct {
		Changes []change `json:"changes"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Changes != nil {
		return wrapped.Changes, nil
	}

	var changes []change
	if err := json.Unmarshal(body, &changes); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	return changes, nil
}

func newEndpoints(base, revision openAPI) []endpoint {
	baseEndpoints := endpointSet(base)
	var added []endpoint

	for path, pathItem := range revision.Paths {
		for method, op := range pathItem {
			method = strings.ToLower(method)
			if _, ok := httpMethods[method]; !ok {
				continue
			}

			key := method + " " + path
			if _, ok := baseEndpoints[key]; ok {
				continue
			}

			added = append(added, endpoint{
				Method:      strings.ToUpper(method),
				Path:        path,
				OperationID: op.OperationID,
				Summary:     op.Summary,
			})
		}
	}

	sort.Slice(added, func(i, j int) bool {
		if added[i].Path == added[j].Path {
			return added[i].Method < added[j].Method
		}

		return added[i].Path < added[j].Path
	})

	return added
}

func endpointSet(spec openAPI) map[string]struct{} {
	endpoints := make(map[string]struct{})
	for path, pathItem := range spec.Paths {
		for method := range pathItem {
			method = strings.ToLower(method)
			if _, ok := httpMethods[method]; !ok {
				continue
			}

			endpoints[method+" "+path] = struct{}{}
		}
	}

	return endpoints
}

func writeReport(added []endpoint, breaking []change) {
	fmt.Println("### New APIs")
	fmt.Println()
	if len(added) == 0 {
		fmt.Println("No new API endpoints.")
	} else {
		fmt.Println("| API | Reason |")
		fmt.Println("| --- | --- |")
		for _, endpoint := range added {
			reason := "Endpoint added"
			if endpoint.Summary != "" {
				reason += ": " + endpoint.Summary
			} else if endpoint.OperationID != "" {
				reason += ": `" + endpoint.OperationID + "`"
			}

			fmt.Printf("| `%s %s` | %s |\n", endpoint.Method, endpoint.Path, escapeTable(reason))
		}
	}

	fmt.Println()
	fmt.Println("### Breaking Changes")
	fmt.Println()
	if len(breaking) == 0 {
		fmt.Println("No breaking OpenAPI changes.")
		return
	}

	fmt.Println("| API | Reason | Rule |")
	fmt.Println("| --- | --- | --- |")
	for _, change := range breaking {
		api := changeAPI(change)
		reason := change.Text
		if reason == "" {
			reason = change.Comment
		}
		if reason == "" {
			reason = "Breaking change detected"
		}

		fmt.Printf("| `%s` | %s | `%s` |\n", api, escapeTable(reason), change.ID)
	}
}

func changeAPI(change change) string {
	method := strings.ToUpper(change.Operation)
	path := change.Path

	if method == "" {
		method = strings.ToUpper(stringAttribute(change.Attributes, "operation"))
	}
	if path == "" {
		path = stringAttribute(change.Attributes, "path")
	}
	if method != "" && path != "" {
		return method + " " + path
	}
	if path != "" {
		return path
	}
	if change.Section != "" {
		return change.Section
	}
	if change.OperationID != "" {
		return change.OperationID
	}

	return "OpenAPI document"
}

func stringAttribute(attributes map[string]any, key string) string {
	if attributes == nil {
		return ""
	}

	value, ok := attributes[key]
	if !ok {
		return ""
	}

	text, ok := value.(string)
	if !ok {
		return ""
	}

	return text
}

func escapeTable(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.ReplaceAll(value, "|", "\\|")
}
