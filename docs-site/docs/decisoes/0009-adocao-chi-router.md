# ADR 0009 — Adoção do Chi como router HTTP

## Contexto

O projeto usava `net/http.ServeMux` (stdlib Go 1.22+) para routing. Com ~40 rotas,
o mux funcionava, mas apresentava limitações:

- Subpath matching frágil (bug: `/docs/subpagina` retornava 404)
- Auth verificada manualmente dentro de cada handler (~25 repetições)
- Sem Method Not Allowed automático (POST em rota GET dava 404 em vez de 405)
- Middleware por grupo de rotas não era possível sem wrappers manuais

## Decisão

Adotar **go-chi/chi v5.3.0** como router. Razões:

- Mesma assinatura de handler (`http.HandlerFunc`) — migração sem reescrita de lógica
- Zero dependências transitivas (~3k LOC)
- Suporte nativo a Groups, Mount, e middleware por sub-router
- Amplamente adotado no ecossistema Go (usado por Cloudflare, Heroku, etc.)
- Performance idêntica ao stdlib (trie-based, O(1) routing)

## Impacto medido

### Código

| Métrica | Antes | Depois | Variação |
|---|---|---|---|
| Linhas em handlers (auth boilerplate) | ~80 linhas repetidas | 0 (middleware) | -80 |
| Arquivos alterados na migração | — | 22 | — |
| Handlers reescritos | 0 | 0 | sem mudança |
| Testes quebrados | 0 | 0 | adaptados, não reescritos |
| Dependências adicionadas | 0 | 1 (chi, zero transitivas) | +1 |
| Net diff de linhas | — | -72 (206 adicionadas, 278 removidas) | redução |

### Qualidade

| Aspecto | Antes | Depois |
|---|---|---|
| Auth esquecida numa rota nova | Possível (cada handler fazia) | Impossível (middleware do grupo) |
| Routing de subpaths (/*) | Bug (404) | Funciona (Mount) |
| Method Not Allowed | Silencioso (retornava 404) | Explícito (405) |
| Logging duplicado em handler | Possível | Teste-guardrail impede |
| Auditoria de segurança | Revisar 40 handlers | Revisar 3 middlewares |

### Performance

Imensurável. Ambos fazem routing em microsegundos. O gargalo do app é I/O
(BigQuery ~200ms, API Shopee ~500ms). O router é irrelevante para latência.

### Testes adicionados

| Teste | O que garante |
|---|---|
| `TestDocsFileServer` | Subpaths de docs funcionam (index, subpage, asset, 404) |
| `TestRouterMethodNotAllowed` | POST em rota GET-only retorna 405 |
| `TestRouterCORS` | OPTIONS retorna headers CORS corretamente |
| `TestNoManualAuthInHandlers` | Nenhum handler faz auth manual (guardrail) |
| `TestNoManualLoggingInHandlers` | Nenhum handler loga request manual (guardrail) |

## Consequências

- Toda rota nova dentro do grupo autenticado herda `requireAuth` automaticamente.
- Para adicionar auth a uma rota pública, basta movê-la para o grupo correto.
- Middlewares vivem em `middleware_auth.go` e `middleware_log.go` — fácil de encontrar.
- O `arch-go.yml` continua validando que `internal/httpapi` não importa BigQuery diretamente.
- Dependabot monitora chi por novas versões (PR semanal se houver update).
