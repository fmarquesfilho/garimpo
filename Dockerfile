# ── 1. Build do frontend ────────────────────────────────────────────────────
FROM node:24-alpine AS web
WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts

COPY web/ .
RUN npm run build

# ── 1b. Build do site de documentação ──────────────────────────────────────
FROM node:24-alpine AS docs
WORKDIR /docs-site

COPY docs-site/package.json docs-site/package-lock.json ./
RUN npm ci --ignore-scripts

COPY docs-site/ .
RUN npm run build

# ── 2. Build do backend Go ──────────────────────────────────────────────────
FROM golang:1.26.4 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags gcp -ldflags="-s -w" -o /out/garimpo-api ./cmd/garimpo-api

# ── 3. Imagem final (binário + frontend estático + docs) ────────────────────
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/garimpo-api /garimpo-api
COPY --from=build /src/api/openapi.yaml /api/openapi.yaml
COPY --from=web /web/build /web
COPY --from=docs /docs-site/dist /docs-site/dist
ENV WEB_DIR=/web
ENV DOCS_DIR=/docs-site/dist
ENTRYPOINT ["/garimpo-api"]
CMD ["-fonte", "shopee"]
