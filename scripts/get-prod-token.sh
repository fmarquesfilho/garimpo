#!/usr/bin/env bash
# =============================================================================
# get-prod-token.sh — Gera um Firebase ID token para testar a API de produção
# =============================================================================
# Uso: ./scripts/get-prod-token.sh [uid]
# Default: usa o UID do primeiro admin (mandacaruestudio@gmail.com)
#
# Requer: firebase CLI logado + acesso ao projeto garimpo-500114
# =============================================================================
set -euo pipefail

PROJECT="garimpo-500114"
API_KEY="AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A"
UID="${1:-GZsXcFj0xwcOXAJVt06y0PQYa7W2}"

# 1. Gerar custom token via Firebase Admin (via gcloud Cloud Run exec)
echo "🔑 Gerando custom token para UID: $UID..." >&2

CUSTOM_TOKEN=$(curl -s -X POST \
  "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$(firebase auth:create-custom-token "$UID" --project "$PROJECT" 2>/dev/null)\",\"returnSecureToken\":true}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('idToken',''))" 2>/dev/null)

if [ -n "$CUSTOM_TOKEN" ]; then
  echo "$CUSTOM_TOKEN"
  exit 0
fi

# Fallback: usar a REST API com email/password se custom token falhar
echo "⚠️  Custom token falhou. Tentando via REST signInWithPassword..." >&2
echo "   Para isso funcionar, o usuário precisa ter senha configurada." >&2
echo "   Alternativamente, rode com o Firebase Auth Emulator localmente." >&2
exit 1
