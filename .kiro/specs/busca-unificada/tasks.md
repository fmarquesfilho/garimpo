# Tasks

## Task 1: Backend — Estender entidade Busca com campos de filtro

- [ ] Em `src/Garimpei.Domain/Entities/Busca.cs`, adicionar: `decimal? ComissaoMin`, `int? VendasMin`, `string[]? Categorias`, `string[]? Fontes`
- [ ] Gerar migration EF Core: `dotnet ef migrations add AddFiltersToBusca --project src/Garimpei.Infrastructure --startup-project src/Garimpei.Api`
- [ ] Verificar que a migration é apenas `AddColumn` (nullable, sem data migration)
- [ ] Verificar `dotnet build` sem erros e `dotnet test` passa

## Task 2: Backend — Estender POST /api/buscas com novos campos

- [ ] Em `SyncBuscaRequest`, adicionar: `long[]? ShopIds`, `decimal? ComissaoMin`, `int? VendasMin`, `string[]? Categorias`, `string[]? Fontes`, `string? Marketplaces`
- [ ] No handler POST /api/buscas, persistir `ShopIds`, `ComissaoMin`, `VendasMin`, `Categorias`, `Fontes` quando presentes no request
- [ ] Se `ShopIds` presente e `CronExpression` presente, chamar `SchedulerJobs.RegisterAsync` com tipo correto (shop_collection se tem shops, keyword_search se não)
- [ ] Manter backward compat: requests sem os novos campos continuam funcionando identicamente
- [ ] No handler GET /api/buscas, incluir `comissao_min`, `vendas_min`, `categorias`, `fontes`, `marketplaces` na response
- [ ] Verificar `dotnet build` e `dotnet test`

## Task 3: Frontend — Criar módulo busca-unificada-logic.js

- [ ] Criar `web/src/lib/busca-unificada-logic.js` com tipos e funções puras
- [ ] Implementar `configToPayload(config)` que converte estado do componente para SyncBuscaRequest
- [ ] Implementar `payloadToConfig(busca)` que converte response do GET /api/buscas para estado interno
- [ ] Implementar `gerarResumo(config)` que gera string compacta (ex: `"sérum" · 2 lojas · 2 filtros`)
- [ ] Implementar `contarFiltrosAtivos(config)` que conta filtros não-default
- [ ] Escrever testes unitários em `web/src/tests/busca-unificada.test.js` para cada função
- [ ] Verificar `npx vitest run`

## Task 4: Frontend — Criar componente BuscaUnificada.svelte

- [ ] Criar `web/src/lib/components/BuscaUnificada.svelte`
- [ ] Implementar input de keywords com debounce 400ms, ESC para limpar, Enter para executar, botão ✕
- [ ] Implementar seleção de lojas via TagInput (resolve URL/ID via `adicionarLoja` API, exibe como tag com nome)
- [ ] Implementar seção de filtros colapsável (Collapsible): comissão mín (Select), vendas mín (Input), categorias (TagInput com autocomplete)
- [ ] Implementar ToggleGroup (multiple) para fontes de dados
- [ ] Implementar botão "💾 Salvar busca" que abre mini-panel com AgendadorBusca (cron opcional) antes de confirmar
- [ ] Implementar chips de buscas salvas (clickáveis para carregar, ✕ para remover)
- [ ] Implementar modo colapsado com compact summary (keywords, shop count, filter count)
- [ ] Emitir resultados via props callback (`onresultados`, `oncarregando`, `onerro`)
- [ ] Script block ≤ 180 linhas (toda lógica complexa em busca-unificada-logic.js)
- [ ] Zero `<style>` blocks, zero inline `<button>`/`<input>`/`<select>` — usar exclusivamente ui/ components
- [ ] Verificar `npm run check` sem erros

## Task 5: Frontend — Integrar BuscaUnificada na página /

- [ ] Em `web/src/routes/+page.svelte`, remover imports de FilterBar, FormAdicionarLoja, GerenciarBuscas
- [ ] Importar e renderizar BuscaUnificada no topo (acima dos resultados)
- [ ] Conectar callbacks: `onresultados` popula o grid, `oncarregando` controla Loading, `onerro` exibe Alert
- [ ] Remover estado duplicado que agora vive dentro do BuscaUnificada (busca, comissaoMin, vendasMin, etc.)
- [ ] Manter o grid de ProductCards e a lógica de publicar/favoritar
- [ ] Verificar que a seção Collapsible "⚙️ Configuração" pode ser removida (tudo está no BuscaUnificada)
- [ ] Script block da página ≤ 180 linhas
- [ ] Verificar `npm run check`, `npm run lint:js`, `npm run build`

## Task 6: Frontend — Remover componentes substituídos

- [ ] Deletar `web/src/lib/components/FilterBar.svelte`
- [ ] Deletar `web/src/lib/components/FormAdicionarLoja.svelte`
- [ ] Deletar `web/src/lib/components/GerenciarBuscas.svelte`
- [ ] Verificar com `npx knip --include files` que nenhum arquivo referencia os deletados
- [ ] Atualizar `docs/componentes.md` (remover FilterBar, FormAdicionarLoja, GerenciarBuscas; adicionar BuscaUnificada)
- [ ] Verificar `npm run check`, `npm run build`, `npx vitest run`

## Task 7: Frontend — Atualizar testes unitários e E2E

- [ ] Atualizar `web/src/tests/descobrir.test.js` se necessário (montarResultados não muda, mas verificar imports)
- [ ] Atualizar `web/tests/descobrir.spec.js` para interagir com BuscaUnificada (input, filtros, toggle fontes)
- [ ] Atualizar `web/tests/buscas-agendadas.spec.js` (não precisa mais expandir "⚙️ Configuração" — salvar está no BuscaUnificada)
- [ ] Atualizar `web/tests/lojas-resolve-shop.spec.js` e `lojas-cadastro.spec.js` (adicionar loja via campo integrado)
- [ ] Verificar `npx vitest run` (todos os 158+ testes passam)
- [ ] Verificar `npm run lint:js` sem warnings

## Task 8: Verificação final e documentação

- [ ] Rodar `mise run prepush` completo (todos os checks devem passar)
- [ ] Atualizar `docs/05-manual-do-usuario.md` com nova descrição do componente unificado
- [ ] Atualizar `docs/componentes.md` com a documentação do BuscaUnificada
- [ ] Atualizar `docs/impacto-migracao-ui.md` com seção sobre a unificação de filtros
- [ ] Criar `docs/legado/SESSAO_2026-07-08_BUSCA_UNIFICADA.md` documentando a sessão
- [ ] Verificar `mise run check:api-contract` (21 rotas devem bater)
