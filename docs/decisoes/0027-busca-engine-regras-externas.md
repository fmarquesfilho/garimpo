# ADR 0027 — BuscaEngine headless + regras externas como JSON declarativo

**Status:** aceite  
**Data:** 2026-07-09  
**Impacto:** Alto — elimina classe inteira de bugs de regressão na página principal

## Por que esta é a decisão mais importante do frontend

A página Garimpar é a experiência central do produto. Antes desta ADR, qualquer
mudança em filtros, fontes, lojas ou salvamento podia quebrar o estado da UI
silenciosamente — sem que nenhum teste detectasse. Os bugs só apareciam quando
um usuário real clicava na combinação certa de controles.

O que esta decisão muda:

1. **O JSON é o contrato.** `rules/busca-rules.json` define formalmente O QUE a
   página deve fazer. Não é documentação — é código executável que os testes leem.

2. **A engine é o runtime.** `BuscaEngine` é uma FSM testável que impede estados
   impossíveis via guards. A view é burra — não tem lógica.

3. **Os testes são a prova.** 294 unit + 24 E2E locais + 15 E2E produção — todos
   validando contra o mesmo JSON. Se alguém mudar uma regra sem atualizar o
   frontend, **os testes quebram no mesmo minuto.**

4. **E2E rodam contra produção.** Não é um ambiente de staging — são testes reais
   contra `garimpei.app.br` com auth Firebase, APIs reais, banco real. Se o deploy
   quebrar algo, os E2E prod detectam em 18 segundos.

Sem esta decisão, o projeto acumularia bugs de regressão a cada feature nova.
Com ela, a rede de segurança é automática e cresce com cada teste adicionado.

---

## Contexto

A página Garimpar (Descobrir) cresceu organicamente e acumulou vários bugs de
estado incoerente:

- Adicionar loja não escopava a busca (resultados de fora da loja apareciam)
- Comissão exibia "7.0000000" em vez de "7%"
- Salvar busca gerava label "(sem keywords)" pois keywords não persistiam
- Clicar em um pill de busca salva não restaurava o contexto completo
- Toggle de fontes (novos/quedas) não refletia dados após adicionar loja
- Categorias eram input livre em vez de seletores da Shopee

A causa raiz era dupla:
1. **Estado espalhado** — múltiplos componentes gerenciavam pedaços do estado
   sem coordenação central (filtros, fontes, lojas, keywords em silos)
2. **Regras hardcoded na view** — lógica de decisão (intent, guards, normalização)
   misturada com o template Svelte, impossível de testar isoladamente

Além disso, uma proposta de rules-service como sidecar gRPC (gorules/zen-go)
foi avaliada e descartada por over-engineering para a complexidade atual.

## Decisão

### 1. BuscaEngine como FSM headless (classe Svelte 5)

Uma máquina de estados única controla toda a página via `send(event)`:

```javascript
class BuscaEngine {
  status = $state('idle');   // idle | searching | results | saving | error
  ctx = $state({...});       // estado completo: keyword, shopIds, fontes, filtros
  
  send(event) { ... }       // DIGITAR, ADICIONAR_LOJA, MUDAR_FILTRO, SALVAR, etc.
}
```

**Propriedades:**
- Guards impedem transições inválidas (não salva sem contexto, não busca sem fonte)
- View é "burra" — renderiza `engine.ctx` e despacha events via `send()`
- Effects injetáveis — testável com `new BuscaEngine(mockEffects())`
- Estado centralizado — impossível que um componente veja dados obsoletos

### 2. Regras como JSON declarativo externo

```
rules/
  busca-rules.json         ← fonte de verdade (dados puros)
  busca-rules.schema.json  ← schema para validação CI
```

O JSON contém:
- **Intent table** (4 rows): keyword × shop → resultado + fontes necessárias
- **Guards**: condições para contexto válido e permissão de salvar
- **Normalização**: comissão (divide por 100 se >1), vendas (floor, min 0)
- **Defaults**: valores iniciais dos filtros e fontes
- **Transições**: comportamento por event (refetch vs client-side, imediato vs debounce)

O frontend importa em build-time (`import rules from '../../../rules/busca-rules.json'`).
Zero latência, zero hop de rede, zero engine externo.

### 3. Sem rules-service (sidecar descartado)

