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

# ── Backlog (tarefas YAML → Markdown) ──────────────────────────────────────

mkdir -p "$DST/backlog"
for task in "$ROOT"/backlog/tasks/T-*.yaml; do
  [ -f "$task" ] || continue
  base=$(basename "$task" .yaml)

  # Extrair campos do YAML (parse simplificado — campos de linha única)
  id=$(grep '^id:' "$task" | sed 's/^id: *//')
  titulo=$(grep '^titulo:' "$task" | sed 's/^titulo: *//; s/^"//; s/"$//')
  epic=$(grep '^epic:' "$task" | sed 's/^epic: *//' || true)
  status=$(grep '^status:' "$task" | sed 's/^status: *//')
  prioridade=$(grep '^prioridade:' "$task" | sed 's/^prioridade: *//')
  estimativa=$(grep '^estimativa:' "$task" | sed 's/^estimativa: *//' || true)
  sprint=$(grep '^sprint:' "$task" | sed 's/^sprint: *//; s/^"//; s/"$//' || true)
  criada=$(grep '^criada_em:' "$task" | sed 's/^criada_em: *//; s/^"//; s/"$//' || true)
  atualizada=$(grep '^atualizada_em:' "$task" | sed 's/^atualizada_em: *//; s/^"//; s/"$//' || true)

  # Extrair valor (campo multi-linha com >)
  valor=$(awk '/^valor:/{found=1; sub(/^valor: *>? */, ""); if ($0) print; next} found && /^  /{sub(/^  /,""); print; next} found{exit}' "$task" | tr '\n' ' ' | sed 's/  */ /g; s/ *$//')

  # Extrair critérios (lista YAML)
  criterios=$(awk '/^criterios:/{found=1; next} found && /^  - /{sub(/^  - /,""); print; next} found{exit}' "$task")

  # Extrair depende_de (inline [X, Y] ou multi-linha)
  deps=$(grep '^depende_de:' "$task" | sed 's/^depende_de: *//; s/\[//; s/\]//; s/,/ /g; s/^ *//; s/ *$//' || true)
  if [ "$deps" = "" ] || [ "$deps" = "[]" ]; then
    deps=""
  fi

  # Extrair tags
  tags=$(grep '^tags:' "$task" | sed 's/^tags: *\[//; s/\]//; s/, */ /g' || true)

  # Status badge
  case "$status" in
    done)    badge="✅ Concluída" ;;
    doing)   badge="🔨 Em andamento" ;;
    next)    badge="⏭️ Próximo" ;;
    blocked) badge="🚫 Bloqueada" ;;
    review)  badge="👀 Revisão" ;;
    *)       badge="📋 Backlog" ;;
  esac

  # Gerar Markdown
  {
    echo "---"
    echo "title: \"$id — $titulo\""
    echo "---"
    echo ""
    echo "| Campo | Valor |"
    echo "|-------|-------|"
    echo "| Status | $badge |"
    echo "| Épico | $epic |"
    echo "| Prioridade | $prioridade |"
    [ -n "$estimativa" ] && echo "| Estimativa | $estimativa |"
    [ -n "$sprint" ] && echo "| Sprint | $sprint |"
    [ -n "$criada" ] && echo "| Criada | $criada |"
    [ -n "$atualizada" ] && echo "| Atualizada | $atualizada |"
    echo ""
    if [ -n "$valor" ]; then
      echo "## Valor"
      echo ""
      echo "$valor"
      echo ""
    fi
    if [ -n "$criterios" ]; then
      echo "## Critérios de aceite"
      echo ""
      echo "$criterios" | while IFS= read -r c; do
        echo "- $c"
      done
      echo ""
    fi
    if [ -n "$deps" ]; then
      echo "## Dependências"
      echo ""
      for d in $deps; do
        # Procurar o ficheiro da tarefa para gerar o link correcto
        task_file=$(find "$ROOT/backlog/tasks" -name "${d}-*" -print -quit 2>/dev/null)
        if [ -n "$task_file" ]; then
          link_base=$(basename "$task_file" .yaml)
          echo "- [$d](/docs/backlog/$link_base/)"
        else
          echo "- $d"
        fi
      done
      echo ""
    fi
    if [ -n "$tags" ]; then
      echo "## Tags"
      echo ""
      printf '`%s` ' $tags
      echo ""
    fi
  } > "$DST/backlog/$base.md"
done

# ── Resumo ──────────────────────────────────────────────────────────────────

n_docs=$(find "$DST" -maxdepth 1 -name "0*.md" | wc -l | tr -d ' ')
n_adrs=$(find "$DST/decisoes" -name "*.md" | wc -l | tr -d ' ')
n_gen=$(find "$DST/gerado" -name "*.md" | wc -l | tr -d ' ')
n_tasks=$(find "$DST/backlog" -name "*.md" | wc -l | tr -d ' ')
echo "✓ Sincronizados: $n_docs docs, $n_adrs ADRs, $n_gen gerados, $n_tasks tarefas → docs-site"
