FROM golang:1.25 AS build
WORKDIR /src

# Copia apenas os módulos primeiro (melhor cache)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags gcp -ldflags="-s -w" \
      -o /out/garimpo-api ./cmd/garimpo-api

# Runtime seguro
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpo-api /garimpo-api
ENTRYPOINT ["/garimpo-api"]
CMD ["-fonte", "shopee"]