# ADR 0008 — Alertas de preço: implementados mas desabilitados

## Contexto

O backend tem alertas de "produtos novos" e variação de preço implementados.
Falta a configuração por usuário (cada tenant define threshold e canais).

## Decisão

Alertas estão **implementados no backend** mas **desabilitados por padrão**
até que a configuração por usuário esteja disponível.

## Consequências

- Código de alertas permanece e é testado.
- Não há trigger automático em produção (exceto testes manuais via `/api/alertas/testar`).
- Será habilitado quando o onboarding incluir configuração de alertas.
