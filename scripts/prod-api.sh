#!/usr/bin/env bash
# =============================================================================
# prod-api.sh — Chama a API de produção com autenticação Firebase
# =============================================================================
# Uso:
#   ./scripts/prod-api.sh GET /api/destinos
#   ./scripts/prod-api.sh POST /api/publicacoes '{"nome":"Test","preco":10,"destino_id":"xxx"}'
#   ./scripts/prod-api.sh GET /api/publicacoes | jq
#
# Setup (uma vez):
#   1. Ter a key em ~/.config/garimpei/firebase-admin-key.json
#   2. pip3 install firebase-admin (ou pip3 install --break-system-packages firebase-admin)
#
# O script gera um Firebase ID token e chama a API via Cloudflare Worker.
# =============================================================================
set -euo pipefail

METHOD="${1:-GET}"
PATH_API="${2:-/api/health}"
BODY="${3:-}"

BASE_URL="https://garimpei.app.br"
PROJECT_ID="garimpo-500114"
API_KEY="AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A"
KEY_FILE="$HOME/.config/garimpei/firebase-admin-key.json"
TOKEN_CACHE="$HOME/.config/garimpei/.token-cache"
# UID do usuário principal (milenygsilva@gmail.com)
FIREBASE_UID="${GARIMPEI_UID:-GZsXcFj0xwcOXAJVt06y0PQYa7W2}"

if [ ! -f "$KEY_FILE" ]; then
  echo "❌ Service account key não encontrada: $KEY_FILE" >&2
  echo "   Gere com: gcloud iam service-accounts keys create ~/.config/garimpei/firebase-admin-key.json \\" >&2
  echo "     --iam-account=firebase-adminsdk-fbsvc@garimpo-500114.iam.gserviceaccount.com --project=garimpo-500114" >&2
  exit 1
fi

# Verifica se token cacheado é válido (tokens Firebase duram 1h)
if [ -f "$TOKEN_CACHE" ]; then
  CACHE_AGE=$(( $(date +%s) - $(stat -f %m "$TOKEN_CACHE" 2>/dev/null || stat -c %Y "$TOKEN_CACHE" 2>/dev/null || echo 0) ))
  if [ "$CACHE_AGE" -lt 3500 ]; then
    TOKEN=$(cat "$TOKEN_CACHE")
  fi
fi

# Gera token se não cacheado
if [ -z "${TOKEN:-}" ]; then
  TOKEN=$(python3 << PYEOF
import json, urllib.request
import firebase_admin
from firebase_admin import credentials, auth

cred = credentials.Certificate("$KEY_FILE")
firebase_admin.initialize_app(cred, {"projectId": "$PROJECT_ID"})

custom_token = auth.create_custom_token("$FIREBASE_UID")
ct = custom_token.decode() if isinstance(custom_token, bytes) else custom_token

# Troca custom token por ID token
data = json.dumps({"token": ct, "returnSecureToken": True}).encode()
req = urllib.request.Request(
    f"https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=$API_KEY",
    data=data,
    headers={"Content-Type": "application/json"}
)
resp = urllib.request.urlopen(req)
result = json.loads(resp.read())
print(result["idToken"])
PYEOF
  )
  # Cache o token
  mkdir -p "$(dirname "$TOKEN_CACHE")"
  echo "$TOKEN" > "$TOKEN_CACHE"
fi

# Faz a chamada
if [ -n "$BODY" ]; then
  curl -s -X "$METHOD" "${BASE_URL}${PATH_API}" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$BODY"
else
  curl -s -X "$METHOD" "${BASE_URL}${PATH_API}" \
    -H "Authorization: Bearer $TOKEN"
fi
echo ""
