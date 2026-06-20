#!/usr/bin/env bash
# Exemplo de coleta periódica via Cloud Scheduler -> Cloud Run.
# Cada job chama POST /api/coletar?categoria=... no cron definido, e o backend
# grava o snapshot (top N do momento) na tabela `snapshots` do BigQuery.
#
# AJUSTE as variáveis e rode os blocos que quiser. NÃO commite o token real.
set -euo pipefail

PROJECT_ID="seu-projeto"
REGION="southamerica-east1"
SERVICE="garimpo-api"

# URL pública do Cloud Run:
URL="$(gcloud run services describe "$SERVICE" --region "$REGION" \
        --format='value(status.url)')"

# Token compartilhado que protege o endpoint de coleta -----------------------
# 1) crie o segredo e ligue na revisão do Cloud Run:
#    printf '%s' "$(openssl rand -hex 24)" | gcloud secrets create COLETA_TOKEN --data-file=-
#    gcloud run services update "$SERVICE" --region "$REGION" \
#      --update-secrets COLETA_TOKEN=COLETA_TOKEN:latest
# 2) pegue o valor para os jobs do Scheduler (eles mandam no header):
TOKEN="$(gcloud secrets versions access latest --secret=COLETA_TOKEN)"

# Função utilitária: cria um job de coleta para uma categoria ----------------
criar_coleta() {
  local nome="$1" categoria="$2" keyword="$3" cron="$4"
  gcloud scheduler jobs create http "coleta-$nome" \
    --location "$REGION" \
    --schedule "$cron" \
    --time-zone "America/Sao_Paulo" \
    --uri "$URL/api/coletar?fonte=shopee&categoria=$categoria&keyword=$keyword&estrategia=nicho&top=20&vendas_min=5&nota_min=4" \
    --http-method POST \
    --headers "X-Garimpo-Token=$TOKEN" \
    --attempt-deadline 60s
}

# Categorias do nicho, com horários ESPAÇADOS para respeitar o rate limit -----
# (não dispare tudo no mesmo minuto)
criar_coleta "perfumaria"  "perfumaria" "perfume"  "0 8 * * *"   # 08:00 todo dia
criar_coleta "skincare"    "cosméticos" "skincare" "10 8 * * *"  # 08:10
criar_coleta "maquiagem"   "cosméticos" "batom"    "20 8 * * *"  # 08:20
criar_coleta "bem-estar"   "bem-estar"  "massageador" "30 8 * * *" # 08:30

echo "Jobs criados. Liste com: gcloud scheduler jobs list --location $REGION"
echo "Teste manual:"
echo "  curl -X POST -H \"X-Garimpo-Token: \$TOKEN\" \"$URL/api/coletar?fonte=shopee&categoria=perfumaria&keyword=perfume&top=20\""