A proposta original previa:
- Sidecar Go com gorules/zen-go (binding Rust via CGO)
- JDM no formato GoRules (Decision Tables + Expression Nodes)
- gRPC na porta 50055, proto + stubs, proxy C# transparente
- Frontend com cache 30s + fallback local

**Descartado porque:**
- gorules/zen-go tem baixa rotatividade (última release há 4 meses)
- Complexidade: +1 container, +1 proto, +1 endpoint, +1 Docker image no CI
- As regras são triviais: 4 intents, 2 guards, 2 normalizações
- Mudam por PR (não por operadores em runtime)
- O frontend já precisa das regras localmente para zero-latência

**Quando seria justificável:**
- Multi-tenant com regras diferentes por cliente
- Operadores não-dev editando regras em produção via UI
- Centenas de regras com lógica complexa (scoring, eligibility dinâmica)

### 4. Testes validam contra o JSON externo

```
Unit tests (Vitest)        → importam rules/busca-rules.json
E2E locais (Playwright)    → leem rules/busca-rules.json via fs
CI drift check             → valida schema + completude + consistência
```

Se alguém alterar o JSON sem atualizar o frontend, os testes falham.
Se alguém alterar o frontend sem respeitar o JSON, os testes falham.

### 5. Evolução v3: modos de interação, duplicatas, e marketplaces (rules 3.0)

A arquitetura de regras externas provou seu valor quando a v3 da state machine
foi implementada. Três features complexas foram adicionadas **apenas editando o JSON
+ implementando as funções puras correspondentes** — sem tocar na view:

**Modos de interação** (declarados em `rules.modos`):
```json
{
  "explorando": { "transicoes": { "CARREGAR_SALVA": "vinculada" } },
  "vinculada": { "desvinculaEm": ["DIGITAR", "MUDAR_FILTRO", ...], "transicoes": {...} },
  "editando": { "transicoes": { "SALVAR": "explorando", "CANCELAR_EDICAO": "explorando" } }
}
```
A função `proximoModo(modoAtual, tipoEvento)` lê essas regras. A engine chama
`proximoModo` em cada `send()` — zero lógica de modo na view.

**Detecção de busca duplicada** (declarada em `rules.buscaDuplicada`):
```json
{
  "camposIdentidade": ["keyword", "shopIds", "categorias", "marketplacesFiltro"],
  "normalizacao": { "shopIds": "sort", "categorias": "sort_lowercase" },
  "erroAoSalvar": true,
  "feedbackReativo": true
}
```
A função `fingerprint(ctx)` gera hash determinístico. `buscarDuplicada(ctx, salvas)`
compara. O guard `buscaDuplicada` bloqueia o salvar se encontrar match.

**Filtro por marketplace** (declarado em `rules.marketplaces`):
```json
{
  "suportados": ["shopee", "mercado_livre", "amazon"],
  "filtro": { "tipo": "toggle_multi", "min": 0 },
  "icones": { "shopee": "🟠", "mercado_livre": "🔵", "amazon": "🟡" }
}
```
Componente `MarketplaceFilter.svelte` lê de `MARKETPLACES` (re-exportado do JSON).

**O que isso prova:** a decisão de externalizar regras não foi prematura. Ao adicionar
features complexas, as regras novas vão no JSON, as funções puras vão no config, e
os testes validam tudo — sem refatorar a engine ou a view.

| Feature | Linhas no JSON | Linhas no código | Testes adicionados |
|---------|---------------|------------------|--------------------|
| Modos | 30 | 25 (`proximoModo`) | 51 (cenários v3) |
| Duplicatas | 10 | 50 (`fingerprint` + `buscarDuplicada`) | 147 (test dedicado) |
| Marketplaces | 8 | 57 (componente) | — (coberto pelos E2E) |

## Alternativas avaliadas

### 1. Rules engine externo (gorules/zen-go)

- ✅ Visual editor para não-devs
- ✅ Hot-reload via SIGHUP
- ❌ CGO (build complexo), baixa manutenção da lib
- ❌ Over-engineering para 4 regras
- ❌ Adiciona latência (gRPC + proxy)
- ❌ +1 container no Cloud Run

### 2. expr-lang/expr (Go puro)

- ✅ Alta manutenção (commits semanais)
- ✅ Zero CGO
- ❌ Ainda requer sidecar separado
- ❌ Não resolve o problema core (regras são simples)

### 3. Node.js + @gorules/zen-engine

