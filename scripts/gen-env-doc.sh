#!/usr/bin/env bash
# scripts/gen-env-doc.sh — Extrai variáveis de ambiente referenciadas no código Go
# e gera docs/gerado/env-vars.md.
#
# Uso: ./scripts/gen-env-doc.sh > docs/gerado/env-vars.md
# Ou via Makefile: make docs-env

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cat <<'HEADER'
---
title: Variáveis de ambiente
description: Lista de variáveis de ambiente extraídas do código-fonte.
---

:::caution[Arquivo gerado]
Não edite manualmente. Rode `make docs-env` para regenerar.
:::

| Variável | Arquivo | Linha |
|---|---|---|
HEADER

# Busca por os.Getenv("...") e os.LookupEnv("...") nos fontes Go
# sort -t'|' -k2,2 -k3,3: ordena por variável, depois por arquivo (determinístico)
# awk: mantém apenas a primeira ocorrência de cada variável
grep -rn "os\.Getenv\|os\.LookupEnv" "$ROOT" \
  --include="*.go" \
  --exclude-dir=vendor \
  --exclude-dir=.git \
  | sed "s|${ROOT}/||" \
  | while IFS='' read -r match; do
      file="${match%%:*}"
      rest="${match#*:}"
      line="${rest%%:*}"
      var=$(echo "$match" | sed -n 's/.*Getenv("\([^"]*\)".*/\1/p')
      if [ -z "$var" ]; then
        var=$(echo "$match" | sed -n 's/.*LookupEnv("\([^"]*\)".*/\1/p')
      fi
      if [ -n "$var" ]; then
        echo "| \`$var\` | \`$file\` | $line |"
      fi
    done \
  | sort -t'|' -k2,2 -k3,3 \
  | awk -F'|' '!seen[$2]++' 
