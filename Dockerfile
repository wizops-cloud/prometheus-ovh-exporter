FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /prometheus-ovh-exporter \
    ./cmd/prometheus-ovh-exporter/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    addgroup -S exporter && \
    adduser -S -G exporter exporter

COPY --from=builder /prometheus-ovh-exporter /usr/local/bin/prometheus-ovh-exporter

USER exporter

EXPOSE 9100

ENTRYPOINT ["prometheus-ovh-exporter"]
