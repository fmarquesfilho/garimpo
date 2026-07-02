#!/usr/bin/env bash
# =============================================================================
# check-data-ownership.sh — Valida fronteiras de dados entre serviços
# =============================================================================
# Regra: cada serviço acessa APENAS seu store designado.
#
#   ┌────────────────┬───────────────────────────────┐
#   │ Componente     │ Store permitido               │
#   ├────────────────┼───────────────────────────────┤
#   │ C# API (src/)  │ PostgreSQL (EF Core)          │
#   │ Go services    │ BigQuery (escrita)            │
#   │ Python analyzer│ BigQuery (leitura)            │
#   └────────────────┴───────────────────────────────┘
#
# Violações detectadas:
#   - Go/Python importando drivers PostgreSQL
#   - C# importando SDK BigQuery
#   - Python escrevendo no PostgreSQL
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "═══════════════════════════════════════════════════════════════"
echo "  Data Ownership Check"
echo "═══════════════════════════════════════════════════════════════"

ERRORS=0

# ── 1. Go services must NOT access PostgreSQL ─────────────────────────────────
echo ""
echo "🔍 Go services → sem acesso direto a PostgreSQL..."

PG_IN_GO=$(grep -rn "database/sql\|pgx\|lib/pq\|gorm\|postgres" \
  --include="*.go" \
  services/ internal/ 2>/dev/null | \
  grep -v "_test.go\|node_modules\|gen/\|vendor/" || true)

if [ -n "$PG_IN_GO" ]; then
  echo -e "${RED}   ✗ Go services importam PostgreSQL:${NC}"
  echo "$PG_IN_GO" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Go services não acessam PostgreSQL${NC}"
fi

# ── 2. C# API must NOT access BigQuery directly ──────────────────────────────
echo ""
echo "🔍 C# API → sem acesso direto a BigQuery..."

BQ_IN_CSHARP=$(grep -rn "Google.Cloud.BigQuery\|BigQueryClient\|BigqueryClient" \
  --include="*.cs" --include="*.csproj" \
  src/ 2>/dev/null | \
  grep -v "bin/\|obj/\|Generated/" || true)

if [ -n "$BQ_IN_CSHARP" ]; then
  echo -e "${RED}   ✗ C# API importa BigQuery SDK:${NC}"
  echo "$BQ_IN_CSHARP" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ C# API não acessa BigQuery diretamente${NC}"
fi

# ── 3. Python analyzer must NOT write to PostgreSQL ───────────────────────────
echo ""
echo "🔍 Python analyzer → sem acesso a PostgreSQL..."

PG_IN_PYTHON=$(grep -rn "psycopg\|asyncpg\|sqlalchemy\|pg8000\|postgres" \
  --include="*.py" --include="*.txt" \
  services/analyzer/ 2>/dev/null || true)

if [ -n "$PG_IN_PYTHON" ]; then
  echo -e "${RED}   ✗ Python analyzer importa PostgreSQL:${NC}"
  echo "$PG_IN_PYTHON" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Python analyzer não acessa PostgreSQL${NC}"
fi

# ── 4. Go services must NOT import C# domain concepts ────────────────────────
echo ""
echo "🔍 Go services → sem importação de conceitos C# (EF Core, DbContext)..."

CSHARP_IN_GO=$(grep -rn "EntityFramework\|DbContext\|EFCore" \
  --include="*.go" \
  services/ internal/ 2>/dev/null | \
  grep -v "_test.go\|gen/" || true)

if [ -n "$CSHARP_IN_GO" ]; then
  echo -e "${RED}   ✗ Go services referenciam conceitos C#:${NC}"
  echo "$CSHARP_IN_GO" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Go services independentes de C#${NC}"
fi

# ── 5. Data flow direction: Collectors write BQ, never read from PG ───────────
echo ""
echo "🔍 Collectors → escrita BQ apenas (sem leitura de PG)..."

PG_CONN_IN_COLLECTORS=$(grep -rn "PostgreSQL\|ConnectionString\|5432\|POSTGRES" \
  --include="*.go" \
  services/collector/ 2>/dev/null | \
  grep -v "_test.go" || true)

if [ -n "$PG_CONN_IN_COLLECTORS" ]; then
  echo -e "${RED}   ✗ Collectors referenciam PostgreSQL:${NC}"
  echo "$PG_CONN_IN_COLLECTORS" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Collectors não acessam PostgreSQL${NC}"
fi

# ── Resultado ─────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════"
if [ "$ERRORS" -eq 0 ]; then
  echo -e "${GREEN}✅ Data ownership respeitado! Fronteiras intactas.${NC}"
  exit 0
else
  echo -e "${RED}❌ $ERRORS violação(ões) de data ownership detectada(s)!${NC}"
  exit 1
fi
