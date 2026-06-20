# Build do binário com a tag gcp (inclui o BigQueryStore).
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum* ./
RUN go mod download || true
COPY . .
# Resolve a dependência do BigQuery e compila estático para o Cloud Run.
RUN go get cloud.google.com/go/bigquery && \
    CGO_ENABLED=0 GOOS=linux go build -tags gcp -ldflags "-s -w" \
      -o /out/garimpo-api ./cmd/garimpo-api

# Runtime mínimo e seguro.
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpo-api /garimpo-api
# Cloud Run injeta PORT; o binário escuta nela. A fonte padrão aqui é Shopee.
ENTRYPOINT ["/garimpo-api"]
CMD ["-fonte", "shopee"]
