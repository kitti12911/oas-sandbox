# ____________________ Go Command ____________________
air:
	air

tidy:
	go mod tidy

run:
	go run ./cmd/server/main.go

fmt:
	go fmt ./...

pretty:
	prettier --write "**/*.{md,markdown,yml,yaml,json,jsonc}"

format: fmt pretty

test:
	env CGO_ENABLED=1 go test --race -v ./...

cov:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

fix:
	go fix ./...

# ____________________ Generate Command ____________________
gen: gen-proto gen-openapi

gen-proto:
	rm -rf gen/grpc
	buf generate https://github.com/kitti12911/proto-sandbox.git --path common/v1 --path user/v1

gen-openapi:
	@echo "Huma generates OpenAPI at runtime: /openapi.json and /openapi.yaml"
