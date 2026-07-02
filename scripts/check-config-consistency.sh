#!/usr/bin/env bash
# =============================================================================
# check-config-consistency.sh — Verifica consistência de configurações
# =============================================================================
# Detecta inconsistências entre configs que devem estar sincronizadas:
#   - Nome do dataset BigQuery (deve ser "garimpo" em todos os lugares)
#   - Nome do projeto GCP (deve ser "garimpo-500114" onde hardcoded)
#   - Portas de serviços (collector=50051, collector-amazon=50055, publisher=50052, analyzer=8060)
#   - Base URLs do analyzer
#
# Roda em CI e local. Exit 1 = inconsistência detectada.
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

echo "═══════════════════════════════════════════════════════════════"
echo "  Config Consistency Check"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# ── 1. Dataset BigQuery ───────────────────────────────────────────────────────
# O dataset real é "garimpo". Nenhum arquivo de config/código deve usar "garimpei"
# como dataset (era o nome antigo errado).

echo "🔍 Verificando nome do dataset BigQuery..."

# Procura "garimpei" como dataset (não como nome de serviço/imagem que é válido)
# Ignora: nomes de serviço (garimpei-api, garimpei-v2), domínio (garimpei.app),
# o próprio script, e diretórios gerados
WRONG_DATASET=$(grep -rn "bq_dataset.*garimpei\|BQ_DATASET.*garimpei\|--dataset=garimpei" \
  --include="*.py" --include="*.yaml" --include="*.yml" \
  --include="*.json" --include="*.toml" --include="*.sql" \
  . 2>/dev/null | \
  grep -v "node_modules\|\.git/\|bin/\|obj/\|check-config-consistency" || true)

if [ -n "$WRONG_DATASET" ]; then
  echo -e "${RED}   ✗ Dataset incorreto 'garimpei' encontrado:${NC}"
  echo "$WRONG_DATASET" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Nenhuma referência ao dataset errado 'garimpei'${NC}"
fi

# ── 2. Portas dos serviços gRPC ───────────────────────────────────────────────
# Padrão: collector=50051, collector-amazon=50055, publisher=50052, alerter=50053, scheduler=50054

echo ""
echo "🔍 Verificando consistência de portas..."

check_port() {
  local service=$1
  local expected_port=$2
  local exclude_pattern=${3:-"XYZNOEXCLUDE"}

  # Only check actual port config (address definitions, port mappings)
  # Not source code, docs, or generated files that mention service names
  WRONG_PORTS=$(grep -rn "${service}.*Address.*localhost:[0-9]\|${service}.*ADDR.*:[0-9]\|\"${expected_port}:" \
    --include="*.cs" --include="*.yaml" --include="*.yml" \
    --include="*.json" --include="*.go" \
    . 2>/dev/null | \
    grep -v "node_modules\|\.git/\|bin/\|obj/\|gen/\|Protos/Generated\|docs-site\|backlog\|api/openapi" | \
    grep -v "$exclude_pattern" | \
    grep "localhost\|ADDR\|Address" | \
    grep -v "$expected_port" || true)

  if [ -n "$WRONG_PORTS" ]; then
    echo -e "${RED}   ✗ Porta inconsistente para $service (esperado: $expected_port):${NC}"
    echo "$WRONG_PORTS" | while read -r line; do echo "     $line"; done
    ERRORS=$((ERRORS + 1))
  fi
}

check_port "Collector" "50051" "Amazon"
check_port "CollectorAmazon" "50055"
check_port "Publisher" "50052"
echo -e "${GREEN}   ✓ Portas de serviços gRPC consistentes${NC}"

# ── 3. URL do Analyzer ────────────────────────────────────────────────────────
# Padrão: porta 8060

echo ""
echo "🔍 Verificando URL do Analyzer..."

WRONG_ANALYZER=$(grep -rn "Analyzer.*BaseUrl\|analyzer.*url\|ANALYZER.*URL" \
  --include="*.cs" --include="*.json" --include="*.yaml" --include="*.yml" \
  . 2>/dev/null | \
  grep -v "node_modules\|.git/\|bin/\|obj/" | \
  grep "localhost" | \
  grep -v "8060" || true)

if [ -n "$WRONG_ANALYZER" ]; then
  echo -e "${RED}   ✗ Porta do Analyzer inconsistente (esperado: 8060):${NC}"
  echo "$WRONG_ANALYZER" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ URL do Analyzer consistente (porta 8060)${NC}"
fi

# ── 4. Tabelas BigQuery — schema SQL vs Go EnsureSchema vs Analyzer ───────────
# Todas as referências a tabelas devem usar os mesmos nomes.

echo ""
echo "🔍 Verificando nomes das tabelas BigQuery..."

EXPECTED_TABLES="snapshots eventos buscas conversoes destinos templates publicacoes favoritos"

for table in $EXPECTED_TABLES; do
  # Verifica se a tabela aparece no schema SQL
  if ! grep -q "$table" deploy/bigquery_schema.sql 2>/dev/null && \
     ! grep -q "$table" internal/store/bigquery_schema.go 2>/dev/null; then
    echo -e "${RED}   ✗ Tabela '$table' não encontrada no schema SQL nem no Go EnsureSchema${NC}"
    ERRORS=$((ERRORS + 1))
  fi
done
echo -e "${GREEN}   ✓ Todas as tabelas esperadas estão documentadas${NC}"

# ── 5. Frontend API base URL não deve estar hardcoded para produção ───────────

echo ""
echo "🔍 Verificando URLs hardcoded no frontend..."

HARDCODED_URLS=$(grep -rn "https://garimpei\|https://garimpo" \
  --include="*.js" --include="*.svelte" --include="*.ts" \
  web/src/ 2>/dev/null | \
  grep -v "node_modules\|.git/" || true)

if [ -n "$HARDCODED_URLS" ]; then
  echo -e "${RED}   ✗ URLs de produção hardcoded no frontend (deveria usar env var):${NC}"
  echo "$HARDCODED_URLS" | while read -r line; do echo "     $line"; done
  ERRORS=$((ERRORS + 1))
else
  echo -e "${GREEN}   ✓ Nenhuma URL de produção hardcoded no frontend${NC}"
fi

# ── Resultado ─────────────────────────────────────────────────────────────────

echo ""
echo "═══════════════════════════════════════════════════════════════"
if [ "$ERRORS" -eq 0 ]; then
  echo -e "${GREEN}✅ Sem inconsistências! Todas as configs estão alinhadas.${NC}"
  exit 0
else
  echo -e "${RED}❌ $ERRORS inconsistência(s) detectada(s)!${NC}"
  exit 1
fi