- ✅ NPM com releases semanais
- ✅ Suporte nativo JDM
- ❌ +1 container Node
- ❌ Mesma complexidade operacional do sidecar Go

### 4. Regras inline no código (status quo anterior)

- ✅ Zero infra
- ❌ Regras misturadas com lógica de apresentação
- ❌ Impossível testar regras isoladamente
- ❌ E2E não podem validar contra spec formal

### 5. ✅ JSON declarativo externo (escolhido)

- ✅ Zero infra adicional
- ✅ Testável por qualquer linguagem
- ✅ Versionado por PR
- ✅ Frontend consome em build-time (zero latência)
- ✅ CI valida schema + propriedades
- ✅ Separação dados (JSON) vs avaliação (código)
- ❌ Não serve para regras dinâmicas em runtime

## Consequências

### Positivas

- **9 bugs corrigidos** (6 estado + 3 arquitetura: store desync, lojas ctx, erro engolido)
- **294 unit tests** cobrindo engine, lógica de filtros, regras, modos, e duplicatas
- **24 E2E locais** passando — todos sem skip, sem hacks de mock
- **15 E2E produção** passando contra garimpei.app.br (auth Firebase real)
- **Drift check** no CI impede regressões silenciosas
- **Documentação executável** — o JSON é a spec E o código lê dele
- **Evoluível** — v3 (modos + duplicatas + marketplaces) provou o pattern: +1003 linhas
  de features adicionadas com zero refatoração da engine ou da view

### Correções de arquitetura (sessão final)

| Bug | Causa raiz | Fix |
|-----|-----------|-----|
| Quedas/Novos não carregavam | `executarBusca` usava store externo (`$buscasSalvas`) que não era sincronizado no init | `INICIALIZAR` agora chama `sincronizarStoreExterno()` antes de `executarBusca` |
| Loja adicionada não mostrava novidades | `buscasComLojas` vinha só do store — lojas do `ctx.shopIds` (ainda não salvas) eram ignoradas | `buildBuscasComLojas()` combina store + ctx.shopIds |
| API 500 mostrava "0 resultados" | `carregarCuradoria` engolia todos os erros (`catch { return [] }`) | `isServerError(e)` propaga erros HTTP ≥400; engole apenas erros de rede |

A filosofia: **o código é a correção, não o teste**. Os testes expõem bugs reais;
os fixes são no código da aplicação.

## Como as regras externas validam a interação entre componentes

O `rules/busca-rules.json` não é apenas "config externalizada" — é a **spec
executável** que garante coerência entre os componentes de UI:

```
rules/busca-rules.json (fonte de verdade)
    │
    ├─► web/src/lib/busca-config.js (importa em build-time)
    │       └─► BuscaEngine usa para avaliar intent, guards, normalização
    │
    ├─► web/src/tests/ (unit tests importam rules para validar)
    │       └─► "engine.intent === rules.intent[row].result" 
    │
    ├─► web/tests/local/ (E2E importam rules via fs.readFileSync)
    │       └─► Validam que a UI se comporta conforme as regras
    │
    └─► .mise/tasks/check/rules-schema (CI valida propriedades)
            └─► Completude, consistência, tipos corretos
```

**O que isso garante:**

1. **Intent table controla fontes de dados**: Se `keyword_na_loja` declara
   `sources: ["curadoria", "lojas"]`, a engine deve ativar exatamente essas fontes.
   O E2E valida isso.

2. **Guards bloqueiam estados incoerentes**: Se `podeSalvar.requiresAny = ["keyword", "shopIds"]`,
   o botão Salvar não funciona sem contexto. O unit test prova.

3. **Normalização é determinística**: Se `comissao.divideBy100IfGt1 = true`,
   digitar "7" no filtro resulta em 0.07 — nunca "7.0000000". O E2E valida que
   o select mostra "7%" e não o float cru.

4. **Transições definem refetch vs client-side**: Se `MUDAR_FILTRO.refetch = false`,
   mudar a comissão NÃO chama a API (só refiltra dados locais). O unit test prova
   que `executarBusca` não é chamado novamente.

5. **Defaults são a configuração inicial**: Se `fontes.novos = true`, o toggle
   "🆕 Novos" começa ativo. O E2E valida que novidades aparecem sem clicar.

**Se uma regra mudar no JSON:**
- O CI quebra se o schema ficar inválido
- Os unit tests quebram se o código não refletir a mudança
- Os E2E quebram se a UI não se comportar conforme a nova regra

