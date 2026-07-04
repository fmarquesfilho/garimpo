#!/usr/bin/env bash
# =============================================================================
# setup-cloud-tasks.sh — Cria a queue de alertas no Cloud Tasks
# =============================================================================
# Uso: ./deploy/setup-cloud-tasks.sh
#
# Idempotente — pode rodar múltiplas vezes sem problema.
# A queue "price-alerts" é configurada com:
#   - max_dispatches_per_second: 1 (respeita rate limit Telegram: 1msg/s por chat)
#   - max_concurrent_dispatches: 1 (sequencial)
#   - max_attempts: 5 (retry com backoff)
#   - min_backoff: 10s, max_backoff: 300s
# =============================================================================
set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:-garimpo-500114}"
LOCATION="${CLOUD_TASKS_LOCATION:-southamerica-east1}"
QUEUE_ID="${ALERT_QUEUE_ID:-price-alerts}"

echo "🔧 Criando Cloud Tasks queue: $QUEUE_ID"
echo "   Project: $PROJECT_ID"
echo "   Location: $LOCATION"
echo ""

# Verifica se já existe
if gcloud tasks queues describe "$QUEUE_ID" \
  --project="$PROJECT_ID" \
  --location="$LOCATION" &>/dev/null; then
  echo "✅ Queue '$QUEUE_ID' já existe. Atualizando configuração..."
else
  echo "📦 Criando queue '$QUEUE_ID'..."
  gcloud tasks queues create "$QUEUE_ID" \
    --project="$PROJECT_ID" \
    --location="$LOCATION"
fi

# Configura rate limits e retry
gcloud tasks queues update "$QUEUE_ID" \
  --project="$PROJECT_ID" \
  --location="$LOCATION" \
  --max-dispatches-per-second=1 \
  --max-concurrent-dispatches=1 \
  --max-attempts=5 \
  --min-backoff=10s \
  --max-backoff=300s

echo ""
echo "✅ Queue configurada:"
gcloud tasks queues describe "$QUEUE_ID" \
  --project="$PROJECT_ID" \
  --location="$LOCATION" \
  --format="table(name,state,rateLimits.maxDispatchesPerSecond,retryConfig.maxAttempts)"
echo ""
echo "📋 Para testar manualmente:"
echo "   gcloud tasks create-http-task \\"
echo "     --queue=$QUEUE_ID --location=$LOCATION --project=$PROJECT_ID \\"
echo "     --url=https://garimpei-v2-vj6afttbza-rj.a.run.app/internal/alerts/check \\"
echo "     --method=POST --body-content='{\"keyword\":\"loja-920292999\",\"threshold\":0.15}' \\"
echo "     --oidc-service-account-email=garimpo-api-sa@$PROJECT_ID.iam.gserviceaccount.com"
