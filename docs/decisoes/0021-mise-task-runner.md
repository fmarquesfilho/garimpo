# ADR-0021: Adoção do mise como task runner e version manager

## Status

Aceito (2026-07-02)

## Contexto

O projeto acumulou 9 scripts bash em `/scripts` (1400 linhas) mais um Makefile
(80 linhas) para orquestrar builds, testes, lints, checks e docs. Problemas:

1. **Sem discovery**: dev precisava ler cada script para saber o que existia
2. **Sem dependências**: pre-push hook reimplementava orquestração manualmente
3. **Bash frágil**: heredocs bugavam, quoting inconsistente, $ROOT vs $PWD
4. **Sem versões**: cada dev instalava Go/Node/Python na versão que quisesse
5. **Duplicação**: CI e pre-push repetiam os mesmos passos de formas diferentes
6. **Sem paralelismo declarativo**: pre-push hook fazia isso com `&` e `wait`

## Decisão

Adotar **mise** (mise-en-place) como task runner + version manager único:

- `mise.toml` na raiz declara runtimes (go, node, python, uv) e tasks TOML
- `.mise/tasks/` contém file tasks (scripts bash com `#MISE` headers)
- CI usa `jdx/mise-action@v2` e chama `mise run` para cada step
- Pre-push hook delega para `mise run prepush`
- Makefile removido

### Por que mise (e não Taskfile, Just, ou Earthly)

| Critério | mise | Taskfile | Just |
|----------|------|----------|------|
| Gerencia runtimes | ✅ go, node, python, uv | ❌ | ❌ |
| Formato | TOML (familiar Go/Rust) | YAML | DSL próprio |
| File tasks | ✅ scripts em `.mise/tasks/` | ❌ | ❌ |
| Paralelismo | ✅ depends nativo | ✅ | ❌ |
| Discovery | `mise tasks` com descrições | `task --list` | `just --list` |
| Monorepo | ✅ target paths | includes | ❌ |
| CI integration | `jdx/mise-action` oficial | manual | manual |
| Onboarding | `mise install` → tudo pronto | precisa instalar deps separado | idem |

O fator decisivo: Garimpo é **polyglot** (Go + C# + Node + Python). Mise resolve
versões + tasks num só lugar. Taskfile/Just só resolvem tasks.

## Estrutura resultante

```
mise.toml                           ← runtimes + tasks TOML (build, test, lint, docs)
.mise/tasks/
├── check/
│   ├── api-contract                ← ex-scripts/check-api-contract.sh
│   ├── config-consistency
│   ├── data-ownership
│   ├── file-size
│   ├── schema-sync
│   ├── service-contracts
│   ├── stale-refs
│   └── api-spec-sync              ← requer server rodando
├── docs/
│   ├── env                         ← ex-scripts/gen-env-doc.sh
│   └── sync                        ← ex-scripts/sync-docs-to-site.sh
└── backlog/
    └── create                      ← ex-scripts/task.sh
```

## Vantagens obtidas

1. **`mise tasks`** — lista 31 tasks com descrições (zero leitura de código)
2. **`mise install`** — instala Go 1.26, Node 24, Python 3.13, uv em um comando
3. **`mise run checks`** — roda 6 checks em paralelo automaticamente
4. **`mise run ci`** — simula CI completo local (build + test + lint + checks + docs)
5. **CI = Local** — mesmas tasks, mesma ferramenta
6. **Onboarding**: clone + `mise install` + `mise run ci` = pronto para contribuir
7. **File tasks** mantêm lógica complexa em bash com syntax highlighting
8. **-517 linhas** removidas (Makefile + boilerplate dos scripts)

## Consequências

### Positivas
- Dev experience significativamente melhor
- CI mais limpa (sem `actions/setup-go` + `actions/setup-node` + `actions/setup-python` separados)
- Nenhum "works on my machine" — versões pinadas no `mise.toml`
- Scripts continuam sendo bash legível (file tasks), editáveis com qualquer editor

### Negativas
- Dependência nova: `mise` precisa estar instalado (`brew install mise`)
- Primeiro `mise install` demora 1-2min (download dos runtimes)
- .NET 10 não é gerenciado pelo mise (usa `actions/setup-dotnet` no CI)
- Curva de aprendizado mínima para quem nunca usou

### Mitigações
- Pre-push hook tem fallback: se mise não instalado, roda file tasks diretamente
- README atualizado com instruções de setup
- .NET mantido via setup-dotnet no CI (mise plugin experimental para .NET)
