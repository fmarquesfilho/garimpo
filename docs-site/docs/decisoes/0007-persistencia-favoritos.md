# ADR 0007 — Persistência de favoritos: localStorage + sync BigQuery

## Contexto

Favoritos precisam de acesso instantâneo (sem latência de rede) mas também
precisam de backup no servidor para não perder dados ao trocar de dispositivo.

## Decisão

- **localStorage** para acesso imediato (frontend-first).
- **Sync para BigQuery** (servidor) como backup e para análises.

## Consequências

- Frontend funciona offline para leitura de favoritos.
- Sync acontece em background quando online.
- Conflitos resolvidos por last-write-wins (timestamp `salvo_em`).
