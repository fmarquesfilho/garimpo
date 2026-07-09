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

## Atualização 2026-07-09 — Layout em raias

A página foi reorganizada em **raias horizontais** (metáfora de piscina), cada uma
agrupando um tipo de configuração da busca. De cima para baixo:

1. **Console superior** — input de palavras-chave + três botões que controlam os grupos
   (Filtros, Lojas, Buscas), cada um com um **contador** de quantas configurações estão
   aplicadas naquele grupo. Ainda no topo: **colapsar tudo** e **limpar tudo**.
2. **Raia Filtros** (duas sub-raias) — em cima, os toggles de fontes (Novos, Quedas,
   Favoritos) à esquerda e os filtros quantitativos (comissão mín., vendas mín.) à direita;
   embaixo, a filtragem de **categorias** via autocomplete. Categorias são as de 1º nível
   extraídas por marketplace; o dropdown mostra o nome à esquerda e os marketplaces a que
   pertence à direita. Cada categoria adicionada vira um card (com seus marketplaces).
3. **Raia Lojas** — autocomplete de lojas para escopar a busca. Lista as **lojas
   monitoradas** (nome + marketplace) e permite **adicionar uma loja nova por link/ID**
   (opção "↳ resolver e adicionar" — fluxo recorrente da operação). Cada loja no escopo
   vira um card com nome, marketplace, **bandeira de origem** (China/Japão/Coreia — ver
   Operação Shopee) e indicador de monitoramento (temporizador com o ciclo, ou "sem
   monitor"). Quando há lojas no escopo, a busca só roda nelas.
4. **Raia Buscas** — buscas salvas e agendadas. Cada card é dividido em seções
   (palavras-chave, categorias, lojas, marketplaces) exibidas conforme presentes, além da
   info de agendamento. Todo card é editável (**edit mode**): altera a busca e re-salva a
   mesma (via `id`), podendo reagendar coletas periódicas.

Cada raia tem seu próprio **limpar raia** e pode ser colapsada individualmente ou em
conjunto. Regras de estado e eventos: ver **ADR-0027**. Componentes: ver `componentes.md`.
