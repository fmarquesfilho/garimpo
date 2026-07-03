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
# Executa blocos independentes em paralelo para reduzir tempo total.
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

TMPDIR_PP=$(mktemp -d)
trap 'rm -rf "$TMPDIR_PP"' EXIT

# ── Bloco paralelo: cada grupo escreve resultado num arquivo ─────────────────

run_block() {
    local block_name=$1
    local log_file="$TMPDIR_PP/${block_name}.log"
    local status_file="$TMPDIR_PP/${block_name}.status"
    shift
    if "$@" > "$log_file" 2>&1; then
        echo "0" > "$status_file"
    else
        echo "1" > "$status_file"
    fi
}

# ── Bloco C# ──────────────────────────────────────────────────────────────────
block_csharp() {
    local log="$TMPDIR_PP/csharp.log"
    local errors=0
    echo "🔨 C# (build + test)" > "$log"
    if ! dotnet build --nologo -v quiet src/Garimpei.sln >> "$log" 2>&1; then
        echo "  ✗ Build falhou" >> "$log"
        echo "1" > "$TMPDIR_PP/csharp.status"
        return
    fi
    if ! dotnet test --nologo --no-build -v quiet src/Garimpei.sln >> "$log" 2>&1; then
        echo "  ✗ Testes falharam" >> "$log"
        echo "1" > "$TMPDIR_PP/csharp.status"
        return
    fi
    echo "  ✓ Build + testes OK" >> "$log"
    echo "0" > "$TMPDIR_PP/csharp.status"
}

# ── Bloco Go ──────────────────────────────────────────────────────────────────
block_go() {
    local log="$TMPDIR_PP/go.log"
    echo "🔨 Go (build + test + lint)" > "$log"
    if ! go build ./... >> "$log" 2>&1; then
        echo "  ✗ Build falhou" >> "$log"
        echo "1" > "$TMPDIR_PP/go.status"
        return
    fi
    if ! go test ./... >> "$log" 2>&1; then
        echo "  ✗ Testes falharam" >> "$log"
        echo "1" > "$TMPDIR_PP/go.status"
        return
    fi
    if ! golangci-lint run ./... >> "$log" 2>&1; then
        echo "  ✗ Lint falhou" >> "$log"
        echo "1" > "$TMPDIR_PP/go.status"
        return
    fi
    echo "  ✓ Build + testes + lint OK" >> "$log"
    echo "0" > "$TMPDIR_PP/go.status"
}

# ── Bloco Python + Drift checks ──────────────────────────────────────────────
block_checks() {
    local log="$TMPDIR_PP/checks.log"
    echo "🐍 Python + 🔍 Drift checks" > "$log"

    if command -v ruff &> /dev/null; then
        if ! ruff check services/analyzer/ >> "$log" 2>&1; then
            echo "  ✗ Ruff lint falhou" >> "$log"
            echo "1" > "$TMPDIR_PP/checks.status"
            return
        fi
    fi

    local scripts_ok=true
    for script in check-api-contract.sh check-config-consistency.sh check-schema-sync.sh check-data-ownership.sh check-stale-refs.sh check-service-contracts.sh; do
        if [ -f "./scripts/$script" ]; then
            if ! "./scripts/$script" >> "$log" 2>&1; then
                echo "  ✗ $script falhou" >> "$log"
                scripts_ok=false
            fi
        fi
    done

    if [ "$scripts_ok" = false ]; then
        echo "1" > "$TMPDIR_PP/checks.status"
        return
    fi

    echo "  ✓ Python lint + drift checks OK" >> "$log"
    echo "0" > "$TMPDIR_PP/checks.status"
}

# ── Bloco Docs ────────────────────────────────────────────────────────────────
block_docs() {
    local log="$TMPDIR_PP/docs.log"
    echo "📚 Docs" > "$log"

    if ! make docs-check >> "$log" 2>&1; then
        echo "  ✗ docs-check falhou (rode: make docs)" >> "$log"
        echo "1" > "$TMPDIR_PP/docs.status"
        return
    fi

    # Build docs-site apenas se houve mudanças relevantes
    if [ -d "docs-site" ] && [ -f "docs-site/package.json" ]; then
        if git diff --name-only HEAD~1..HEAD 2>/dev/null | grep -qE "^(docs/|docs-site/|backlog/|scripts/sync-docs-to-site\.sh)"; then
            if ! (./scripts/sync-docs-to-site.sh && cd docs-site && npm ci --silent && npm run build) >> "$log" 2>&1; then
                echo "  ✗ docs-site build falhou" >> "$log"
                echo "1" > "$TMPDIR_PP/docs.status"
                return
            fi
        fi
    fi

    echo "  ✓ Docs OK" >> "$log"
    echo "0" > "$TMPDIR_PP/docs.status"
}

# ── Disparar tudo em paralelo ─────────────────────────────────────────────────
block_csharp &
PID_CSHARP=$!

block_go &
PID_GO=$!

block_checks &
PID_CHECKS=$!

block_docs &
PID_DOCS=$!

# Esperar todos terminarem
wait $PID_CSHARP $PID_GO $PID_CHECKS $PID_DOCS

# ── Coletar resultados ────────────────────────────────────────────────────────
ERRORS=0

for block in csharp go checks docs; do
    cat "$TMPDIR_PP/${block}.log"
    echo ""
    status=$(cat "$TMPDIR_PP/${block}.status" 2>/dev/null || echo "1")
    if [ "$status" != "0" ]; then
        ERRORS=$((ERRORS + 1))
    fi
done

# ── Frontend (opcional — só se houve mudanças) ────────────────────────────────
if git diff --name-only HEAD~1..HEAD 2>/dev/null | grep -q "^web/"; then
    echo "🔨 Frontend (lint + unit tests):"
    if ! npm run lint:js --prefix web > /dev/null 2>&1; then
        echo -e "  ${RED}✗ Lint JS falhou${NC}"
        ERRORS=$((ERRORS + 1))
    elif ! npm run test:unit --prefix web > /dev/null 2>&1; then
        echo -e "  ${RED}✗ Vitest falhou${NC}"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "  ${GREEN}✓ Frontend OK${NC}"
    fi
    echo ""
fi

# ── Resultado ─────────────────────────────────────────────────────────────────
echo "═══════════════════════════════════════════════════════════════"
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✅ Todos os checks passaram. Push liberado.${NC}"
    exit 0
else
    echo -e "${RED}❌ $ERRORS bloco(s) falharam. Push bloqueado.${NC}"
    echo -e "${RED}   Veja os logs acima para detalhes.${NC}"
    exit 1
fi
