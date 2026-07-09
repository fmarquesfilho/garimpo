# Tasks

## Task 1: Proto — rules/v1/rules.proto

- [ ] Criar `protos/rules/v1/rules.proto` com `RulesService`, `EvaluateRules`, `ReloadRules`
- [ ] Definir messages: `EvaluateRulesRequest`, `EvaluateRulesResponse`, `GuardsResult`, `NormalizedValues`, `ValidationResult`
- [ ] Configurar `go_package` e `csharp_namespace`
- [ ] Gerar stubs Go: `buf generate` → `gen/go/rules/v1/`
- [ ] Gerar stubs C#: `buf generate` → `src/Garimpei.Protos/`
- [ ] Verificar `buf lint` passa

## Task 2: JDM — busca-rules.json

- [ ] Criar diretório `rules/` na raiz do repo
- [ ] Criar `rules/busca-rules.json` com o JDM contendo:
  - Decision Table: Intent (hasKeyword × hasShop → intent)
  - Decision Table: Guards (hasKeyword × hasShop → temContextoBusca, podeSalvar)
  - Expression Node: Normalização (comissaoMin, vendasMin)
- [ ] Validar formato JDM (pode usar editor GoRules online para visualizar)
- [ ] Adicionar `rules/` ao `.gitignore` exceptions se necessário

## Task 3: Go — rules-service server

- [ ] Criar `services/rules/main.go` com bootstrap (gRPC server, health check, signal handler)
- [ ] Criar `services/rules/server.go` com `RulesServer` struct e `EvaluateRules` RPC
- [ ] Implementar `loadRules()` — lê JDM do disco, cria zen engine
- [ ] Implementar `reloadRules()` — atomic swap via `atomic.Pointer`
- [ ] Implementar `setupSignalHandler()` — SIGHUP → reload, SIGINT → shutdown
- [ ] Implementar `buildEvalContext()` — converte map string→string para map com tipos
- [ ] Implementar `mapResultToResponse()` — converte output do zen para proto response
- [ ] Adicionar dependência `gorules/zen-go` ao go.mod
- [ ] Verificar `go build ./services/rules/...` compila

## Task 4: Go — testes

- [ ] Criar `services/rules/server_test.go`
- [ ] Criar `services/rules/testdata/busca-rules.json` (cópia do JDM para testes)
- [ ] Testar 4 combinações de intent (keyword×shop)
- [ ] Testar guard consistency (podeSalvar → temContextoBusca)
- [ ] Testar normalização (comissão >1 → /100, idempotência)
- [ ] Testar reload (load → modify file → reload → assert new behavior)
- [ ] Testar context vazio → error INVALID_ARGUMENT
- [ ] Property-based: determinismo (rapid)
- [ ] Verificar `go test ./services/rules/...`

## Task 5: C# — proxy endpoint

- [ ] Registrar `RulesService.RulesServiceClient` no DI (`Program.cs`, address localhost:50055)
- [ ] Criar `src/Garimpei.Api/Endpoints/RulesEndpoints.cs` com `POST /api/rules/evaluate`
- [ ] Implementar proxy: JSON body → gRPC request → gRPC response → JSON
- [ ] Tratar indisponibilidade (gRPC Unavailable → HTTP 503)
- [ ] Adicionar `RequireAuthorization` + Tag "Rules"
- [ ] Verificar `dotnet build`

## Task 6: Frontend — integração na BuscaEngine

- [ ] Em `busca-engine-effects.js`, adicionar `evaluateRules(ctx)` que chama `POST /api/rules/evaluate`
- [ ] Em `busca-engine.svelte.js`, no `#adicionarLoja` e `#salvar`, chamar `evaluateRules` para obter intent/validation
- [ ] Implementar cache de 30s para evitar chamadas redundantes
- [ ] Implementar fallback: se rules service indisponível, usar guards locais (código existente)
- [ ] Verificar `npm run check` e `npx vitest run`

## Task 7: Docker + Deploy

- [ ] Criar `services/rules/Dockerfile` (multi-stage: Go build → Alpine runtime + rules/)
- [ ] Atualizar `deploy/` Cloud Run config (adicionar container rules na porta 50055)
- [ ] Atualizar CI (`ci.yml`) para build da imagem rules-service
- [ ] Adicionar startup probe gRPC no service.yaml
- [ ] Verificar que `docker build` funciona localmente

## Task 8: Documentação + Drift checks

- [ ] Atualizar `docs/02-arquitetura.md` — adicionar rules-service ao diagrama e tabela
- [ ] Atualizar `docs/08-fluxos-sequencia.md` — novo fluxo "Avaliar Regras"
- [ ] Atualizar `contracts/registry.yaml` (se existir) com novo serviço + fronteiras
- [ ] Atualizar `mise run check:service-contracts` para incluir rules
- [ ] Verificar `mise run check:api-contract` (nova rota /api/rules/evaluate)
- [ ] Rodar `mise run prepush` completo
