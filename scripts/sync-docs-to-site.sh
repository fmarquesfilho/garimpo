#!/usr/bin/env bash
# scripts/sync-docs-to-site.sh — Gera docs-site/src/content/docs a partir de docs/
#
# Fonte única: docs/ (Markdown puro, legível no GitHub)
# Destino: docs-site/src/content/docs/ (com frontmatter para Starlight)
#
# Ficheiros originais do site (index.mdx, gerado/api.mdx) NÃO são tocados.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/docs"
DST="$ROOT/docs-site/src/content/docs"

# ── Helpers ─────────────────────────────────────────────────────────────────

# Converte um .md com "# Título" na linha 1 em versão com frontmatter YAML.
add_frontmatter() {
  local src_file="$1" dst_file="$2"
  local title
  title=$(head -1 "$src_file" | sed 's/^# //')
  {
    echo "---"
    echo "title: \"$title\""
    echo "---"
    echo ""
    # Pula a primeira linha (título já está no frontmatter)
    tail -n +2 "$src_file"
  } > "$dst_file"
}

# Copia ficheiro que já tem frontmatter YAML (---\n...\n---)
copy_with_frontmatter() {
  local src_file="$1" dst_file="$2"
  cp "$src_file" "$dst_file"
}

# Wraps conteúdo sem frontmatter com título e aviso de gerado.
wrap_generated() {
  local src_file="$1" dst_file="$2" title="$3" desc="$4"
  {
    echo "---"
    echo "title: \"$title\""
    echo "description: \"$desc\""
    echo "---"
    echo ""
    echo ":::caution[Arquivo gerado]"
    echo "Não edite manualmente. Rode \`make docs\` para regenerar."
    echo ":::"
    echo ""
    # Pula a primeira linha se for um heading (evita duplicação de título)
    if head -1 "$src_file" | grep -q '^# '; then
      tail -n +2 "$src_file"
    else
      cat "$src_file"
    fi
  } > "$dst_file"
}

# ── Docs canónicos (01-07) ──────────────────────────────────────────────────

for doc in "$SRC"/0[1-7]-*.md; do
  [ -f "$doc" ] || continue
  base=$(basename "$doc")
  add_frontmatter "$doc" "$DST/$base"
done

# ── ADRs (decisoes/) ───────────────────────────────────────────────────────

mkdir -p "$DST/decisoes"
for adr in "$SRC"/decisoes/*.md; do
  [ -f "$adr" ] || continue
  base=$(basename "$adr")
  add_frontmatter "$adr" "$DST/decisoes/$base"
done

# ── Gerados ─────────────────────────────────────────────────────────────────

mkdir -p "$DST/gerado"

# ENTIDADES e env-vars já têm frontmatter no source
copy_with_frontmatter "$SRC/gerado/ENTIDADES.md" "$DST/gerado/entidades.md"
copy_with_frontmatter "$SRC/gerado/env-vars.md" "$DST/gerado/env-vars.md"

# BOARD e ROADMAP não têm frontmatter — wrappear
wrap_generated "$SRC/gerado/BOARD.md" "$DST/gerado/board.md" \
  "Quadro (Kanban)" "Quadro do sprint atual, gerado do backlog YAML."

wrap_generated "$SRC/gerado/ROADMAP.md" "$DST/gerado/roadmap.md" \
  "Roadmap" "Roadmap Now/Next/Later, gerado do backlog YAML."

# ── Resumo ──────────────────────────────────────────────────────────────────

n_docs=$(find "$DST" -maxdepth 1 -name "0*.md" | wc -l | tr -d ' ')
n_adrs=$(find "$DST/decisoes" -name "*.md" | wc -l | tr -d ' ')
n_gen=$(find "$DST/gerado" -name "*.md" | wc -l | tr -d ' ')
echo "✓ Sincronizados: $n_docs docs, $n_adrs ADRs, $n_gen gerados → docs-site"
