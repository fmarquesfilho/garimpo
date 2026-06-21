#!/usr/bin/env bash
# Coleta periódica dirigida pelas BUSCAS SALVAS.
#
# Lê os perfis de busca do servidor (GET /api/buscas, sincronizados no BigQuery)
# e cria um job no Cloud Scheduler por perfil que TENHA cron, chamando
# POST /api/coletar com os filtros do perfil. Cada coleta grava um snapshot
# (com timestamp) na tabela `snapshots`.
#
# Pré-requisitos: gcloud autenticado, `jq` instalado, e o COLETA_TOKEN criado
# (ver docs/DEPLOY_GCP.md → Coleta periódica). NÃO commite o token.
#
# Uso:
#   ./deploy/scheduler-exemplo.sh                 # cria/atualiza os jobs
#   DRY_RUN=1 ./deploy/scheduler-exemplo.sh       # só mostra o que faria
set -euo pipefail

PROJECT_ID="${PROJECT_ID:-garimpo-500114}"
REGION="${REGION:-southamerica-east1}"
SERVICE="${SERVICE:-garimpo-api}"
TZ_AGENDA="${TZ_AGENDA:-America/Sao_Paulo}"

URL="$(gcloud run services describe "$SERVICE" --region "$REGION" --format='value(status.url)')"
TOKEN="$(gcloud secrets versions access latest --secret=COLETA_TOKEN)"

echo "Servidor: $URL"
echo "Lendo buscas salvas de $URL/api/buscas ..."
BUSCAS_JSON="$(curl -fsS "$URL/api/buscas")"

# percorre cada busca com cron não-vazio
echo "$BUSCAS_JSON" | jq -c '.buscas[]? | select(.cron != null and .cron != "")' | while read -r b; do
  nome="$(echo "$b"      | jq -r '.nome')"
  keyword="$(echo "$b"   | jq -r '.keyword // ""')"
  categoria="$(echo "$b" | jq -r '.categoria // ""')"
  estrategia="$(echo "$b"| jq -r '.estrategia // "nicho"')"
  top="$(echo "$b"       | jq -r '.top // 20')"
  vendas="$(echo "$b"    | jq -r '.vendas_min // 5')"
  nota="$(echo "$b"      | jq -r '.nota_min // 0')"
  cron="$(echo "$b"      | jq -r '.cron')"

  # nome de job sanitizado (Scheduler aceita [a-z0-9-])
  job="coleta-$(echo "$nome" | tr '[:upper:] ' '[:lower:]-' | tr -cd 'a-z0-9-')"
  alvo="$URL/api/coletar?fonte=shopee&keyword=$(jq -rn --arg v "$keyword" '$v|@uri')&categoria=$(jq -rn --arg v "$categoria" '$v|@uri')&estrategia=$estrategia&top=$top&vendas_min=$vendas&nota_min=$nota"

  echo "→ $job  ($cron)  keyword='$keyword' categoria='$categoria'"
  if [ "${DRY_RUN:-0}" = "1" ]; then continue; fi

  # cria OU atualiza (create falha se já existe → cai no update)
  gcloud scheduler jobs create http "$job" \
    --location "$REGION" --schedule "$cron" --time-zone "$TZ_AGENDA" \
    --uri "$alvo" --http-method POST \
    --headers "X-Garimpo-Token=$TOKEN" --attempt-deadline 60s 2>/dev/null \
  || gcloud scheduler jobs update http "$job" \
    --location "$REGION" --schedule "$cron" --time-zone "$TZ_AGENDA" \
    --uri "$alvo" --http-method POST \
    --update-headers "X-Garimpo-Token=$TOKEN" --attempt-deadline 60s
done

echo "Pronto. Liste com: gcloud scheduler jobs list --location $REGION"
