# Handoff — Omnibox Smart Search (2026-07-12)

> **Status: ✅ IMPLEMENTADO. Waves 1-5 completas. Deploy pendente.**
> Feature "Omnibox Smart Search". Spec em `.kiro/specs/omnibox-smart-search/`.
> ADR-0033 documenta o pattern Headless UI Controller.

## Resumo

Evolução do Omnibox para padrão **Headless UI Controller**: a BuscaEngine se torna o
controlador único de toda a UI da página Descobrir. Componentes passam a ser renderizadores
puros (zero lógica de decisão, zero $state local). Comportamento é configurável via
`rules/busca-rules.json`, verificável em CI, e observável via OpenTelemetry.

**Funcionalidades novas:**
- Smart Dropdown com detecção de intenção (Pesquisar Produtos, Pesquisar Lojas, Resolver Link, Categoria)
- Store Cards enriquecidos (imagem, seguidores, avaliação, bandeira de origem)
- Busca de lojas por nome (registro local — Shopee não expõe busca por nome via API)
- Monitoramento inline a partir dos resultados (sem scroll para o topo)
- Prefixos (@, #, !) continuam funcionando para power users

## O que já foi implementado (infra)

Estas mudanças já estão na branch `main`, compilando, testes passando:

1. **Proto `ResolveShopResponse` enriquecido** (+7 campos): follower_count, item_count,
   rating_star, image_url, cover_url, shop_location, description
2. **Collector Go** (`services/collector/server_shop.go`): parseia os campos extras de
   `shopee.com.br/api/v4/shop/get_shop_detail`
3. **C# API** (`POST /api/lojas/resolver`): retorna campos enriquecidos no response
4. **`buf generate`** executado: código Go + C# gerado

**Falta:**
- Migration para persistir os campos enriquecidos na tabela Lojas (Task 1)
- O restante das 14 tasks (endpoint buscar, handlers engine, componentes puros, OTel)

## Decisões técnicas confirmadas

| Decisão | Justificativa |
|---------|---------------|
| Engine como Headless UI Controller | Elimina estado disperso em componentes. Single source of truth. |
| Sub-estado `ui` separado de `ctx` | Domínio vs. apresentação. Permite resetar UI sem perder dados. |
| Omnibox emite eventos brutos | Componente não interpreta teclas — engine decide. |
| Configuração declarativa (JSON) | Mudança de comportamento = editar JSON, zero code. Verificável em CI. |
| Busca de lojas local-only | Shopee não expõe API de busca por nome. Registro enriquece via resolução de links. |
| Bandeira de origem = manual | API da Shopee não expõe país de origem. Mileny marca ao adicionar loja. |
| Campos enriquecidos persistidos (opção B) | Store Cards aparecem ricos na listagem sem re-resolver. |
| OTel spans por evento | Observabilidade de uso real para decisões de produto. |

## Arquitetura (pós-refactor)

```
rules/busca-rules.json
  └─ omnibox.intencao, storeCard.camposVisiveis, transicoes

BuscaEngine (Headless UI Controller)
  ├─ ctx.* (domínio: keyword, shopIds, resultados, fontes)
  ├─ ui.omnibox (inputValue, aberto, opcoes, highlightIdx, modo, placeholder)
  ├─ ui.resultados (modo: produtos|lojas, lojas[])
  ├─ ui.paineis (buscasSalvas, filtros, salvar)
  ├─ Handlers: OMNIBOX_INPUT, OMNIBOX_KEYDOWN, OMNIBOX_SELECIONAR, BUSCAR_LOJAS, MONITORAR_LOJA
  └─ OTel: span por send()

Componentes (renderizadores puros):
  Omnibox.svelte (~50 linhas, lê engine.omnibox, emite eventos brutos)
  StoreCard.svelte (configurável por marketplace via JSON)
  BuscaUnificada.svelte (condicional: produtos vs lojas)
```

## Waves de implementação (14 tasks)

| Wave | Tasks | Descrição |
|------|-------|-----------|
| 1 | 1, 2, 3 | Migration + endpoint buscar + módulo intencao (paralelas) |
| 2 | 4, 12 | Sub-estado ui na engine + config JSON expandida |
| 3 | 5, 6, 7 | Handlers engine (omnibox, buscar lojas, monitorar) — core |
| 4 | 8, 9, 10 | Componentes puros (Omnibox reescrito, StoreCard, BuscaUnificada) |
| 5 | 11, 13 | OTel + contracts/drift checks |
| 6 | 14 | Validação E2E + commit |

## Arquivos-chave

### Já modificados (prontos)
- `protos/collector/v1/collector.proto` — ResolveShopResponse enriquecido
- `services/collector/server_shop.go` — parseia campos extras
- `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` — response enriquecido no resolver

### A criar
- `web/src/lib/omnibox-intencao.js` — detecção de intenção (puro)
- `web/src/lib/components/StoreCard.svelte` — card configurável

### A refatorar significativamente
- `web/src/lib/busca-engine.svelte.js` — handlers OMNIBOX_*, sub-estado ui, OTel
- `web/src/lib/busca-engine-state.js` — criarUIInicial()
- `web/src/lib/components/Omnibox.svelte` — reescrever como renderizador puro
- `web/src/lib/components/BuscaUnificada.svelte` — migrar estado, condicional lojas
- `rules/busca-rules.json` — blocos omnibox.intencao, storeCard, transições novas

## Como rodar / verificar

```bash
# Backend
dotnet build src/Garimpei.sln        # build
dotnet test                           # 88 testes C#
dotnet ef database update             # aplicar migrations

# Frontend
cd web && bunx vitest run             # 367 testes
bun run check && bun run build        # svelte-check + build

# Checks completos
mise run checks                       # api-contract, schema-sync, rules-schema, etc.

# Playtest local contra prod
cd web && DEV_API_PROXY=https://garimpei.app.br VITE_API_BASE= bun run dev
```

## Restrições técnicas importantes

1. **Shopee não expõe busca de lojas por nome** — o endpoint `GET /api/lojas/buscar` opera
   exclusivamente sobre o registro local (tabela Lojas PostgreSQL).
2. **Shopee não expõe país de origem dos produtos** — a bandeira (🇰🇷) é `origem_padrao`,
   marcação manual da Mileny ao adicionar loja. Ver `docs/operacao-shopee`.
3. **Dados enriquecidos** (imagem, seguidores, avaliação) vêm do `get_shop_detail` da API
   pública Shopee — já são parseados pelo Collector Go.
4. **Testes são críticos** — essa feature é central. Cada task deve incluir seus testes
   unitários. A Task 14 (validação E2E) deve cobrir todos os fluxos principais.

## Referências

- Spec: `.kiro/specs/omnibox-smart-search/` (requirements.md, design.md, tasks.md)
- ADR-0027: BuscaEngine headless + regras externas
- ADR-0032: Store Workflow Registro de Lojas
- Handoff anterior: `docs/legado/HANDOFF_2026-07-12_STORE-WORKFLOW.md`
- Operação Shopee: `docs/operacao-shopee` (limitações da API)
- Docs-site live: `garimpei.app.br/docs`
