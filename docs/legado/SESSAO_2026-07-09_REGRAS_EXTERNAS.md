# Sessão 09/Julho 2026 — Regras de busca como JSON externo + E2E desacoplados

## Objetivo

Externalizar as regras de decisão da BuscaEngine para um JSON declarativo
versionado no git (`rules/busca-rules.json`), de forma que:
1. Testes E2E possam validar o comportamento do frontend contra a fonte de verdade
2. CI valide schema + completude + consistência em tempo de build
3. Qualquer linguagem/serviço futuro possa ler o mesmo arquivo

## Contexto e decisões

### O que foi descartado: rules-service como sidecar

O spec original (`.kiro/specs/rules-service/`) propunha um sidecar Go com
gorules/zen-go (binding Rust via CGO), gRPC na porta 50055, JDM no formato
GoRules, proxy C# e cache no frontend. Após análise:

1. **zen-go** — última release há 4 meses, baixa rotatividade
2. **Complexidade desproporcional** — +1 container, +1 hop de rede, proto + stubs +
   proxy + cache + fallback + Docker + CI para 4 regras de intent e 2 guards
3. **O frontend já resolvia** — `busca-config.js` tinha `INTENT_TABLE`, `GUARDS`,
   `NORMALIZE` como dados declarativos, funções puras para avaliar

### O que foi implementado: regras como dados puros testáveis

A solução mais lightweight que atende ao requisito:
- JSON externo (`rules/busca-rules.json`) — dados puros, sem engine opaco
- Frontend importa em build-time (zero latência, zero hop de rede)
- E2E importam o mesmo JSON e validam comportamento da UI
- CI drift check valida schema + propriedades invariantes

**Princípio:** para 4 intents, 2 guards, 2 normalizações que mudam por PR —
a solução correta é dados versionados + testes, não um rules engine.

## Entregas

### 1. `rules/busca-rules.json` — fonte de verdade

```json
{
  "version": "1.0.0",
  "intent": [4 rows: keyword×shop → result + sources],
  "guards": { "temContextoBusca": {...}, "podeSalvar": {...} },
  "normalize": { "comissao": {...}, "vendas": {...} },
  "defaults": { "comissaoMin": 0.07, ... },
  "transicoes": { "DIGITAR": {...}, "ADICIONAR_LOJA": {...}, ... }
}
```

Toda regra que antes era hardcoded em `busca-config.js` agora vive aqui.

### 2. `rules/busca-rules.schema.json` — JSON Schema

Valida tipos, enumerações de intent, completude da tabela (4 combinações boolean),
e estrutura de guards/normalize/defaults.

### 3. `web/src/lib/busca-config.js` — refatorado

Antes: dados hardcoded (DEFAULTS, NORMALIZE, GUARDS, TRANSICOES, INTENT_TABLE).
Agora: importa de `../../../rules/busca-rules.json` e re-exporta no formato da engine.
Funções puras (`normalizarComissao`, `checarGuard`, `intentBusca`) permanecem — operam
sobre os dados importados.

### 4. `.mise/tasks/check/rules-schema` — drift check

Script bash que valida:
- JSON válido (python3 json.load)
- Schema (jsonschema, se disponível)
- Intent table cobre 4 combinações (TT, TF, FT, FF)
- Intents válidos (enum check)
- Guards consistentes (podeSalvar ⊆ temContextoBusca)

### 5. `web/tests/local/busca-rules.spec.js` — E2E contra regras

5 testes Playwright que importam `rules/busca-rules.json` via `fs.readFileSync`
e validam o comportamento da UI:

| # | Cenário | O que prova |
|---|---------|-------------|
| 1 | Busca "serum" → adicionar Le Botanic | Intent muda para `keyword_na_loja`, resultados escopados |
| 2 | Filtro comissão | Nunca mostra float cru — sempre "7%", "10%" |
| 3 | Salvar busca → chip → restaurar | Chip com label correto, click restaura contexto |
| 4 | Agendar (cron) | POST inclui campo cron, badge ⏱ no chip |
| 5 | Toggle novos + loja | Fonte "novos" ativa por default (conforme `rules.defaults`) |

### 6. Documentação atualizada

| Doc | O que foi adicionado |
|-----|---------------------|
| `docs/02-arquitetura.md` | Seção "Regras de negócio externalizadas" + `check:rules-schema` na tabela de drift |
| `docs/06-qualidade-e-testes.md` | E2E local na tabela de cobertura + rules-schema nas métricas |
| `docs/08-fluxos-sequencia.md` | Intent resolution no fluxo Descobrir |
| `docs/componentes.md` | busca-config.js → importa de rules/ externo |
| `contracts/registry.yaml` | `rules` como componente estático com path e schema |

## Arquivos criados/modificados

```
rules/
  busca-rules.json              ← NOVO: fonte de verdade
  busca-rules.schema.json       ← NOVO: schema para validação
.mise/tasks/check/
  rules-schema                  ← NOVO: drift check
web/src/lib/
  busca-config.js               ← MODIFICADO: importa do JSON externo
web/tests/local/
  busca-rules.spec.js           ← NOVO: 5 E2E contra regras
contracts/
  registry.yaml                 ← MODIFICADO: +rules como componente
docs/
  02-arquitetura.md             ← MODIFICADO: seção rules
  06-qualidade-e-testes.md      ← MODIFICADO: métricas
  08-fluxos-sequencia.md        ← MODIFICADO: intent no fluxo
  componentes.md                ← MODIFICADO: busca-config
```

## Verificação

```bash
# Frontend (tudo passa):
cd web && npm run check && npm run lint:js && npm run format:check && npx vitest run && npm run build
# → 208 unit tests, 0 errors, 0 warnings, build ok

# Drift check rules:
.mise/tasks/check/rules-schema
# → ✅ Rules schema válido e consistente!

# Go (sem regressão):
go build ./... && go test ./...
# → ok

# E2E local (requer preview rodando):
cd web && npm run test:e2e:local
# → 3 specs originais + 5 novos
```

## Commits

```
f9dbf6d feat: regras de busca como JSON externo + E2E validando contra rules
```

## Decisões a respeitar (para sessões futuras)

1. **Sem rules engine externo.** O JSON é a spec; o código (funções puras) é o evaluator.
   Só justifica engine se: multi-tenant com regras por cliente, ou operadores não-dev
   editando em produção.

2. **Frontend consome em build-time.** O `import` resolve via Vite; zero fetch em runtime.
   Se no futuro precisar de regras dinâmicas, adicionar endpoint GET que retorna o JSON.

3. **E2E testam contra o JSON, não contra hardcoded.** Se alguém mudar o JSON sem
   atualizar o frontend, os E2E falham. Essa é a garantia.

4. **Drift check no CI.** `mise run check:rules-schema` deve estar no `mise run checks`
   ou no `mise run prepush`.

5. **spec rules-service descartada operacionalmente** mas mantida em
   `.kiro/specs/rules-service/` como referência caso a complexidade justifique no futuro.

## Estado da branch

Branch: `claude/monitored-stores-refactor-xazzf9`
Base: `origin/main`

Todos os commits desde a sessão anterior (bugs + config declarativa + E2E harness +
spec) continuam válidos. Este commit adiciona a externalização das regras.

Pronto para push (aguardando confirmação do usuário).
