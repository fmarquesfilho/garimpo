#!/usr/bin/env bash
# check-file-size.sh — Impede que arquivos de código cresçam além do limite.
#
# Uso: ./scripts/check-file-size.sh [--max-lines N]
#
# Default: 400 linhas para código de produção.
# Exceções (arquivos _test.go) permitem até 900 linhas.
#
# Rode localmente ou no CI para bloquear merges de arquivos inchados.

set -euo pipefail

MAX_LINES="${1:-400}"
MAX_LINES_TEST=900
EXIT_CODE=0

echo "🔍 Verificando arquivos de código (limite: ${MAX_LINES} linhas, testes: ${MAX_LINES_TEST} linhas)..."
echo ""

# ── Backend Go ─────────────────────────────────────────────────────────────
while IFS= read -r file; do
    lines=$(wc -l < "$file")
    if [[ "$file" == *_test.go ]]; then
        if (( lines > MAX_LINES_TEST )); then
            echo "⚠️  TESTE LONGO: ${file} (${lines} linhas, limite: ${MAX_LINES_TEST})"
            # Testes longos são warning, não bloqueiam
        fi
    else
        if (( lines > MAX_LINES )); then
            echo "❌ ${file} (${lines} linhas, limite: ${MAX_LINES})"
            EXIT_CODE=1
        fi
    fi
done < <(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./gen/*")

# ── Frontend Svelte/JS/TS ──────────────────────────────────────────────────
while IFS= read -r file; do
    lines=$(wc -l < "$file")
    if [[ "$file" == *test* || "$file" == *spec* ]]; then
        if (( lines > MAX_LINES_TEST )); then
            echo "⚠️  TESTE LONGO: ${file} (${lines} linhas, limite: ${MAX_LINES_TEST})"
        fi
    else
        if (( lines > MAX_LINES )); then
            echo "❌ ${file} (${lines} linhas, limite: ${MAX_LINES})"
            EXIT_CODE=1
        fi
    fi
done < <(find ./web/src -type f \( -name "*.svelte" -o -name "*.js" -o -name "*.ts" \) \
    -not -path "*/node_modules/*" -not -path "*/.svelte-kit/*" -not -path "*/build/*" 2>/dev/null || true)

echo ""
if (( EXIT_CODE == 0 )); then
    echo "✅ Todos os arquivos dentro do limite."
else
    echo ""
    echo "💡 Refatore os arquivos acima antes de mergear."
    echo "   Regra: max ${MAX_LINES} linhas por arquivo de produção."
    echo "   Exceção: _test.go / *.spec.* podem ter até ${MAX_LINES_TEST} linhas."
    echo ""
    echo "   Dicas:"
    echo "   - Extraia tipos/helpers para um novo arquivo no mesmo package"
    echo "   - Extraia componentes Svelte menores"
    echo "   - Separe HTTP handlers de lógica de negócio"
fi

exit $EXIT_CODE
