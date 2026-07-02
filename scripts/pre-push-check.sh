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
echo "🔨 Go (build + testes + lint):"
run_check "Build" go build ./...
run_check "Testes" go test ./...
run_check "Lint (golangci-lint)" golangci-lint run ./...

# ── Python ────────────────────────────────────────────────────────────────────
echo ""
echo "🐍 Python (lint):"
if command -v ruff &> /dev/null; then
    run_check "Lint (ruff)" ruff check services/analyzer/
else
    echo "  [skip] ruff não instalado — instale com: pip install ruff"
fi

# ── Drift checks ─────────────────────────────────────────────────────────────
echo ""
echo "🔍 Drift checks (cross-stack):"
run_check "API contract (frontend↔backend)" ./scripts/check-api-contract.sh
run_check "Config consistency (dataset, portas)" ./scripts/check-config-consistency.sh
run_check "Schema sync (BQ↔Go↔C#↔Analyzer)" ./scripts/check-schema-sync.sh
run_check "Data ownership (PG↔BQ boundaries)" ./scripts/check-data-ownership.sh
run_check "Stale refs (dead code, serviços removidos)" ./scripts/check-stale-refs.sh

# ── Docs (generated files + dead links) ───────────────────────────────────────
echo ""
echo "📚 Docs (geração + build site):"
run_check "Docs check (generated up to date)" make docs-check
if [ -d "docs-site" ] && [ -f "docs-site/package.json" ]; then
    # Só roda sync + build se houve mudanças em docs, docs-site, backlog ou sync script
    if git diff --cached --name-only HEAD 2>/dev/null | grep -qE "^(docs/|docs-site/|backlog/|scripts/sync-docs-to-site\.sh)" || \
       git diff --name-only HEAD~1..HEAD 2>/dev/null | grep -qE "^(docs/|docs-site/|backlog/|scripts/sync-docs-to-site\.sh)"; then
        run_check "Docs sync + site build (dead links)" bash -c "./scripts/sync-docs-to-site.sh && cd docs-site && npm ci --silent && npm run build"
    else
        echo "  [skip] docs-site build — sem mudanças em docs/, docs-site/, backlog/"
    fi
fi

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
