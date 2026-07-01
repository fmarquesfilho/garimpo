#!/usr/bin/env bash
# check-api-spec-sync.sh — Valida que todos os endpoints documentados no openapi.yaml
# existem no C# (não retornam 404). Evita drift entre spec e implementação.
#
# Uso: ./scripts/check-api-spec-sync.sh [base_url]
# Default base_url: http://localhost:8090 (C# dev local)
#
# Requer: o C# API rodando (ou no CI, usando o service container).
# Endpoints autenticados retornam 401 (ok — existe). 404 = drift.

set -euo pipefail

BASE_URL="${1:-http://localhost:8090}"
SPEC="api/openapi.yaml"
EXIT_CODE=0

echo "🔍 Verificando endpoints do openapi.yaml contra ${BASE_URL}..."
echo ""

# Extrair paths do YAML (linhas que começam com /api ou /health)
paths=$(grep -E "^  /api/|^  /health" "$SPEC" | sed 's/://; s/^  //' | sort -u)

for path in $paths; do
    # Substituir path params por valores dummy
    url="${BASE_URL}${path}"
    url=$(echo "$url" | sed 's/{[^}]*}/1/g')

    # Fazer GET (sem auth — esperamos 200, 401, 400, 405... qualquer coisa exceto 404)
    status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")

    if [ "$status" = "404" ]; then
        echo "❌ ${path} → 404 (não implementado)"
        EXIT_CODE=1
    elif [ "$status" = "000" ]; then
        echo "⚠️  ${path} → timeout/erro de conexão"
        EXIT_CODE=1
    else
        echo "✅ ${path} → ${status}"
    fi
done

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Todos os endpoints documentados existem na implementação."
else
    echo ""
    echo "❌ Endpoints com drift detectado!"
    echo "   Ou implemente o endpoint no C#, ou remova do api/openapi.yaml."
fi

exit $EXIT_CODE
