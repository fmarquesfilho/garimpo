# Sessão 06-07/Julho 2026 — Lojas Monitoradas + Agendamento + CI

## Objetivo

Implementar o fluxo completo de monitoramento de lojas (ResolveShop → Scheduler → Coleta → Detecção), integrar tracking de conversões via GenerateAffiliateLink, publicações agendadas via Scheduler, e otimizar o CI com path filtering.

## Entregas (51 commits)

### 1. ResolveShop — Adicionar Lojas Monitoradas

- RPC `ResolveShop` no Collector Go (API pública Shopee v4)
- Suporte a links curtos `s.shopee.com.br/xxx` (HTTP redirect follow)
- Suporte a links diretos `shopee.com.br/{username}` e username puro
- Campo `ShopIds` (bigint[]) na entidade Busca + migration
- POST /api/lojas resolve → persiste → responde com nome da loja

### 2. Tracking de Conversões

- Campo `SourceUrl` na Busca — preserva link original da afiliada
- RPC `GenerateAffiliateLink` — chama Shopee `generateShortLink` com sub_ids
- Integrado nos endpoints de publicação (POST /api/publicar e /api/publicacoes)
- Sub IDs: `[canal, estratégia, data]` → rastreáveis via `conversionReport`

### 3. Buscas Agendadas — Integração com Scheduler

- `SchedulerServiceClient` registrado no DI
- POST /api/lojas → `SetSchedule(enabled=true)` com params do job
- DELETE /api/lojas → `SetSchedule(enabled=false)` para pausar
- Campos `Keywords` (text[]) e `CronExpression` (text) na Busca
- Scheduler `executeJob()` corrigido: roteia `shop_collection` para `FetchShop(shop_id)`

### 4. Publicações Agendadas

- POST /api/publicacoes com `agendada_em` → `SetSchedule` one-shot cron
- Scheduler `executeScheduledPublish()` → callback POST `/internal/publish-scheduled`
- C# resolve destino, gera link afiliada, publica via Publisher gRPC
- Job removido após execução (cleanup)

### 5. Testes E2E (22 cenários)

- `lojas-resolve-shop.spec.js` — 5 cenários (links curtos/diretos/username/erro)
- `buscas-agendadas.spec.js` — 5 cenários (Scheduler + keywords + delete)
- `publicar-agendada.spec.js` — 5 cenários (envio + agendamento + callback)
- `alertas-novidades.spec.js` — 7 cenários (novidades + variações + alertas)
- Tasks: `mise run test:e2e:lojas`, `test:e2e:buscas-agendadas`, `test:e2e:publicar`, `test:e2e:alertas`

### 6. CI/CD Otimizado

- `dorny/paths-filter` para skip condicional de jobs
- Deploy conservador (ADR-0025): "na dúvida, deploya"
- Migrations EF Core automáticas antes do deploy
- `paths-ignore` expandido (.kiro/, .vscode/, *.md, etc.)

### 7. Refatoração

- Removida nomenclatura "Compat" — endpoints são definitivos
- Removidos endpoints v2 incompletos
- Removido `PublicarCompatRequest` → `PublicarRequest`

### 8. Segurança

- Service account key exposta no git → revogada (ADR-0026)
- Binários Go compilados removidos + .gitignore
- Scripts legados removidos

### 9. Documentação

- Portal unificado de APIs (`docs/09-api-reference.md`) gerado automaticamente
- Task `mise run gen:api-reference` + drift check no pre-push
- READMEs atualizados (raiz, src/, web/, docs/)
- ADR-0025 (deploy conservador) + ADR-0026 (rotação key)
- Sprint S28 planejada (7 tasks de validação final)

## Métricas

| Métrica | Antes | Depois |
|---------|-------|--------|
| Testes C# | 61 | 68 |
| Testes E2E | 0 | 22 |
| Testes frontend unit | 141 | 141 |
| ADRs | 24 | 26 |
| Endpoints novos | — | 3 |
| RPCs gRPC novos | — | 2 |
| Linhas removidas | — | ~1200 |
| Tasks mise novas | — | 5 |

## Tasks concluídas

- T-0005: Alertas configuráveis por usuário
- T-0028: Configurar coleta no scheduler
- T-0029: Deploy da nova API em produção
- T-0036: Migração scripts → mise

## Sprint S28 planejada

- T-0045: Encriptar credenciais tenant
- T-0046: Validar 11 fluxos em produção
- T-0047: Observabilidade
- T-0048: Limpeza final
- T-0049: Reconciliação buscas → Scheduler
- T-0050: Testar envio real Publisher
- T-0051: Documentação final
- T-0052: API reference gerado automaticamente