Isso fecha o ciclo: regra declarada → código obedece → testes provam.

### Negativas

- Regras não podem mudar em runtime (requer build + deploy)
- Sem UI visual para editar regras (PRs no git)
- E2E que dependem do fluxo completo (dados de lojas monitoradas) ainda
  precisam de mocks complexos

### Neutras

- O spec original (`.kiro/specs/rules-service/`) permanece como referência
  futura caso a complexidade justifique

## Lojas de teste (validação real)

URLs reais da Shopee usadas em testes de integração:

| URL | Loja | Shop ID |
|-----|------|---------|
| `https://s.shopee.com.br/70IKp57jnV` | Glory of Seoul | 920292999 |
| `https://s.shopee.com.br/8fQYnxWQqu` | Le Botanic | — |
| `https://s.shopee.com.br/1gGoSgfopD` | — | — |

Resolução testável via `mise run test:e2e:resolve-shop` (requer collector rodando).

## Atualização 2026-07-09 — regras v2 (redesign em raias)

O redesign da página Descobrir em raias (ADR-0004) estendeu as regras para `version: 2.0.0`,
de forma aditiva e ainda validada pelo drift check:

- **Busca só-categorias é um contexto válido.** `guards.temContextoBusca` e
  `guards.podeSalvar` passaram a aceitar `categorias` (além de `keyword`/`shopIds`). Novo
  bloco `contextoCategorias.sources` define os sources globais usados quando há categorias
  mas nem keyword nem loja — avaliado pela função pura `sourcesBusca(ctx)`. A intent table
  de 4 linhas (keyword × loja) permanece intacta.
- **Multi-marketplace.** Novo bloco `marketplaces` (`suportados` + `default`). Categorias e
  lojas carregam seus marketplaces; o payload de salvar leva o filtro `marketplaces`. O
  drift check ganhou a checagem `marketplaces.default ∈ suportados`.
- **Novas transições/eventos.** `ADICIONAR_CATEGORIA`, `REMOVER_CATEGORIA`,
  `MUDAR_MARKETPLACES` nas `transicoes`; a engine ganhou ainda `EDITAR_SALVA` (edit mode) e
  `SALVAR` com update in-place via `editandoId` (reusa o `id` da busca).
- **Effects novos.** `carregarCategorias` agrupa por marketplace; `listarLojasMonitoradas`
  deriva as lojas do dropdown a partir das buscas salvas (sem endpoint novo).

Ver `componentes.md` para a lista de eventos/getters e os componentes de raia.

## Arquivos-chave

| Arquivo | Papel |
|---------|-------|
| `rules/busca-rules.json` | Fonte de verdade — regras declarativas (v3: modos, duplicatas, marketplaces) |
| `rules/busca-rules.schema.json` | JSON Schema (atualizado para v3) |
| `web/src/lib/busca-engine.svelte.js` | FSM headless (classe Svelte 5, v3: modos de interação) |
| `web/src/lib/busca-engine-state.js` | Estado inicial, guards, constantes MODOS |
| `web/src/lib/busca-engine-effects.js` | Effects injetáveis (API calls, buildBuscasComLojas) |
| `web/src/lib/busca-config.js` | Adapter: JSON → formato da engine + funções puras (proximoModo, fingerprint, buscarDuplicada) |
| `web/src/lib/descobrir-logic.js` | Filtragem client-side (funções puras) |
| `web/src/lib/components/BuscaUnificada.svelte` | View burra (v3: raias, MarketplaceFilter) |
| `web/src/lib/components/BuscasSalvasPanel.svelte` | Painel de buscas salvas (v3: modos vinculada/editando) |
| `web/src/lib/components/MarketplaceFilter.svelte` | Filtro multi-marketplace (v3) |
| `.mise/tasks/check/rules-schema` | CI drift check |
| `web/src/tests/busca-engine.test.js` | Unit: engine core + modos v3 |
| `web/src/tests/busca-duplicata.test.js` | Unit: fingerprint + detecção duplicatas |
| `web/src/tests/busca-engine-cenarios.test.js` | Unit: cenários expandidos + v3 |
| `web/src/tests/descobrir.test.js` | Unit: lógica de filtragem |
| `web/tests/local/` | E2E locais (Playwright + mocks) |
| `web/tests/prod/` | E2E produção (Firebase Auth real + APIs reais) |
