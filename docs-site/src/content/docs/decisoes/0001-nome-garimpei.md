---
title: "ADR 0001 — Nome do produto: Garimpei"
---


## Contexto

O binário e os serviços internos usam `garimpo`/`garimpo-api` (nome técnico original).
A marca voltada ao público e o domínio são `garimpei.app.br`.

Havia ambiguidade nos docs: alguns chamavam de "Garimpo", outros de "Garimpei".

## Decisão

- **Garimpei** é o nome do produto/marca.
- `garimpo` / `garimpo-api` permanecem como nome técnico (binário, Cloud Run, dataset BQ).
- Documentação voltada ao usuário sempre usa "Garimpei".

## Consequências

- Docs antigos que dizem "Garimpo" como nome do produto devem ser atualizados.
- Código e infra não precisam renomear — custo alto sem benefício.
