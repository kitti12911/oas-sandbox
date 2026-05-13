# oas-sandbox

OpenAPI sandbox for homelab API experiments. It exposes REST endpoints with
[Huma](https://huma.rocks/) and will add gRPC-backed APIs step by step using
generated clients from
[`proto-sandbox`](https://github.com/kitti12911/proto-sandbox).

## features

- Huma HTTP API with generated OpenAPI 3.1
- embedded Swagger UI docs at `/docs`
- OpenAPI documents at `/openapi.json` and `/openapi.yaml`
- downloadable OpenAPI documents at `/openapi.json/download` and
  `/openapi.yaml/download`
- `/health` operational endpoint
- `GET /v1/users` REST gateway to
  [`grpc-sandbox`](https://github.com/kitti12911/grpc-sandbox)
- `GET /v1/users/{id}` REST gateway to
  [`grpc-sandbox`](https://github.com/kitti12911/grpc-sandbox)
- `POST /v1/worker/jobs` REST gateway to submit background jobs through
  [`grpc-sandbox`](https://github.com/kitti12911/grpc-sandbox)
- structured logs from [`lib-util`](https://github.com/kitti12911/lib-util)
- tracing and profiling from
  [`lib-monitor`](https://github.com/kitti12911/lib-monitor)
- graceful shutdown

## requirements

- go 1.26 or higher
- [buf](https://buf.build/) for protobuf generation
- running [`grpc-sandbox`](https://github.com/kitti12911/grpc-sandbox) for user
  API calls

Optional:

- [prettier](https://prettier.io/) for Markdown, YAML, JSON, and JSONC formatting

## ci commands

reusable CI entrypoints live in `scripts/ci/` so GitHub Actions and GitLab CI
can call the same commands with provider-specific orchestration around them.

| command                                            | purpose                                       |
| -------------------------------------------------- | --------------------------------------------- |
| `./scripts/ci/generate-code.sh`                    | generate protobuf and PATCH code              |
| `./scripts/ci/openapi-report.sh`                   | compare and report OpenAPI changes            |
| `./scripts/ci/go-lint.sh`                          | run `go vet` and `golangci-lint`              |
| `./scripts/ci/go-test.sh`                          | run tests with filtered coverage              |
| `./scripts/ci/markdownlint.sh`                     | run Markdown linting                          |
| `./scripts/ci/security-scan.sh`                    | run `govulncheck` and Semgrep                 |
| `./scripts/ci/supply-chain-scan.sh`                | run Trivy and Gitleaks                        |
| `./scripts/ci/semantic-release-plan.sh`            | preview the next semantic release             |
| `./scripts/ci/semantic-release-publish.sh`         | publish the semantic release                  |
| `./scripts/ci/fast-forward-prerelease-branches.sh` | fast-forward `uat` and `develop` after `main` |
| `./scripts/ci/update-helm-image-values.sh`         | update homelab GitOps image values            |

GitHub Actions uses `TOOLCHAIN_REGISTRY` and `TOOLCHAIN_IMAGE_NAMESPACE` to
resolve shared CI toolchain images, and `IMAGE_REGISTRY` plus `IMAGE_NAMESPACE`
to publish the application image. GitLab should map its CI variables and image
credentials to the same script inputs instead of duplicating command logic.
The `homelab-devops` values update in `.github/workflows/go-ci.yml` is
GitHub-specific homelab orchestration, not part of the portable script contract.
The prerelease branch fast-forward helper is also GitHub-specific because it
pushes through a GitHub App token.
GitLab deployments can use a different project, folder layout, or deployment
tool by calling the same `scripts/ci` build/release helpers and adding its own
deploy job. `DEPLOY_IMAGE_REGISTRY` and `DEPLOY_IMAGE_NAMESPACE` only affect the
homelab GitOps values update and can be omitted outside that workflow.

`GO_TEST_RACE=true` or `GO_TEST_CGO=true` requires a C compiler in the selected
toolchain image. `oas-sandbox` sets `GO_TEST_RACE=false` in GitHub Actions while
using `image-toolchain` v1.1.0 because that image does not include one.

## project structure

```bash
oas-sandbox/
├── cmd/
│   └── server/                 # HTTP server entrypoint
├── gen/
│   └── grpc/                   # generated protobuf clients
├── internal/
│   ├── api/
│   │   ├── system/             # health endpoint
│   │   ├── users/              # user REST resource
│   │   └── worker/             # worker job REST resource
│   ├── config/                 # config structs
│   └── server/                 # Huma HTTP server and Swagger UI assets
├── buf.gen.yaml
├── config.example.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## configuration

Copy `config.example.yml` to `config.yml` and adjust local values:

```bash
cp config.example.yml config.yml
```

Important sections:

- `service`: service name, HTTP port, and shutdown timeout
- `logging`: slog level and trace id injection
- `tracing`: OTLP exporter settings
- `profiling`: Pyroscope settings
- `user_service`: gRPC address for `grpc-sandbox`

## generate code

```bash
make gen
```

`make gen` runs protobuf generation from
[`github.com/kitti12911/proto-sandbox`](https://github.com/kitti12911/proto-sandbox).
It also runs `gen-patch`, which generates tri-state PATCH mappers from
`//openapi:patch` markers. Huma generates OpenAPI at runtime from Go route and
DTO types.

## run locally

```bash
make run
```

Then open:

- <http://localhost:8080/docs>
- <http://localhost:8080/openapi.json>
- <http://localhost:8080/openapi.yaml>
- <http://localhost:8080/openapi.json/download>
- <http://localhost:8080/openapi.yaml/download>
- <http://localhost:8080/health>
- <http://localhost:8080/v1/users?page=1&pageSize=10>
- <http://localhost:8080/v1/users/0198f8f0-0000-7000-8000-000000000001>
- <http://localhost:8080/v1/worker/jobs>

## API

Implemented routes:

- `GET /health`
- `GET /v1/users`
- `POST /v1/users` for create
- `POST /v1/users/search` for advanced list/search
- `GET /v1/users/{id}`
- `PUT /v1/users/{id}`
- `PATCH /v1/users/{id}`
- `POST /v1/worker/jobs`

`GET /v1/users` supports query parameters for the common gRPC list request:

- `page` and `pageSize` for pagination
- `filterCol`, `filterOp`, `filterVal`, and `filterVals` for one filter clause
- `orderBy` and `order` for one order clause

Use `POST /v1/users/search` when the request needs multiple filters or order
clauses:

```json
{
    "pagination": {
        "page": 1,
        "pageSize": 10
    },
    "filters": [
        {
            "col": "username",
            "op": "like_ci",
            "val": "kit"
        },
        {
            "col": "status",
            "op": "in",
            "vals": ["active", "pending"]
        }
    ],
    "orderBy": [
        {
            "col": "username",
            "order": "desc"
        },
        {
            "col": "createdAt",
            "order": "asc"
        }
    ]
}
```

`PATCH /v1/users/{id}` builds its field mask from the JSON body:

- omitted fields are ignored
- `null` fields are written as null
- fields with values are updated

Example patch:

```json
{
    "displayName": null,
    "profile": {
        "firstName": "Patched",
        "address": {
            "city": "Phuket"
        }
    }
}
```

`POST /v1/worker/jobs` submits a background job through `grpc-sandbox`, which
publishes it to `worker-sandbox`:

```json
{
    "id": "job-1",
    "type": "debug.print",
    "payload": {
        "message": "hello"
    }
}
```

## contract checks

CI generates and compares OpenAPI documents for API changes. Pull requests fail
when `oasdiff` detects breaking OpenAPI changes. On a protected branch push, an
intentional breaking release can continue only when the head commit message
contains `[allow-breaking-api]`.

Example intentional breaking release message:

```text
feat!: rename user response field [allow-breaking-api]
```

Prefer adding a new versioned route, such as `/v2/users`, over breaking an
existing route in place. Use the bypass only after the break is reviewed and
documented in release notes for API consumers.

## available commands

| Command            | Description                                     |
| ------------------ | ----------------------------------------------- |
| `make air`         | Run the service with Air live reload            |
| `make tidy`        | Run `go mod tidy`                               |
| `make run`         | Start the HTTP server locally                   |
| `make lint`        | Run Go and Markdown linting                     |
| `make fmt`         | Format Go code with `go fmt`                    |
| `make pretty`      | Format Markdown, YAML, JSON, and JSONC          |
| `make format`      | Run Go and document/config formatting           |
| `make test`        | Run tests with the race detector                |
| `make cov`         | Generate and open an HTML coverage report       |
| `make fix`         | Apply standard Go source rewrites with `go fix` |
| `make gen`         | Generate protobuf clients and PATCH helpers     |
| `make gen-proto`   | Generate protobuf clients from `proto-sandbox`  |
| `make gen-patch`   | Generate OpenAPI PATCH helper code              |
| `make gen-openapi` | Print the generated OpenAPI document to stdout  |
