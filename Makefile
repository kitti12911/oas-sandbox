# Files outside the business logic surface (main, generators, route
# registration in *api.go, deps wiring) are dropped from coverage so the
# reported % reflects code worth testing. The server package stays in: its
# middleware (access log, gzip, trace extraction) is real logic. Patterns are
# awk regexes matched against the file:line column of coverage.out.
GO_COVERAGE_EXCLUDE_REGEX = /cmd/|/api\.go:|/internal/api/system/|/internal/api/deps\.go:|/internal/api/helpers\.go:

# ____________________ Go Command ____________________
air:
	air

tidy:
	go mod tidy

run:
	go run ./cmd/server/main.go

lint: vet golangci-lint markdownlint

vet:
	go vet ./...

golangci-lint:
	golangci-lint run --timeout=5m

markdownlint:
	markdownlint-cli2

fmt:
	go fmt ./...

pretty:
	prettier --write "**/*.{md,markdown,yml,yaml,json,jsonc}"

format: fmt pretty

test:
	env CGO_ENABLED=1 go test --race -v ./...

ci-test:
	GO_COVERAGE_EXCLUDE_REGEX='$(GO_COVERAGE_EXCLUDE_REGEX)' ./scripts/ci/go-test.sh

cov:
	GO_COVERAGE_EXCLUDE_REGEX='$(GO_COVERAGE_EXCLUDE_REGEX)' ./scripts/ci/go-test.sh
	go tool cover -html=coverage.out

fix:
	go fix ./...

# ____________________ Generate Command ____________________
gen: gen-proto gen-patch

gen-proto:
	rm -rf gen/grpc
	buf generate

gen-patch:
	go run ./cmd/gen-patch

gen-openapi:
	@go run ./cmd/gen-oas

# pipeline triggered #1
