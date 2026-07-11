# Desenvolvimento Local

## Pré-requisitos

| Ferramenta | Instalação |
|------------|-----------|
| mise | `brew install mise` |
| Go 1.26+ | `mise install` (gerenciado pelo mise) |
| Node 24+ | `mise install` (gerenciado pelo mise) |
| .NET 10 | [dotnet.microsoft.com](https://dotnet.microsoft.com/download) |
| Docker Desktop | [docker.com](https://www.docker.com/products/docker-desktop) |
| Firebase CLI | `npm install -g firebase-tools` |

Após clonar o repositório:

```bash
mise install          # instala Go, Node, Python, tooling
cd web && npm ci      # instala deps do frontend
```

---

## Comandos do dia-a-dia

### Checks rápidos (roda no pre-push, ~1min)

```bash
mise run prepush
```

Inclui: build (Go + C#), testes unitários (Go + C# + Vitest), lint (golangci-lint + arch-go + ruff + ESLint + Stylelint + Semgrep), checks de integridade (contracts, API drift, schemas, docs drift), svelte-check, file-size limit, Prettier format check.

### Testes E2E (manual, ~2min)

```bash
mise run test:e2e
```

Inicia o Firebase Auth Emulator automaticamente, builda o frontend, e roda os 53 testes Playwright. Requer Docker Desktop rodando (para o emulator). Se o emulator já estiver ativo, reutiliza.

### Testes C# com PostgreSQL

```bash
mise run test:csharp
```

Sobe o PostgreSQL via Docker automaticamente (se não estiver rodando), aplica migrations, e roda os 57 testes xUnit.

### Simular CI completo

```bash
mise run ci
```

Mesmos checks do pre-push + file-size. Não inclui E2E (eles rodam separadamente).

---

## Comandos individuais

| Comando | O que faz |
|---------|----------|
| `mise run build` | Build Go + C# |
| `mise run test` | Testes Go + C# (com PG) + Vitest |
| `mise run test:go` | Apenas testes Go |
| `mise run test:csharp` | Testes C# (auto-start PostgreSQL) |
| `mise run test:web` | Vitest unitários |
| `mise run test:e2e` | Playwright + Firebase Emulator |
| `mise run lint` | Todos os linters |
| `mise run lint:go` | golangci-lint + arch-go |
| `mise run lint:python` | Ruff no analyzer |
| `mise run lint:web` | ESLint + Stylelint + audit + Semgrep |
| `mise run checks` | Checks de integridade (contracts, schemas, docs) |
| `mise run check:web` | svelte-check (tipos, props, a11y) |
| `mise run check:format` | Prettier --check (sem alterar) |
| `mise run check:file-size` | Limite de 400 linhas por arquivo |
| `mise run docs` | Regenera docs (ER, env, board, sync) |
| `cd web && npm run format` | Formata com Prettier (altera arquivos) |

---

## Serviços locais

### PostgreSQL (para testes C#)

```bash
docker compose up -d postgres    # inicia
docker compose down              # para
```

Gerenciado automaticamente pelo `mise run test:csharp`.

### Firebase Auth Emulator (para E2E)

```bash
firebase emulators:start --only auth    # manual
mise run test:e2e                       # auto-start
```

### Stack completa (dev)

```bash
mise run up     # sobe todos (PostgreSQL, BigQuery emulator, API C#, analyzer)
mise run down   # para tudo
mise run logs   # acompanha logs
```

---

## Estrutura dos testes

| Tipo | Comando | Dependências | Tempo |
|------|---------|-------------|-------|
| Go unit | `mise run test:go` | Nenhuma | ~3s |
| C# unit/integration | `mise run test:csharp` | PostgreSQL (Docker) | ~8s |
| Frontend unit | `mise run test:web` | Nenhuma (jsdom) | ~5s |
| Frontend E2E | `mise run test:e2e` | Firebase Emulator | ~90s |

**Nenhum teste depende de serviços em produção.** Tudo roda localmente:
- APIs são mockadas nos E2E via `page.route()`
- PostgreSQL é local (Docker)
- Firebase Auth usa emulator
- BigQuery usa emulator (quando necessário)

---

## Fluxo de push

```
git push
  └─ pre-push hook
       └─ mise run prepush (~1min)
            ├── build (Go + C#)
            ├── test (Go + C# + Vitest)
            ├── lint (todos)
            ├── checks (contracts, docs-drift, schemas)
            ├── check:web (svelte-check)
            ├── check:file-size (400 linhas)
            └── check:format (Prettier)
```

Se quiser rodar E2E antes de push:

```bash
mise run test:e2e && git push
```

---

## Formatação

O projeto usa Prettier com ordenação automática de classes Tailwind.

```bash
cd web && npm run format          # formata todos os arquivos
cd web && npm run format:check    # verifica sem alterar (roda no pre-push)
```

Configuração em `web/.prettierrc`:
- Tabs
- Single quotes
- Print width: 120
- Plugins: prettier-plugin-svelte + prettier-plugin-tailwindcss (class sorting)
