# ── 1. Build do frontend ────────────────────────────────────────────────────
FROM node:24-alpine AS web
WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts

COPY web/ .
RUN npm run build

# ── 2. Build do backend Go ──────────────────────────────────────────────────
FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags gcp -ldflags="-s -w" -o /out/garimpo-api ./cmd/garimpo-api

# ── 3. Imagem final (binário + frontend estático) ───────────────────────────
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpo-api /garimpo-api
COPY --from=build /src/docs/openapi.yaml /docs/openapi.yaml
COPY --from=web /web/build /web
ENV WEB_DIR=/web
ENTRYPOINT ["/garimpo-api"]
CMD ["-fonte", "shopee"]
