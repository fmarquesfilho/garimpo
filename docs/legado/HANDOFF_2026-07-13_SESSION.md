# Handoff de Sessão — 2026-07-13

> **Status: Push feito. Deploy em produção. Próxima sessão implementa feed default.**
> Tudo commitado em main. 488 Vitest + 94 xUnit + 23 E2E local + 8 E2E prod = verdes.

## O que foi feito nesta sessão

### Features implementadas

1. **T-0060 Omnibox Smart Search** — BuscaEngine como Headless UI Controller
   - Smart Dropdown com detecção de intenção (produtos, lojas, categorias, resolver link)
   - StoreCard com dados enriquecidos (imagem, seguidores, avaliação, bandeira)
   - Busca de lojas por nome (endpoint local-only GET /api/lojas/buscar)
   - Monitoramento inline via Store Card
   - Migration EF Core (7 campos enriquecidos na tabela Lojas)
   - OTel spans por evento

2. **Omnibox Inline Chips** — UX unificada
   - Lojas = chips ouro (🏪), categorias = chips rosa (🏷️) DENTRO do input
   - Scope cards separados eliminados
   - X no chip remove do escopo (sincronizado)

3. **Pipeline de Qualidade de Testes**
   - Coverage v8 com thresholds por módulo
   - Stryker (JS + .NET) configurado
   - 15 tasks mise (test:fast, test:descobrir, test:mutate, etc.)
   - Paridade cross-language (21 pares fixture)
   - CI mutation testing job

4. **Enforcement Arquitetural**
   - ESLint bane $state em pure-renderers (compile-time)
   - Teste arch-pure-renderers.test.js (detecta novos componentes)
   - gen:auto no pre-push (api-reference sempre atualizada)

5. **Refatoração da Engine** — SRP
   - busca-engine.svelte.js (360 linhas, orquestrador)
   - busca-engine-omnibox.js (smart search)
   - busca-engine-lojas.js (adicionar/remover/resolver)
   - busca-engine-persistencia.js (salvar/carregar)
   - busca-engine-state.js (estado inicial)
   - busca-engine-effects.js (side effects)

6. **Limpeza**
   - Zero seletores obsoletos (UI v3 removida)
   - Zero código morto (knip)
   - Zero naming "legado"
   - fix: resolver link não muda modo para lojas

### Bugs corrigidos

- gen-board test incompatível com gitignore (drift check)
- Semgrep false positive (regexp from trusted config)
- bun.lock desatualizado
- Resolver link setava modo=lojas sem dados (mostrava "Nenhuma loja encontrada" após sucesso)

---

## O que precisa ser implementado AGORA (próxima sessão)

### Feature: Feed Default on-load (produtos sem interação do usuário)

**Problema:** A página Descobrir abre vazia se o usuário não tem buscas salvas e não digita nada. O empty state "Nenhum resultado" é confuso — deveria mostrar produtos interessantes automaticamente.

**Decisão:** Implementar via config-driven, marketplace-agnóstico.

**Arquitetura proposta:**

```
rules/busca-rules.json
  └─ feedDefault: {
       habilitado: true,
       estrategia: "categoria_top_vendidos",
       categorias: [
         { marketplace: "shopee", categoryId: 100630, nome: "Beleza", keyword: "beleza" },
         { marketplace: "shopee", categoryId: 100664, nome: "Cuidados com a Pele", keyword: "skincare" },
         { marketplace: "shopee", categoryId: 100640, nome: "Perfumaria", keyword: "perfume" }
       ],
       rotacao: "random",  // random | sequential
       limit: 20,
       sortBy: "sales"
     }
```

**Implementação em 2 fases:**

#### Fase 1 (pragmática — hoje)
- Adicionar bloco `feedDefault` ao JSON
- Engine: no boot, se não há contexto, pegar a primeira categoria do feedDefault e disparar busca com keyword dela
- O `Fetch` RPC existente já aceita keyword + sort_by=sales
- Frontend funciona imediato com dados reais
- Genérico: trocar categorias no JSON muda o feed sem code change

#### Fase 2 (robusta — spec separada)
- Nova RPC `FetchCategory(category_id, marketplace, sort, limit)` no proto
- Collector Go implementa via Shopee Affiliate API `productOfferV2` (retorna comissão)
- C# chama via gRPC, cachea, serve sem keyword
- Suporta Amazon (Browse Nodes), Mercado Livre (categorias MLA)
- Elimina o "hack" de keyword como proxy de categoria

**Arquivos a modificar (Fase 1):**
- `rules/busca-rules.json` — adicionar `feedDefault`
- `rules/busca-rules.schema.json` — validar novo bloco
- `web/src/lib/busca-config.js` — exportar FEED_DEFAULT
- `web/src/lib/busca-engine.svelte.js` — alterar `#inicializar()` para usar feedDefault
- `web/src/tests/busca-engine-omnibox.test.js` — testes do boot com feed default
- `web/tests/local/smart-search.spec.js` — E2E que valida produtos on-load
- `web/tests/prod/descobrir.spec.js` — ajustar "empty state" para aceitar feed

**Critério de sucesso:**
- Abrir garimpei.app.br sem digitar nada → produtos aparecem (top vendidos de uma categoria)
- Rotação: cada reload pode mostrar categoria diferente
- A busca manual sobrescreve o feed (DIGITAR limpa o feedDefault e busca normalmente)
- Config-driven: mudar categorias no JSON = mudar feed sem deploy de código
- Funciona para qualquer marketplace (estrutura genérica)

**Testes esperados:**
- Unit: engine boot dispara busca com keyword do feedDefault quando ctx vazio
- Unit: se há buscas salvas com lojas, prefere carregar a última (não usa feedDefault)
- E2E local: página abre com produtos visíveis (mock retorna candidatos)
- E2E prod: página abre com produtos da Shopee (dados reais)

---

## Estado da base de código

| Métrica | Valor |
|---------|-------|
| Vitest | 488 |
| xUnit | 94 |
| E2E local | 23 |
| E2E prod | 8 |
| Arquivos engine | 6 módulos (max 360 linhas) |
| svelte-check | 0 erros |
| Seletores obsoletos | 0 |
| Código morto | 0 |
| Deploy | Frontend ✅ Backend ✅ |

## Como rodar

```bash
# Testes rápidos
mise run test:fast          # unit only, <10s
mise run test:descobrir     # só página Descobrir

# Testes completos
mise run test:unit          # 488 vitest + 94 xUnit
mise run test:e2e-smart-search  # 23 E2E local

# Produção
mise run test:e2e-prod -- tests/prod/smart-search.spec.js
mise run test:e2e-prod -- tests/prod/descobrir.spec.js

# Checks
bun run check && bun run build
mise run check:pure-renderers
mise run gen:auto
```

## Referências

- ADR-0033: docs/decisoes/0033-headless-ui-controller-omnibox.md
- Spec Smart Search: .kiro/specs/omnibox-smart-search/
- Spec Inline Chips: .kiro/specs/omnibox-inline-chips/
- Spec Test Pipeline: .kiro/specs/test-quality-pipeline/
- Backlog: T-0060 (done)
