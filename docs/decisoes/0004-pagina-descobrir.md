# ADR 0004 — Página principal: Descobrir

## Contexto

Historicamente a UI usou nomes como "Busca", "Oportunidades", "Curadoria" para a
página principal. Isso gerava confusão nos docs.

## Decisão

A página principal é **Descobrir** (`/`). Unifica as funcionalidades antes separadas.
"Curadoria", "Buscar", "Oportunidades" são nomes legados.

## Consequências

- Docs devem usar "Descobrir" ao se referir à página principal.
- Rotas internas do frontend podem manter `/` sem rename.
