# Análise Estática — Garimpei

Ferramentas configuradas em 2026-06-27 para qualidade, segurança e manutenibilidade.

## Ferramentas Instaladas

| Ferramenta | Propósito | Como rodar |
|-----------|-----------|-----------|
| **golangci-lint** | Meta-linter Go (50+ checagens) | `golangci-lint run ./...` |
| **govulncheck** | Vulnerabilidades em dependências Go | `govulncheck ./...` |
| **arch-go** | Validação de regras de arquitetura | `arch-go` |
| **ESLint** | Linting JS/Svelte (já no CI) | `cd web && npm run lint:js` |
| **Stylelint** | CSS/design tokens (já no CI) | `cd web && npm run lint:css` |
| **knip** | Código/deps mortas no frontend | `cd web && npx knip` |
| **svelte-check** | Tipos e erros em componentes | `cd web && npx svelte-check` |

## Resultados da Primeira Execução (baseline)

### golangci-lint

```
Linters com issues:
  gocritic   84 (performance: hugeParam, rangeValCopy)
  gofmt      27 (formatação)
  misspell   22 (ortografia em PT-BR confundida com EN)
  errcheck    4 (erros não tratados reais)
  prealloc    3 (slices sem pre-alocação)
  gosec       2 (math/rand, http sem timeout)
  unused      1 (função resolveShopSlug)
```

**Prioridades de correção:**
1. `errcheck` (4) — erros silenciados que podem causar bugs
2. `gosec` (2) — http.ListenAndServe sem timeout, math/rand fraco
3. `unused` (1) — código morto
4. `gocritic/hugeParam` — performance (structs grandes copiadas por valor)
5. `gofmt` — rodar `gofmt -w .` uma vez resolve todos
6. `misspell` — falsos positivos com português, configurar exclusões

### govulncheck

```
12 vulnerabilidades na stdlib Go (go1.26)
Corrigidas em: go1.26.4
Ação: atualizar Go para 1.26.4 no go.mod e Dockerfile
```

**Vulnerabilidades críticas:**
- GO-2026-4918: HTTP/2 infinite loop (DoS) — fixado em go1.26.3
- GO-2026-4870: TLS KeyUpdate DoS — fixado em go1.26.2
- GO-2026-5039: net/textproto input não escapado — fixado em go1.26.4

### arch-go

```
Compliance: 100% ✅ (todas as 6 regras passam)
Coverage:    33% (6 de 18 packages têm regras)
```

**Regras configuradas:**
- `domain` → não importa nada externo ✅
- `source` → não importa httpapi/store/publish ✅
- `engine` → não importa httpapi/store ✅
- `strategy` → não importa httpapi/store/source ✅
- `store` → não importa httpapi/source/engine ✅
- `tenant` → não importa httpapi/source/store ✅

### Frontend (ESLint + Stylelint + knip)

```
ESLint:    0 errors, 16 warnings (no-unused-vars em código existente)
Stylelint: 0 errors
knip:      0 issues
```

## Ações Recomendadas (por prioridade)

### P0 — Segurança (fazer agora)
- [ ] Atualizar Go de 1.26 para 1.26.4 (`go.mod` + Dockerfile)
- [ ] Corrigir `http.ListenAndServe` → usar `http.Server{ReadTimeout, WriteTimeout}`

### P1 — Bugs potenciais (próxima sessão)
- [ ] Corrigir 4 `errcheck` (json.Decode sem tratar erro pode causar panic)
- [ ] Remover `resolveShopSlug` não usada

### P2 — Performance (quando tiver tempo)
- [ ] Converter `hugeParam` para ponteiros nos hot paths (Params, Busca, Scored)
- [ ] Pre-alocar slices identificados pelo `prealloc`

### P3 — Estilo (batch)
- [ ] `gofmt -w .` para resolver os 27 issues de formatação
- [ ] Configurar `misspell` para ignorar termos em português

## Integração com CI

As ferramentas podem ser adicionadas ao workflow `.github/workflows/deploy-gcp.yml`:

```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest

- name: govulncheck
  run: go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...

- name: arch-go
  run: go install github.com/arch-go/arch-go@latest && arch-go
```

**Recomendação:** adicionar ao CI de forma não-bloqueante inicialmente (warnings)
e tornar bloqueante após corrigir os issues existentes.
