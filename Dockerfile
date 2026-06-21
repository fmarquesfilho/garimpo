FROM golang:1.25 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ❌ Remova "-tags gcp" para usar a implementação NopStore
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" \
      -o /out/garimpo-api ./cmd/garimpo-api

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpo-api /garimpo-api
ENTRYPOINT ["/garimpo-api"]
CMD ["-fonte", "shopee"]