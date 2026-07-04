#!/usr/bin/env bash
# =============================================================================
# test-alerts.sh — Testar o fluxo de alertas de preço end-to-end
# =============================================================================
#
# ESTRATÉGIA DE TESTE (3 camadas):
#
# 1. LOCAL (sem Telegram real)
#    Roda o endpoint /internal/alerts/check contra analyzer mock.
#    Publisher está offline → verifica que a lógica funciona até o ponto de envio.
#
# 2. LOCAL + PUBLISHER (com Telegram real)
#    Roda com publisher real apontando para um chat de teste.
#    Confirma que a mensagem chega formatada no Telegram.
#
# 3. PRODUÇÃO
#    Cria uma Cloud Task manualmente que dispara o fluxo completo.
#    Usa dados reais do BigQuery.
#
# =============================================================================
set -euo pipefail

MODE="${1:-local}"

case "$MODE" in
  # ─── Camada 1: Local sem envio ───────────────────────────────────────────
  local)
    echo "🧪 Teste LOCAL (analyzer mock, sem publisher)"
    echo "   Verifica: endpoint recebe, chama analyzer, formata, retorna OK"
    echo ""

    # 1. Subir analyzer mock
    echo "1️⃣  Subindo analyzer mock (porta 8060)..."
    MOCK_DATA=true python3 -m uvicorn main:app --host 127.0.0.1 --port 8060 \
      --app-dir services/analyzer &
    ANALYZER_PID=$!
    sleep 2

    # 2. Subir API C# em Development (sem publisher real)
    echo "2️⃣  Subindo API C# (porta 5000)..."
    ASPNETCORE_ENVIRONMENT=Development \
    Analyzer__BaseUrl=http://localhost:8060 \
    dotnet run --project src/Garimpei.Api --no-build &
    API_PID=$!
    sleep 6

    # 3. Chamar o endpoint
    echo "3️⃣  Chamando POST /internal/alerts/check..."
    echo ""
    RESPONSE=$(curl -s -X POST http://localhost:5000/internal/alerts/check \
      -H "Content-Type: application/json" \
      -d '{"keyword":"perfumes-importados","threshold":0.15,"owner_uid":"dev-user-001"}')

    echo "📋 Resposta:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
    echo ""

    # Verificar resultado
    ALERTS=$(echo "$RESPONSE" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('alerts_sent', d.get('drops', 0)))" 2>/dev/null || echo "0")

    # Cleanup
    kill $API_PID $ANALYZER_PID 2>/dev/null || true
    wait $API_PID $ANALYZER_PID 2>/dev/null || true

    if echo "$RESPONSE" | grep -q "drops"; then
      echo "✅ Endpoint funcionou! Detectou quedas e tentou enviar."
      echo "   (Publisher offline = gRPC error esperado neste modo)"
    elif echo "$RESPONSE" | grep -q "analyzer_unavailable"; then
      echo "⚠️  Analyzer não respondeu. Verifique se a porta 8060 está livre."
    else
      echo "ℹ️  Resposta: $RESPONSE"
    fi
    ;;

  # ─── Camada 2: Local com envio real ─────────────────────────────────────
  telegram)
    echo "🧪 Teste LOCAL + PUBLISHER REAL (envia no Telegram)"
    echo "   ⚠️  Vai enviar mensagem de teste no chat configurado!"
    echo ""
    echo "   Pré-requisitos:"
    echo "   - Docker rodando (para PostgreSQL)"
    echo "   - Variáveis TELEGRAM_BOT_TOKEN e ALERTAS_TELEGRAM_CHAT_ID definidas"
    echo ""

    if [ -z "${TELEGRAM_BOT_TOKEN:-}" ] || [ -z "${ALERTAS_TELEGRAM_CHAT_ID:-}" ]; then
      echo "❌ Defina TELEGRAM_BOT_TOKEN e ALERTAS_TELEGRAM_CHAT_ID"
      echo "   export TELEGRAM_BOT_TOKEN=seu_token"
      echo "   export ALERTAS_TELEGRAM_CHAT_ID=seu_chat_id"
      exit 1
    fi

    read -p "Enviar alerta de teste no Telegram? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Cancelado."
      exit 0
    fi

    echo "1️⃣  Subindo analyzer mock..."
    MOCK_DATA=true python3 -m uvicorn main:app --host 127.0.0.1 --port 8060 \
      --app-dir services/analyzer &
    ANALYZER_PID=$!
    sleep 2

    echo "2️⃣  Subindo publisher (porta 50052)..."
    PORT=50052 \
    TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" \
    go run ./services/publisher/ &
    PUB_PID=$!
    sleep 2

    echo "3️⃣  Subindo API C# (porta 5000)..."
    ASPNETCORE_ENVIRONMENT=Development \
    Analyzer__BaseUrl=http://localhost:8060 \
    Alerts__TelegramChatId="$ALERTAS_TELEGRAM_CHAT_ID" \
    dotnet run --project src/Garimpei.Api --no-build &
    API_PID=$!
    sleep 6

    echo "4️⃣  Chamando POST /internal/alerts/check..."
    RESPONSE=$(curl -s -X POST http://localhost:5000/internal/alerts/check \
      -H "Content-Type: application/json" \
      -d '{"keyword":"perfumes-importados","threshold":0.15,"owner_uid":"dev-user-001"}')

    echo ""
    echo "📋 Resposta:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"

    # Cleanup
    kill $API_PID $PUB_PID $ANALYZER_PID 2>/dev/null || true
    wait $API_PID $PUB_PID $ANALYZER_PID 2>/dev/null || true

    if echo "$RESPONSE" | grep -q '"alerts_sent":1'; then
      echo ""
      echo "✅ Alerta enviado com sucesso! Verifique o Telegram."
    else
      echo ""
      echo "⚠️  Verifique a resposta acima."
    fi
    ;;

  # ─── Camada 3: Produção (via Cloud Tasks manual) ────────────────────────
  prod)
    echo "🧪 Teste PRODUÇÃO (cria Cloud Task manual)"
    echo "   Usa dados reais do BigQuery."
    echo "   A task será processada pelo Cloud Run em ~1s."
    echo ""

    PROJECT_ID="garimpo-500114"
    LOCATION="southamerica-east1"
    QUEUE_ID="price-alerts"
    TARGET_URL="https://garimpei-v2-vj6afttbza-rj.a.run.app/internal/alerts/check"
    SA_EMAIL="garimpo-api-sa@$PROJECT_ID.iam.gserviceaccount.com"

    # Verificar se queue existe
    if ! gcloud tasks queues describe "$QUEUE_ID" \
      --project="$PROJECT_ID" --location="$LOCATION" &>/dev/null; then
      echo "❌ Queue '$QUEUE_ID' não existe. Rode primeiro:"
      echo "   ./deploy/setup-cloud-tasks.sh"
      exit 1
    fi

    KEYWORD="${2:-loja-920292999}"
    echo "   Keyword: $KEYWORD"
    echo "   Queue: $QUEUE_ID"
    echo ""

    read -p "Criar task de alerta em produção? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Cancelado."
      exit 0
    fi

    # Criar task via gcloud
    TASK_NAME="test-alert-$(date +%s)"
    BODY='{"keyword":"'"$KEYWORD"'","threshold":0.15}'

    gcloud tasks create-http-task "$TASK_NAME" \
      --queue="$QUEUE_ID" \
      --location="$LOCATION" \
      --project="$PROJECT_ID" \
      --url="$TARGET_URL" \
      --method=POST \
      --header="Content-Type: application/json" \
      --body-content="$BODY" \
      --oidc-service-account-email="$SA_EMAIL"

    echo ""
    echo "✅ Task criada: $TASK_NAME"
    echo "   Será processada em ~1s pelo Cloud Run."
    echo ""
    echo "   Verificar logs:"
    echo "   gcloud run services logs read garimpei-v2 --region=$LOCATION --project=$PROJECT_ID --limit=10"
    ;;

  *)
    echo "Uso: $0 [local|telegram|prod] [keyword]"
    echo ""
    echo "  local    — Testa lógica sem envio (analyzer mock, publisher offline)"
    echo "  telegram — Testa com envio real no Telegram (requer tokens)"
    echo "  prod     — Cria Cloud Task em produção (dados reais do BQ)"
    ;;
esac
