# ADR-0025: Deploy conservador — errar para o lado de deployar

## Status

Aceito (2026-07-07)

## Contexto

Ao implementar path filtering no CI (dorny/paths-filter) para evitar builds
desnecessários, introduzimos um bug: commits que tocavam código de backend
ficaram sem deploy porque o commit que adicionou os filtros não matchava os
paths de backend (`src/`, `services/`, etc.). O resultado foi 14 commits
deployados no frontend mas não no backend — produção ficou desatualizada.

### O problema com "só deploya se detectar mudança"

A lógica otimista ("pula deploy se nada mudou") é frágil em monorepos porque:

1. Commits que mudam `.github/workflows/` não matcham paths de `src/` — mas
   podem afetar como o deploy funciona
2. Merge commits podem conter mudanças de múltiplos PRs que individualmente
   não matcham
3. Ferramentas de path detection dependem de `fetch-depth` e podem falhar
   silenciosamente
4. Um deploy skipped que deveria ter rodado causa bug em produção (grave).
   Um deploy que roda desnecessariamente custa ~4min (trivial — idempotente).

### Incidente

Em 2026-07-07, 14 commits de backend (incluindo novas features: publicações
agendadas, GenerateAffiliateLink, correção do Scheduler) ficaram sem deploy
por ~2 horas até ser detectado manualmente.

## Decisão

### Deploy conservador: errar para o lado de deployar

O deploy backend roda **sempre**, exceto quando é **comprovadamente seguro**
pular — ou seja, quando o path filter confirma que **apenas `web/`** mudou.

```yaml
if: |
  always() &&
  github.ref == 'refs/heads/main' && github.event_name == 'push' &&
  (needs.changes.outputs.backend == 'true' || needs.changes.outputs.web != 'true') &&
  !contains(needs.*.result, 'failure')
```

**Tradução:** "Deploya se backend mudou OU se não foi apenas o frontend que mudou."

### Path filtering para checks (não deploys)

Os jobs de **validação** (Go, C#, Python, Frontend, Proto, Contracts) continuam
usando path filtering agressivo — é seguro pular lint/test se o código não mudou.
Isso economiza ~3min por push irrelevante.

Os jobs de **deploy** usam a lógica conservadora — na dúvida, deploya.

### Validação YAML obrigatória

Qualquer edição em `.github/workflows/*.yml` deve ser validada com:
```bash
yq '.' .github/workflows/ci.yml > /dev/null
```
Isso previne erros de syntax como `jobs:` duplicado.

## Consequências

### Positivas

- Produção nunca fica desatualizada por falha de detecção de paths
- Deploy é idempotente — rodar sem necessidade não causa dano
- Checks ainda são otimizados por path (economia de CI minutes)
- Regra simples de entender: "na dúvida, deploya"

### Negativas

- Deploy backend roda em pushes que só tocam `.github/`, `contracts/`,
  ou arquivos na raiz — ~4min "desperdiçados" em casos edge
- Não é a abordagem mais eficiente em CI minutes

### Alternativas rejeitadas

| Alternativa | Motivo da rejeição |
|-------------|-------------------|
| Path filtering agressivo (só deploya se backend mudou) | Causou o incidente — 14 commits sem deploy |
| Nenhum path filtering (sempre roda tudo) | Desperdício de ~7min em pushes de docs/config |
| Deploy manual (mise run deploy) | Erro humano — esquece de deployar |
| Deploy na startup do Cloud Run (auto-migrations) | Aumenta cold start, não resolve imagens desatualizadas |

## Referências

- [dorny/paths-filter](https://github.com/dorny/paths-filter) — detecção de paths no nível de job
- [GitHub Actions path filtering docs](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onpushpull_requestpull_request_targetpathspaths-ignore)
- Steering rule: `.kiro/steering/ci.md`
