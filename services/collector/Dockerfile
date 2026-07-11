# Garimpei Unified Collector
# Context: repo root (docker build -f services/collector/Dockerfile .)
# Substitui: services/collector-amazon/Dockerfile + services/coupon-collector/Dockerfile
# Ref: docs/decisoes/0018-collector-unificado.md
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags gcp -ldflags="-s -w" -o /out/garimpei-collector ./services/collector

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpei-collector /garimpei-collector
COPY services/collector/collector.yaml /etc/garimpei/collector.yaml

ENV COLLECTOR_CONFIG=/etc/garimpei/collector.yaml
EXPOSE 50051
ENTRYPOINT ["/garimpei-collector"]
