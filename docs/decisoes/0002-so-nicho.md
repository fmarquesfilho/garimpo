# ADR 0002 — Apenas estratégia nicho ativa

## Contexto

O motor de ranking implementa o Strategy pattern com duas estratégias: `nicho` e
`diversificada`. A "diversificada" foi removida da UI e não é usada em produção.

## Decisão

Apenas a estratégia **nicho** está ativa. A "diversificada" é dívida técnica
documentada — o código permanece mas não é feature exposta.

## Consequências

- Toda referência à "diversificada" como feature ativa é incorreta.
- O endpoint `/api/candidatos?estrategia=diversificada` continua funcionando
  (retrocompatibilidade) mas não é promovido na UI.
- O código do Strategy pode ser removido no futuro se não houver plano de reativação.
