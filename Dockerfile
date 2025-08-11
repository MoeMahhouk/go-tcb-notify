FROM golang:1.24.5-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/go-tcb-notify ./cmd/go-tcb-notify

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /out/go-tcb-notify ./go-tcb-notify

EXPOSE 8080
CMD ["./go-tcb-notify"]