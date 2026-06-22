#!/usr/bin/env bash
# Coleta periódica dirigida pelas BUSCAS SALVAS.
#
# Lê os perfis de busca do servidor (GET /api/buscas, sincronizados no BigQuery)
# e cria um job no Cloud Scheduler por KEYWORD de cada perfil que TENHA cron.
# Cada job chama POST /api/coletar com os filtros do perfil.
#
# O modelo novo usa `keywords` (array). Para cada perfil com N keywords,
# são criados N jobs (um por keyword), com nomes coleta-<id>-0, coleta-<id>-1...
# Perfis com cron vazio são ignorados.
#
# Pré-requisitos: gcloud autenticado, `jq` instalado, COLETA_TOKEN criado.
# Uso:
#   ./deploy/scheduler-exemplo.sh
#   DRY_RUN=1 ./deploy/scheduler-exemplo.sh   # só mostra o que faria
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
  id="$(echo "$b"         | jq -r '.id // ""')"
  categoria="$(echo "$b"  | jq -r '.categoria // ""')"
  estrategia="$(echo "$b" | jq -r '.estrategia // "nicho"')"
  top="$(echo "$b"        | jq -r '.top // 20')"
  vendas="$(echo "$b"     | jq -r '.vendas_min // 5')"
  nota="$(echo "$b"       | jq -r '.nota_min // 0')"
  cron="$(echo "$b"       | jq -r '.cron')"

  # suporta tanto keywords[] (novo) quanto keyword string (legado)
  keywords="$(echo "$b" | jq -r 'if .keywords then .keywords[] else .keyword end' 2>/dev/null || echo "")"

  if [ -z "$keywords" ]; then
    echo "→ Busca '$id' sem keywords, pulando."
    continue
  fi

  idx=0
  while IFS= read -r keyword; do
    [ -z "$keyword" ] && continue

    # nome de job sanitizado (Scheduler aceita [a-z0-9-], máx 64 chars)
    suffix="$(echo "$keyword" | tr '[:upper:] ' '[:lower:]-' | tr -cd 'a-z0-9-')"
    job="coleta-${id}-${suffix}"
    job="${job:0:63}"  # trunca para o limite do Scheduler

    alvo="$URL/api/coletar?fonte=shopee"
    alvo+="&keyword=$(jq -rn --arg v "$keyword" '$v|@uri')"
    alvo+="&categoria=$(jq -rn --arg v "$categoria" '$v|@uri')"
    alvo+="&estrategia=$estrategia&top=$top&vendas_min=$vendas&nota_min=$nota"

    echo "→ $job  cron='$cron'  keyword='$keyword'"
    if [ "${DRY_RUN:-0}" = "1" ]; then
      idx=$((idx + 1))
      continue
    fi

    gcloud scheduler jobs create http "$job" \
      --location "$REGION" --schedule "$cron" --time-zone "$TZ_AGENDA" \
      --uri "$alvo" --http-method POST \
      --headers "X-Garimpo-Token=$TOKEN" --attempt-deadline 60s 2>/dev/null \
    || gcloud scheduler jobs update http "$job" \
      --location "$REGION" --schedule "$cron" --time-zone "$TZ_AGENDA" \
      --uri "$alvo" --http-method POST \
      --update-headers "X-Garimpo-Token=$TOKEN" --attempt-deadline 60s

    idx=$((idx + 1))
  done <<< "$keywords"
done

echo "Pronto. Liste com: gcloud scheduler jobs list --location $REGION"
