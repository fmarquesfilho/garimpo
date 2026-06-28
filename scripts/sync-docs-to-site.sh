#!/usr/bin/env bash
# scripts/sync-docs-to-site.sh — Sincroniza docs canônicos para docs-site/src/content/docs
# Os docs/ são a fonte; docs-site/ é gerado para o Starlight.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/docs"
DST="$ROOT/docs-site/src/content/docs"

# Mapear docs canônicos
declare -a DOCS=(
  "01-visao-e-negocio.md"
  "02-arquitetura.md"
  "03-fluxos-e-modelo.md"
  "04-operacao-shopee.md"
  "05-manual-do-usuario.md"
  "06-qualidade-e-testes.md"
  "07-dados-e-ia.md"
)

for doc in "${DOCS[@]}"; do
  if [ -f "$SRC/$doc" ]; then
    # Extrair título (primeira linha com #)
    title=$(head -1 "$SRC/$doc" | sed 's/^# //')
    # Criar versão com frontmatter
    {
      echo "---"
      echo "title: \"$title\""
      echo "---"
      echo ""
      tail -n +2 "$SRC/$doc"
    } > "$DST/$doc"
  fi
done

# Sincronizar ADRs
bash "$ROOT/scripts/copy-adrs.sh"

# Sincronizar gerados (já têm frontmatter)
cp "$ROOT/docs/gerado/ENTIDADES.md" "$DST/gerado/entidades.md" 2>/dev/null || true

# Board e Roadmap: copiar conteúdo para dentro do frontmatter do placeholder
if [ -f "$ROOT/docs/gerado/BOARD.md" ]; then
  {
    echo "---"
    echo "title: \"Quadro (Kanban)\""
    echo "description: \"Quadro do sprint atual, gerado do backlog YAML.\""
    echo "---"
    echo ""
    echo ":::caution[Arquivo gerado]"
    echo "Não edite manualmente. Rode \`make docs-board\` para regenerar."
    echo ":::"
    echo ""
    cat "$ROOT/docs/gerado/BOARD.md"
  } > "$DST/gerado/board.md"
fi

if [ -f "$ROOT/docs/gerado/ROADMAP.md" ]; then
  {
    echo "---"
    echo "title: \"Roadmap\""
    echo "description: \"Roadmap Now/Next/Later, gerado do backlog YAML.\""
    echo "---"
    echo ""
    echo ":::caution[Arquivo gerado]"
    echo "Não edite manualmente. Rode \`make docs-board\` para regenerar."
    echo ":::"
    echo ""
    cat "$ROOT/docs/gerado/ROADMAP.md"
  } > "$DST/gerado/roadmap.md"
fi

echo "Sincronizados $(ls "$DST"/*.md 2>/dev/null | wc -l) docs canônicos para docs-site"
