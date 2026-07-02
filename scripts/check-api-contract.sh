#!/usr/bin/env bash
# =============================================================================
# check-api-contract.sh — Verifica drift entre frontend (api.js) e backend (C#)
# =============================================================================
# Extrai todas as chamadas de API do frontend e verifica se o backend registra
# essas rotas. Roda em CI ou local, SEM precisar do servidor rodando.
#
# Uso: ./scripts/check-api-contract.sh
# Exit code 0 = sem drift, 1 = rotas faltantes
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

FRONTEND_API="web/src/lib/api.js"
BACKEND_DIR="src/Garimpei.Api"

echo "═══════════════════════════════════════════════════════════════"
echo "  API Contract Check — Frontend ↔ Backend"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# --- 1. Extrair rotas do frontend (api.js) ---
echo "📋 Extraindo rotas do frontend ($FRONTEND_API)..."

# Extrai todos os paths de fetch/pegar/postar chamados no api.js
FRONTEND_ROUTES=$(grep -oE "'/api/[^'?]*" "$FRONTEND_API" | \
  sed "s/^'//; s/\${[^}]*}/:param/g" | \
  sort -u)

echo "$FRONTEND_ROUTES" | while read -r route; do
  echo "   $route"
done
FRONTEND_COUNT=$(echo "$FRONTEND_ROUTES" | wc -l | tr -d ' ')
echo "   → Total: $FRONTEND_COUNT rotas"
echo ""

# --- 2. Extrair rotas do backend (C# endpoint files) ---
echo "📋 Extraindo rotas do backend ($BACKEND_DIR/Endpoints/*.cs + Program.cs)..."

# Extrai MapGet/MapPost/MapDelete patterns dos arquivos C#
BACKEND_ROUTES=$(grep -rohE '(Map(Get|Post|Delete|Put|Patch))\s*\(\s*"[^"]*"' \
  "$BACKEND_DIR/Endpoints/"*.cs "$BACKEND_DIR/Program.cs" 2>/dev/null | \
  grep -oE '"[^"]*"' | tr -d '"' | \
  sed 's|{[^}]*}|:param|g' | \
  sort -u)

# Também pegar MapGroup prefixes para montar rotas completas
# (rotas como "/api/destinos" + MapGet("/") = /api/destinos)
GROUP_PREFIXES=$(grep -ohE 'MapGroup\s*\(\s*"[^"]*"' \
  "$BACKEND_DIR/Endpoints/"*.cs "$BACKEND_DIR/Program.cs" 2>/dev/null | \
  grep -oE '"[^"]*"' | tr -d '"' | sort -u)

# Combinar prefixes com sub-rotas para ter a lista completa
BACKEND_FULL=$(echo "$BACKEND_ROUTES"$'\n'"$GROUP_PREFIXES" | sort -u)

echo "$BACKEND_FULL" | head -30 | while read -r route; do
  [ -n "$route" ] && echo "   $route"
done
BACKEND_COUNT=$(echo "$BACKEND_FULL" | grep -c '/api/' || echo "0")
echo "   → Total com /api/: $BACKEND_COUNT rotas"
echo ""

# --- 3. Comparar: quais rotas do frontend NÃO existem no backend? ---
echo "🔍 Verificando rotas do frontend presentes no backend..."
echo ""

MISSING=0
FOUND=0

echo "$FRONTEND_ROUTES" | while read -r route; do
  # Normaliza a rota (remove trailing slash, :param → wildcard)
  normalized=$(echo "$route" | sed 's|/$||')
  
  # Procura a rota (ou parte dela) nos arquivos do backend
  # Usa grep direto nos fontes C# — mais confiável que parsing
  base_path=$(echo "$normalized" | sed 's|/api/||; s|/:param||g; s|/.*||')
  
  if grep -rq "$normalized\|$base_path" "$BACKEND_DIR/Endpoints/"*.cs "$BACKEND_DIR/Program.cs" 2>/dev/null; then
    echo -e "   ${GREEN}✓${NC} $route"
  else
    echo -e "   ${RED}✗ FALTANDO${NC} $route"
    MISSING=$((MISSING + 1))
  fi
  FOUND=$((FOUND + 1))
done

echo ""
echo "═══════════════════════════════════════════════════════════════"

# Re-count missing (subshell acima não propaga variáveis)
MISSING_COUNT=$(echo "$FRONTEND_ROUTES" | while read -r route; do
  base_path=$(echo "$route" | sed 's|/api/||; s|/:param||g; s|/.*||')
  if ! grep -rq "$route\|$base_path" "$BACKEND_DIR/Endpoints/"*.cs "$BACKEND_DIR/Program.cs" 2>/dev/null; then
    echo "MISS"
  fi
done | wc -l | tr -d ' ')

if [ "$MISSING_COUNT" -eq "0" ]; then
  echo -e "${GREEN}✅ Sem drift! Todas as $FRONTEND_COUNT rotas do frontend têm backend correspondente.${NC}"
  exit 0
else
  echo -e "${RED}❌ $MISSING_COUNT rota(s) do frontend sem backend correspondente!${NC}"
  echo "   Corrija adicionando os endpoints faltantes no C# ou removendo do frontend."
  exit 1
fi
