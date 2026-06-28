---
title: "ADR 0006 — Campo categorias é plural (array)"
---


## Contexto

O schema original usava `categoria` (singular, string). A evolução do produto
precisa de múltiplas categorias por busca com filtro OR.

## Decisão

O campo é **`categorias[]`** (array de strings, filtro por OR).
O campo `categoria` singular é legado — mantido por retrocompatibilidade mas
a fonte de verdade é o array.

## Consequências

- Frontend e API usam `categorias[]`.
- BigQuery mantém `categoria` STRING na tabela de buscas (compatibilidade).
- Docs devem referenciar `categorias[]` como o formato atual.
