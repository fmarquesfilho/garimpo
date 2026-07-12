# Handoff — Omnibox Input (2026-07-12)

> Próxima sessão: gerar tasks e implementar a spec `omnibox-input`.
> Branch: main (push direto, sem PR — MVP com 2 usuários).
> Spec completa em `.kiro/specs/omnibox-input/`.

## Estado atual

- **314 unit tests passando** (frontend vitest)
- **87 C# tests passando** (arch + integration)
- **13 Go packages passando** (inclui cache sidecar)
- **8 E2E checks passando** (test:e2e-novos contra produção)
- **9 drift checks passando**
- Cache Layer L1+L2 implementado e em produção (ADR-0031)
- BuscaContract unificado end-to-end (ADR-0030)
- CI sem path filtering — todos os jobs rodam sempre
- Bancos limpos (reset recente para testes do zero)

## O problema que a spec resolve

A página Descobrir usa **raias (lanes)** para organizar controles de busca:
- Raia de keyword (input de texto)
- Raia de lojas (combobox + autocomplete)
- Raia de categorias (combobox)
- Raia de filtros (comissão, vendas, fontes, marketplaces)
- Raia de buscas salvas (painel colapsável)

**Problema:** Muitas lanes = muito espaço vertical + muitos cliques + UX fragmentada.
O usuário precisa "saber onde olhar" para cada tipo de ação.

**Hipótese:** Um input unificado (omnibox) que infere o tipo pelo conteúdo reduz
a carga cognitiva e o espaço vertical, mantendo toda a funcionalidade.

## Decisão de design

### Omnibox = input único com inferência + prefixos opcionais

- Digita texto → sistema busca em keywords, lojas E categorias simultaneamente
- Dropdown de sugestões agrupado por tipo (Lojas, Categorias, Marketplaces, Buscas Salvas)
- Prefixos opcionais para power users: `@loja`, `#categoria`, `!marketplace`
- Filtros numéricos (comissão, vendas, nota) permanecem separados
- Frontend-only — sem novos endpoints, usa cache L2 existente

### Substituição direta (sem migração gradual)

O `Omnibox.svelte` substitui o `BuscaUnificada.svelte` diretamente na rota principal.
Sem rota `/v2`, sem feature flag — testa-se em produção com os 2 usuários.

### Pesquisa de componentes (OBRIGATÓRIO antes de implementar)

A próxima sessão deve avaliar se o **Bits UI Combobox** (já no projeto) suporta
o padrão de groups com headers ou se precisa de `cmdk-sv` (command palette com
fuzzy search e groups nativos). Critérios no `design.md` seção "Pesquisa de Componentes".

## Spec (3 docs)

| Arquivo | Status | O quê |
|---------|--------|-------|
| `.kiro/specs/omnibox-input/requirements.md` | ✅ Completo | 10 requisitos, ~40 AC |
| `.kiro/specs/omnibox-input/design.md` | ✅ Completo | Arquitetura, parser, sugestões, migração |
| `.kiro/specs/omnibox-input/tasks.md` | ⬜ A gerar | Implementation plan |

## Arquivos-chave para ler

### Frontend (Svelte 5 + Tailwind + shadcn-svelte)

| Arquivo | O quê | Relevância |
|---------|-------|-----------|
| `web/src/lib/components/BuscaUnificada.svelte` | Componente atual (lanes) | A substituir parcialmente |
| `web/src/lib/busca-engine.svelte.js` | BuscaEngine FSM (Svelte 5 runes) | Integração via events |
| `web/src/lib/busca-config.js` | Config derivada de rules.json | Funções puras existentes |
| `web/src/lib/busca-engine-effects.js` | Side-effects (API calls) | Reutilizar |
| `rules/busca-rules.json` | Fonte de verdade: intents, guards, transições | Adicionar seção `omnibox` |
| `web/src/lib/components/ui/` | Biblioteca shadcn-svelte (Combobox, Input, Button) | Usar como base |

### Backend (não precisa mudar, mas contexto útil)

| Arquivo | O quê |
|---------|-------|
| `src/Garimpei.Api/Endpoints/CoreEndpoints.cs` | `/api/candidatos` (keyword search via cache) |
| `src/Garimpei.Api/Endpoints/CuradoriaEndpoints.cs` | `/api/v2/curadoria/ranking` (shop search via cache) |
| `contracts/schemas/busca-contract.json` | BuscaContract (collection_keys, tipos) |

### Testes

| Arquivo | O quê |
|---------|-------|
| `web/src/tests/busca-engine.test.js` | Testes da FSM (padrão a seguir) |
| `web/src/tests/busca-config.test.js` | Testes de funções puras (padrão a seguir) |
| `web/src/tests/busca-unificada.test.js` | Testes do componente atual |

## Decisões já tomadas

1. **Prefixos são opcionais** — inferência é o default
2. **Texto literal no input** — sem chips dentro do campo
3. **Dropdown agrupado por tipo** — max 7 por grupo
4. **Buscas salvas como primeira sugestão** — se matcham
5. **Debounce 400ms** no evento DIGITAR (keyword search) — conforme rules
6. **Seleção de loja/categoria é imediata** — sem debounce
7. **Componente alternativo** — não substitui lanes existentes na V1
8. **3 módulos puros:** `omnibox-parser.js`, `omnibox-sugestoes.js`, `Omnibox.svelte`
9. **Config em rules/busca-rules.json** — seção `omnibox` (minChars, maxSugestoes, prefixos)
10. **Acessibilidade ARIA combobox** — role, aria-expanded, aria-activedescendant

## Próximos passos

1. Gerar `tasks.md` (implementation plan baseado no design)
2. Implementar parser + testes unitários (wave 1)
3. Implementar gerador de sugestões + testes (wave 2)
4. Implementar componente Svelte (wave 3)
5. Integrar com BuscaEngine + adicionar rota /v2 (wave 4)
6. Testar com Mileny, decidir se substitui lanes (wave 5)

## Steering rules ativas

- `git.md` — nunca `--no-verify`, nunca push automático
- `ci.md` — nunca E2E real no CI, deploy conservador
- `dependencies.md` — sem deps novas sem release recente

## Como verificar

```bash
# Testes unitários (rápido)
cd web && bunx vitest run

# Build (svelte-check zero errors)
cd web && bun run check

# Lint
cd web && bun run lint:js && bun run lint:css

# E2E contra produção (valida que nada quebrou)
mise run test:e2e-novos

# Todos os checks
mise run checks
```

## Contexto técnico extra

- **Svelte 5** com runes (`$state`, `$derived`, `$effect`) — NÃO usa stores/writable
- **shadcn-svelte** (Bits UI) — componentes em `web/src/lib/components/ui/`
- **Tailwind CSS v4** — utility-first, dark mode via tokens CSS
- **BuscaEngine** é headless — view despacha events, engine calcula estado
- **Testes Vitest** com `@testing-library/svelte` + happy-dom
- **Cache L2** responde em ~5ms para keywords já buscadas (X-Cache-Source: l2-hit)
- **Limite 400 linhas** por arquivo — CI bloqueia se exceder
