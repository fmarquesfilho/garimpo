#!/usr/bin/env bash
# =============================================================================
# pre-push-check.sh — Roda TODOS os testes e verificações antes de push
# =============================================================================
# Instalar como git hook:
#   ln -sf ../../scripts/pre-push-check.sh .git/hooks/pre-push
#
# Ou rodar manualmente:
#   ./scripts/pre-push-check.sh
#
# Bloqueia o push se qualquer check falhar.
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo ""
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}  Pre-push: verificação completa antes de enviar${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo ""

ERRORS=0
TOTAL=0

run_check() {
    local name=$1
    shift
    TOTAL=$((TOTAL + 1))
    echo -n "  [$TOTAL] $name... "
    if "$@" > /tmp/pre-push-output-$TOTAL.log 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ FALHOU${NC}"
        echo -e "      Veja: /tmp/pre-push-output-$TOTAL.log"
        ERRORS=$((ERRORS + 1))
    fi
}

# ── C# ────────────────────────────────────────────────────────────────────────
echo "🔨 C# (build + testes + arquitetura):"
run_check "Build" dotnet build --nologo src/Garimpei.sln
run_check "Testes (38: persistence + architecture + integration)" dotnet test --nologo --no-build src/Garimpei.sln

# ── Go ────────────────────────────────────────────────────────────────────────
echo ""
echo "🔨 Go (build + testes):"
run_check "Build" go build ./...
run_check "Testes" go test ./...

# ── Drift checks ─────────────────────────────────────────────────────────────
echo ""
echo "🔍 Drift checks (cross-stack):"
run_check "API contract (frontend↔backend)" ./scripts/check-api-contract.sh
run_check "Config consistency (dataset, portas)" ./scripts/check-config-consistency.sh
run_check "Schema sync (BQ↔Go↔C#↔Analyzer)" ./scripts/check-schema-sync.sh

# ── Frontend (opcional — só se houve mudanças) ────────────────────────────────
if git diff --cached --name-only HEAD 2>/dev/null | grep -q "^web/" || \
   git diff --name-only HEAD~1..HEAD 2>/dev/null | grep -q "^web/"; then
    echo ""
    echo "🔨 Frontend (lint + unit tests):"
    run_check "Lint JS" npm run lint:js --prefix web
    run_check "Vitest" npm run test:unit --prefix web
fi

# ── Resultado ─────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════"
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✅ $TOTAL/$TOTAL checks passaram. Push liberado.${NC}"
    exit 0
else
    echo -e "${RED}❌ $ERRORS/$TOTAL checks falharam. Push bloqueado.${NC}"
    echo -e "${RED}   Corrija os erros acima e tente novamente.${NC}"
    exit 1
fi
