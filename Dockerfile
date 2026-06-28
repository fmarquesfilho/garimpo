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
# Sincronizar docs canônicos (fonte única: docs/) para content do Starlight
COPY docs/ /docs-src/
COPY scripts/sync-docs-to-site.sh /scripts/sync-docs-to-site.sh
RUN chmod +x /scripts/sync-docs-to-site.sh
ENV ROOT=/
RUN /bin/sh -c '\
  SRC=/docs-src && DST=/docs-site/src/content/docs && \
  mkdir -p "$DST/decisoes" "$DST/gerado" && \
  for doc in "$SRC"/0[1-7]-*.md; do \
    [ -f "$doc" ] || continue; \
    base=$(basename "$doc"); \
    title=$(head -1 "$doc" | sed "s/^# //"); \
    printf "%s\n" "---" "title: \"$title\"" "---" "" > "$DST/$base"; \
    tail -n +2 "$doc" >> "$DST/$base"; \
  done && \
  for adr in "$SRC"/decisoes/*.md; do \
    [ -f "$adr" ] || continue; \
    base=$(basename "$adr"); \
    title=$(head -1 "$adr" | sed "s/^# //"); \
    printf "%s\n" "---" "title: \"$title\"" "---" "" > "$DST/decisoes/$base"; \
    tail -n +2 "$adr" >> "$DST/decisoes/$base"; \
  done && \
  cp "$SRC/gerado/ENTIDADES.md" "$DST/gerado/entidades.md" && \
  cp "$SRC/gerado/env-vars.md" "$DST/gerado/env-vars.md" && \
  printf "%s\n" "---" "title: \"Quadro (Kanban)\"" "description: \"Quadro do sprint atual, gerado do backlog YAML.\"" "---" "" ":::caution[Arquivo gerado]" "Não edite manualmente. Rode \`make docs\` para regenerar." ":::" "" > "$DST/gerado/board.md" && \
  tail -n +2 "$SRC/gerado/BOARD.md" >> "$DST/gerado/board.md" && \
  printf "%s\n" "---" "title: \"Roadmap\"" "description: \"Roadmap Now/Next/Later, gerado do backlog YAML.\"" "---" "" ":::caution[Arquivo gerado]" "Não edite manualmente. Rode \`make docs\` para regenerar." ":::" "" > "$DST/gerado/roadmap.md" && \
  tail -n +2 "$SRC/gerado/ROADMAP.md" >> "$DST/gerado/roadmap.md"'
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
