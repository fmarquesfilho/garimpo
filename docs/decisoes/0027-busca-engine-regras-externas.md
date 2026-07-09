# ADR 0027 — BuscaEngine headless + regras externas como JSON declarativo

**Status:** aceite  
**Data:** 2026-07-09  

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
- **243 unit tests** cobrindo engine, lógica de filtros, e regras
- **24 E2E locais** passando — todos sem skip, sem hacks de mock
- **Drift check** no CI impede regressões silenciosas
- **Documentação executável** — o JSON é a spec E o código lê dele
- **Evoluível**: se a complexidade crescer, o JSON pode ser consumido por um
  engine externo sem mudar a estrutura

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

## Arquivos-chave

| Arquivo | Papel |
|---------|-------|
| `rules/busca-rules.json` | Fonte de verdade — regras declarativas |
| `rules/busca-rules.schema.json` | JSON Schema |
| `web/src/lib/busca-engine.svelte.js` | FSM headless (classe Svelte 5) |
| `web/src/lib/busca-engine-effects.js` | Effects injetáveis (API calls) |
| `web/src/lib/busca-config.js` | Adapter: JSON → formato da engine |
| `web/src/lib/descobrir-logic.js` | Filtragem client-side (funções puras) |
| `web/src/lib/components/BuscaUnificada.svelte` | View burra |
| `.mise/tasks/check/rules-schema` | CI drift check |
| `web/src/tests/busca-engine.test.js` | Unit: engine core |
| `web/src/tests/busca-engine-cenarios.test.js` | Unit: cenários expandidos |
| `web/src/tests/descobrir.test.js` | Unit: lógica de filtragem |
| `web/tests/local/` | E2E locais (Playwright + mocks) |
