#!/usr/bin/env bash
# ============================================================
# audit-ui-coverage.sh — Relatório de cobertura da biblioteca UI
#
# Mostra quais elementos/padrões ainda usam HTML nativo ou
# reimplementações ad-hoc ao invés dos componentes padronizados.
#
# Uso: ./scripts/audit-ui-coverage.sh
# ============================================================

set -uo pipefail
cd "$(dirname "$0")/../src"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}  AUDITORIA DE COBERTURA — Biblioteca de Componentes UI${NC}"
echo -e "${BOLD}═══════════════════════════════════════════════════════${NC}"
echo ""

# ── Adoção dos componentes (quem já usa) ────────────────────
echo -e "${GREEN}✓ COMPONENTES UI ADOTADOS (imports contados)${NC}"
echo ""
grep -rh "from.*components/ui" --include="*.svelte" | \
  grep -v node_modules | \
  sed 's/.*import {//' | sed 's/}.*//' | tr ',' '\n' | \
  sed 's/^ *//' | sed 's/ *$//' | grep -v '^$' | \
  sort | uniq -c | sort -rn | \
  while read count name; do
    printf "  %3s× %s\n" "$count" "$name"
  done
echo ""

# ── Compostos Bits UI: disponíveis mas não usados ───────────
echo -e "${RED}✗ COMPOSTOS BITS UI DISPONÍVEIS MAS NÃO USADOS${NC}"
echo ""
for comp in Dialog DropdownMenu Select Tabs Tooltip; do
  count=$(grep -r "import.*{[^}]*$comp" --include="*.svelte" | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
  if [ "$count" -eq 0 ]; then
    printf "  ⚠  %-14s → 0 consumidores (componente existe mas ninguém importa)\n" "$comp"
  else
    printf "  ✓  %-14s → %s consumidores\n" "$comp" "$count"
  fi
done
echo ""

# ── Elementos nativos com equivalente UI ────────────────────
echo -e "${YELLOW}⚡ ELEMENTOS NATIVOS COM EQUIVALENTE UI${NC}"
echo ""

btn_count=$(grep -r '<button' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/Button' | grep -v 'components/ui/Tabs' | grep -v 'components/ui/DropdownMenu' | grep -v 'components/ui/Dialog' | wc -l | tr -d ' ')
input_count=$(grep -r '<input' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/Input' | wc -l | tr -d ' ')
select_count=$(grep -r '<select' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/Select' | wc -l | tr -d ' ')

printf "  <button> inline:  %s  (equivalente: <Button>)\n" "$btn_count"
printf "  <input> inline:   %s  (equivalente: <Input>)\n" "$input_count"
printf "  <select> nativo:  %s   (equivalente: <Select>)\n" "$select_count"
echo ""

# ── Padrões reimplementados ─────────────────────────────────
echo -e "${YELLOW}⚡ PADRÕES DE INTERAÇÃO REIMPLEMENTADOS${NC}"
echo ""

dialog_files=$(grep -rln "backdrop\|modal\|overlay" --include="*.svelte" | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
tabs_files=$(grep -rln "tab.*ativa\|tab.*active\|aba.*ativa\|TabBar" --include="*.svelte" | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
tooltip_files=$(grep -rln 'title="[^"]*"' --include="*.svelte" | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
dropdown_files=$(grep -rln "dropdown\|menu.*aberto\|menu.*open" --include="*.svelte" | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')

printf "  Modais/overlays ad-hoc:   %s arquivos  (usar <Dialog>)\n" "$dialog_files"
printf "  Tabs/abas com TabBar:     %s arquivos  (usar <Tabs>)\n" "$tabs_files"
printf "  title= sem Tooltip:       %s arquivos  (considerar <Tooltip>)\n" "$tooltip_files"
printf "  Dropdowns custom:         %s arquivos  (usar <DropdownMenu>)\n" "$dropdown_files"
echo ""

# ── Utility classes legadas ─────────────────────────────────
echo -e "${YELLOW}⚡ UTILITY CLASSES LEGADAS (devem migrar para componentes)${NC}"
echo ""

badge_count=$(grep -r 'class.*badge' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
msg_count=$(grep -r 'class.*msg-' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')
btn_class=$(grep -r 'class.*btn' --include='*.svelte' | grep -v node_modules | grep -v 'components/ui/' | wc -l | tr -d ' ')

printf "  .badge class:       %s  (usar <Badge>)\n" "$badge_count"
printf "  .msg-erro/sucesso:  %s   (usar <Alert>)\n" "$msg_count"
printf "  .btn class:         %s  (usar <Button>)\n" "$btn_class"
echo ""

# ── Hex colors ──────────────────────────────────────────────
hex_count=$(grep -rn '#[0-9a-fA-F]\{6\}' --include='*.svelte' | grep -v node_modules | grep -v 'tokens.css' | grep -v 'app.css' | wc -l | tr -d ' ')
if [ "$hex_count" -eq 0 ]; then
  echo -e "${GREEN}✓ ZERO hex colors hardcoded nos componentes${NC}"
else
  echo -e "${RED}✗ $hex_count hex colors hardcoded (devem usar tokens)${NC}"
fi
echo ""

# ── Score resumo ────────────────────────────────────────────
total_native=$((btn_count + input_count + select_count))
echo -e "${BOLD}── RESUMO ──${NC}"
echo ""
echo "  Elementos nativos com equivalente:  $total_native"
echo "  Compostos Bits UI sem consumidores: Dialog, DropdownMenu, Select, Tabs, Tooltip"
echo "  Utility classes legadas:            $((badge_count + msg_count + btn_class))"
echo "  Hex colors hardcoded:               $hex_count"
echo ""
echo -e "  ${BOLD}Próximos alvos de migração:${NC}"
echo "  1. Substituir TabBar por Tabs nas routes (lojas, publicacoes)"
echo "  2. Migrar <select> nativos para <Select> Bits UI (7 arquivos)"
echo "  3. Usar <Dialog> para NavDrawer e confirmações"
echo "  4. Converter title= para <Tooltip> em interações importantes"
echo ""
