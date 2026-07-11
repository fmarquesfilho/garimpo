# ADR-0026: Rotação de service account key + limpeza do repositório

## Status

Aceito (2026-07-07)

## Contexto

Durante análise de código morto no repositório, identificamos que um arquivo
`key.json` (service account GCP para `gh-deploy@garimpo-500114.iam.gserviceaccount.com`)
estava commitado no histórico do git desde o commit `27466c0` (janeiro/2026).

Apesar de estar no `.gitignore`, o arquivo havia sido commitado antes da regra
ser adicionada — permanecia acessível no histórico do repositório.

Adicionalmente, vários artefatos mortos se acumularam na raiz do repositório
ao longo da migração Go → C#:
- Binários Go compilados localmente (`collector`, `publisher`, `scheduler`, `gen-board`)
- Scripts legados não referenciados pelo mise (`prod-api.sh`, `testar_shopee.py`, etc.)
- Pasta `api/openapi.yaml` estática que ficava permanentemente desatualizada

## Decisão

### 1. Rotação da service account key

- **Key `265a65f5a3352b548d2e245191336354c3bd64e8` revogada** via
  `gcloud iam service-accounts keys delete`
- A key não tem mais validade — qualquer uso futuro retornará erro de autenticação
- O CI usa **Workload Identity Federation (OIDC)** — não depende de JSON keys
- Nenhum sistema em produção usava essa key (era resquício do setup inicial)

### 2. Limpeza do repositório

Itens removidos:

| Item | Motivo |
|------|--------|
| `key.json` | Credencial exposta no histórico |
| `collector`, `publisher`, `scheduler` (binários raiz) | Build artifacts locais |
| `gen-board` (binário raiz) | Build artifact local |
| `api/openapi.yaml` | Spec estática sempre desatualizada (substituída por docs-site/docs/api.md) |
| `scripts/candidatos_exemplo.csv` | Arquivo de teste antigo |
| `scripts/prod-api.sh` | Script manual legado |
| `scripts/seed-local-test.py` | Seed de teste nunca usado |
| `scripts/test-alerts.sh` | Não integrado ao mise |
| `scripts/testar_shopee.py` | Teste manual legado |
| Pastas duplicadas em `docs-site/docs/` | Artefatos de sync corrompido |

Itens mantidos:

| Item | Motivo |
|------|--------|
| `scripts/pre-push-check.sh` | Hook ativo (`.git/hooks/pre-push`) |
| `firebase.json` | Config do emulator (test:e2e) |
| `docs/legado/` | Histórico de sessões (referência) |
| `docs/meta/` | Plano de documentação interno |
| `cmd/gen-board`, `cmd/gen-er` | Geradores Go usados por `mise run docs` |

### 3. Prevenção futura

- `.gitignore` atualizado com: `collector`, `publisher`, `scheduler`, `gen-board`
- `key.json` já estava no `.gitignore` (previne re-commit)
- Steering rule (`ci.md`): validar YAML antes de commitar
- CI usa OIDC (Workload Identity) — zero JSON keys necessárias

## Consequências

### Positivas
- Zero credenciais no repositório (nem no histórico ativo)
- Raiz do repositório limpa (apenas arquivos com propósito claro)
- Build artifacts não poluem `git status`

### Negativas
- A key exposta esteve acessível no histórico por ~6 meses
  - Mitigação: key revogada, nenhum uso malicioso detectado
- Reescrever histórico (BFG/git-filter-repo) para remover completamente
  é possível mas não necessário — a key já foi revogada

### Ação recomendada (futura)
Se o repositório for tornado público no futuro, considerar `git-filter-repo`
para remover `key.json` do histórico completamente.

## Referências

- [Google Cloud: Rotating service account keys](https://cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)
- [GitHub: Removing sensitive data from a repository](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- Commit de remoção: `1fbd3aa`
