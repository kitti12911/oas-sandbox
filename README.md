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
- structured logs from [`lib-util`](https://github.com/kitti12911/lib-util)
- tracing and profiling from
  [`lib-monitor`](https://github.com/kitti12911/lib-monitor)
- graceful shutdown

## requirements

- go 1.26 or higher
- [buf](https://buf.build/) for protobuf generation

Optional:

- [prettier](https://prettier.io/) for Markdown, YAML, JSON, and JSONC formatting

## project structure

```bash
oas-sandbox/
├── cmd/
│   └── server/                 # HTTP server entrypoint
├── gen/
│   └── grpc/                   # generated protobuf clients
├── internal/
│   ├── api/
│   │   └── system/             # health endpoint
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

## generate code

```bash
make gen
```

`make gen` runs protobuf generation from
[`github.com/kitti12911/proto-sandbox`](https://github.com/kitti12911/proto-sandbox).
Huma generates OpenAPI at runtime from Go route and DTO types.

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

## API

Implemented routes:

- `GET /health`

## available commands

```bash
make tidy
make fmt
make pretty
make format
make test
make cov
make gen
make run
```
