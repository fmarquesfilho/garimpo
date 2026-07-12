# Handoff — Omnibox Input (2026-07-12)

> **Status: ✅ IMPLEMENTADO, MERGED em `main`, DEPLOY em produção.**
> Feature "Omnibox" (T-0055). Spec completa em `.kiro/specs/omnibox-input/`.
> Este doc é o ponto de retomada para o próximo agente/dev.

## O que foi feito

O sistema de **raias (lanes)** da página Descobrir foi substituído por um
**input unificado (omnibox)** que infere o tipo pelo conteúdo e aceita prefixos
opcionais: `@loja`, `#categoria`, `!marketplace`. Ex.: `serum @lebotanic #beleza !shopee`.
O dropdown de sugestões é agrupado por tipo (buscas salvas primeiro). Filtros
numéricos (comissão, vendas, fontes, marketplaces) permanecem como controles
separados abaixo do input. É 100% frontend — sem novos endpoints.

- **Merge:** commits `3a0f938` (feat) + `d787786` (chore: dev API proxy) em `main`.
- **Deploy:** CI verde (frontend Cloudflare Pages + backend Cloud Run). Live em `garimpei.app.br`.
- **Testes:** +51 novos (365 vitest no total); `check` 0/0; `build` OK; `test:e2e-novos` 8/8;
  87 C# tests; lint/semgrep/docs-drift/rules-schema OK.
- ⚠️ **Não houve playtest manual** (o login do app depende de Google OAuth + 2FA do dono).
  A validação real da UX acontece em produção com os 2 usuários (filosofia "test in prod").

## Arquitetura entregue

```
Omnibox.svelte (input + dropdown agrupado, ARIA, teclado)
  ├─ omnibox-parser.js    (puro)  parsearInput / serializarTokens / tokensParaContexto
  ├─ omnibox-sugestoes.js (puro)  gerarSugestoes → Map<tipo, Sugestao[]>
  └─ despacha eventos → BuscaEngine (FSM headless, inalterada)
BuscaUnificada.svelte  = dono da engine; renderiza Omnibox + filtros + cards + buscas salvas
rules/busca-rules.json = bloco `omnibox` (prefixos, minChars:2, maxSugestoes:7, debounceMs:400)
```

| Arquivo | Papel |
|---------|-------|
| `web/src/lib/omnibox-parser.js` | Tokeniza texto multi-token; resolve → ctx |
| `web/src/lib/omnibox-sugestoes.js` | Gera sugestões agrupadas (loja/categoria/marketplace/busca_salva) |
| `web/src/lib/components/Omnibox.svelte` | Componente visual (hand-rolled, ~220 linhas) |
| `web/src/lib/components/BuscaUnificada.svelte` | View: omnibox + filtros + escopo + salvas |
| `web/src/lib/busca-config.js` | Exporta `OMNIBOX` (de `rules.omnibox`) |
| `web/src/lib/components/Lane.svelte` | **REMOVIDO** (morto após tirar as raias) |
| `web/src/tests/omnibox-*.test.js` | 3 suites (parser, sugestões, componente) |

## Decisões de design (importantes para quem continuar)

1. **Hand-rolled, não Bits UI Combobox.** O Combobox do Bits UI (já no projeto) SUPORTA
   groups+headers e é Svelte 5, mas seu modelo de `value`/`inputValue` (selecionar
   autopreenche o campo) conflita com o requisito de **texto literal multi-token** +
   Enter-para-buscar. Seguiu-se o padrão do `ui/Combobox.svelte`, hand-rolled pelo mesmo
   motivo. Zero deps novas; `cmdk-sv` avaliado e dispensado.
2. **Sem eventos novos na engine.** Seleção reusa `ADICIONAR_LOJA`, `ADICIONAR_CATEGORIA`,
   `MUDAR_MARKETPLACES`, `CARREGAR_SALVA`, `DIGITAR`. A engine (ADR-0027) ficou intacta.
3. **Keyword-only vai para `DIGITAR`.** Ao digitar, só os tokens keyword viram
   `engine.send(DIGITAR)` (a engine debounce 400ms). `@loja`/`#cat`/`!mkt` não poluem a keyword.
4. **Seleção remove o token ativo** e mostra card abaixo (não reinsere `@nome`). Evita o
   bug de nomes com espaço (`Glory of Seoul`) quebrarem a tokenização por espaço.

## ⚠️ Conhecido / a melhorar

- **O workflow de LOJAS precisa de refactor e correção** (feedback do dono do produto).
  É o ponto fraco atual: resolução/adição de loja, nomes com espaço vs. tokens sem espaço,
  derivação de `lojasMonitoradas` a partir de buscas salvas, e o escopo `shopIds` × marketplace.
  A engine e o resto do omnibox estão "getting there" e a **FSM está boa**. Ver **T-0056**.
- **E2E locais/prod obsoletos.** `web/tests/local/*.spec.js` e `web/tests/prod/descobrir.spec.js`
  miram as raias (removidas). Precisam ser reescritos para o omnibox. Ver **T-0054** (reaberta).

## Como rodar / verificar

```bash
cd web && bunx vitest run && bun run check && bun run build   # gates rápidos
mise run test:e2e-novos                                       # pipeline backend (não toca UI)
```

**Playtest local contra a API de produção** (sem subir backend):
```bash
# proxy server-side → sem CORS; login via Google no navegador
DEV_API_PROXY=https://garimpei.app.br VITE_API_BASE= bun run dev   # (dentro de web/)
```
`vite.config.js` já tem o `server.proxy` configurável via `DEV_API_PROXY` (default `:8080`).
Obs.: `dotnet` e `semgrep` podem não estar no PATH de shells não-interativos — o pre-push
hook (`mise run prepush`) precisa deles; rode o `git push` com o PATH do login shell.

## Próximos passos sugeridos

1. **Refactor do subsistema de lojas** (T-0056) — prioridade.
2. **Reescrever E2E** para o omnibox (T-0054).
3. Validar a UX do omnibox em produção com os 2 usuários; ajustar sugestões/atalhos.
4. Decidir se mantém o `DEV_API_PROXY` no vite.config (é benigno; default seguro).

## Referências

- Spec: `.kiro/specs/omnibox-input/` (requirements.md, design.md, tasks.md)
- ADR-0027 (BuscaEngine + regras externas) — seção "v4: omnibox"
- ADR-0030 (BuscaContract unificado) — omnibox compõe o mesmo contrato via eventos existentes
- Tasks: T-0055 (feito), T-0056 (lojas — pendente), T-0054 (E2E — reaberta)
