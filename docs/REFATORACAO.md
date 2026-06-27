# Refatoração — Garimpei

Sessão de 27/06/2026. Objetivo: reduzir tamanho dos arquivos, eliminar duplicação, melhorar manutenibilidade. Regra implantada: **máximo 400 linhas por arquivo de produção**.

---

## Resumo

Todos os arquivos de código de produção agora têm no máximo 400 linhas. Uma verificação automática (`scripts/check-file-size.sh`) foi adicionada à CI para impedir regressões futuras.

---

## Frontend SvelteKit (Svelte 5)

### Páginas refatoradas

| Página | Antes | Depois | Redução |
|--------|:-----:|:------:|:-------:|
| `/lojas` | 712 | 309 | −57% |
| `/publicar` | 497 | 379 | −24% |
| `/oportunidades` | 415 | 231 | −44% |
| `/publicacoes` | 359 | 317 | −12% |
| `+layout.svelte` | 409 | 119 | −71% |

### Componentes extraídos

| Componente | Linhas | Origem |
|------------|:------:|--------|
| `FormAdicionarLoja.svelte` | 115 | /lojas — form de add loja |
| `PainelAlertas.svelte` | 164 | /lojas — config alertas Telegram |
| `ListaProdutosLoja.svelte` | 74 | /lojas — grid de produtos |
| `CardOportunidade.svelte` | 108 | /oportunidades — card queda/alta/novo |
| `ResolverLink.svelte` | 101 | /publicar — input com resolução de link |
| `PreviewPublicacao.svelte` | 39 | /publicar — preview Telegram |
| `NavDrawer.svelte` | 116 | layout — menu lateral |
| `LandingHero.svelte` | 61 | layout — landing page |

### Componentes reutilizados (já existiam)

- `PeriodSelector` — substituiu inline duplicado em oportunidades e publicacoes
- `TabBar` — substituiu hand-rolled em lojas e publicacoes
- `Loading`, `EmptyState`, `ErrorMessage`, `PageHeader` — adotados nas 4 páginas

### Padrão seguido

- Svelte 5: `$props()`, `$state`, `$derived`, `$effect`, `$bindable`
- Sem dependências novas — tudo com Svelte nativo
- Cada página é coordenador: estado + fetch + delegação para componentes

---

## Backend Go

### `internal/store/bigquery_store.go` (1121 → 6 arquivos ≤ 400 linhas cada)

| Arquivo | Linhas | Responsabilidade |
|---------|:------:|-----------------|
| `bigquery_store.go` | ~300 | Struct, constructor, Registrar, Snapshot, Buscas |
| `bigquery_schema.go` | ~160 | EnsureSchema + evolução de schema |
| `bigquery_queries.go` | ~160 | HistoricoColetas, Estatisticas, Conversoes |
| `bigquery_publicacoes.go` | ~100 | CRUD de publicações |
| `bigquery_novidades.go` | ~200 | Novidades + EvolucaoLojas |
| `bigquery_destinos.go` | ~96 | BQDestinoStore |
| `bigquery_templates.go` | ~95 | BQTemplateStore |

### `internal/httpapi/lojas.go` (441 → 2 arquivos)

| Arquivo | Linhas | Responsabilidade |
|---------|:------:|-----------------|
| `lojas.go` | 241 | HTTP handlers (CRUD de lojas) |
| `shopee_resolver.go` | 209 | Regex + resolução URL/slug/shortlink |

---

## Guarda de tamanho na CI

### Script: `scripts/check-file-size.sh`

```bash
./scripts/check-file-size.sh 400
```

- Limites: **400 linhas** para código de produção, **900 linhas** para testes
- Verifica `.go`, `.svelte`, `.js`, `.ts`
- Integrado na CI em ambos os jobs (`test-go` e `test-web`)
- Bloqueia deploy se qualquer arquivo de produção exceder o limite

### Onde roda na CI

```yaml
# Job test-go (Go):
- name: Limite de tamanho de arquivo (max 400 linhas — bloqueia deploy)
  run: ./scripts/check-file-size.sh 400

# Job test-web (frontend):
- name: Limite de tamanho de arquivo (max 400 linhas)
  run: ../scripts/check-file-size.sh 400
```

### Para ignorar o limite (exceções)

Arquivos `_test.go` e `*.spec.*` permitem até 900 linhas (warning, não bloqueiam). Se um arquivo de produção legitimamente precisa exceder 400 linhas, a abordagem correta é **extrair** — não aumentar o limite.

---

## Validação

Toda a refatoração passou:

```
go build ./...        ✔
go test ./...         ✔ (todos passando)
npm run build         ✔ done
npm run lint:js       ✔ 0 errors, 0 warnings
npm run lint:css      ✔ limpo
npx vitest run        ✔ 34 testes passando
check-file-size.sh    ✔ todos dentro do limite
```

---

## Princípios aplicados

1. **Refator puro** — nenhuma feature nova, nenhum comportamento alterado
2. **Cada arquivo tem uma responsabilidade coesa** — não misturar HTTP handlers com lógica de negócio
3. **Páginas como coordenadoras** — delegam UI para componentes, mantêm apenas estado e fetch
4. **Reutilização sobre duplicação** — PeriodSelector, TabBar, Loading usados em vez de copiar CSS inline
5. **Prevenção > correção** — CI bloqueia regressões futuras
