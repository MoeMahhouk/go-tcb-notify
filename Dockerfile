FROM golang:1.24.5-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ingest-registry ./cmd/ingest-registry && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/fetch-pcs      ./cmd/fetch-pcs && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/evaluate-quotes ./cmd/evaluate-quotes && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/validate-quotes ./cmd/validate-quotes

# ---- runtime stage ----
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/ingest-registry /app/ingest-registry
COPY --from=builder /out/fetch-pcs      /app/fetch-pcs
COPY --from=builder /out/evaluate-quotes /app/evaluate-quotes
COPY --from=builder /out/validate-quotes /app/validate-quotes

# Default is no-op; docker-compose will override command per service.
CMD ["/bin/sh", "-lc", "echo 'Provide a command (ingest-registry | fetch-pcs | evaluate-quotes)' && sleep 1"]
