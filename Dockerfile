FROM zot.kittiaccess.work/kitti12911/image-toolchain@sha256:47355f96a059465947c38aa956da1c4502c11d1e8f53eb2c8b3980ba58983d42 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY Makefile buf.gen.yaml ./
COPY cmd ./cmd
COPY internal ./internal

RUN make gen

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go build -trimpath -ldflags="-s -w" -o /out/oas-sandbox ./cmd/server

FROM alpine:3.22@sha256:310c62b5e7ca5b08167e4384c68db0fd2905dd9c7493756d356e893909057601

RUN apk add --no-cache ca-certificates tzdata \
	&& addgroup -S app \
	&& adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/oas-sandbox /app/oas-sandbox
COPY --chown=app:app config.example.yml /app/config.yml

USER app

EXPOSE 8080

ENTRYPOINT ["/app/oas-sandbox"]
