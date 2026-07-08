# Tasks

## Task 1: Ajustar Scheduler Go para suportar keywords plural no branch default

- [x] Em `services/scheduler/jobs.go`, no branch `default` de `executeJob`, adicionar leitura de `params["keywords"]` (plural, comma-separated)
- [x] Se `params["keywords"]` presente e não vazio, iterar sobre cada keyword e chamar `Collector.Fetch` individualmente (mesmo padrão do `shop_collection` com keywords)
- [x] Manter fallback para `params["keyword"]` (singular) para compatibilidade com jobs legados
- [x] Passar `params["owner_uid"]` no `FetchRequest` para que os snapshots fiquem associados ao tenant
- [x] Verificar que `go build ./services/scheduler/...` compila sem erros

## Task 2: Criar componente PainelNovidades.svelte

- [x] Criar `web/src/lib/components/PainelNovidades.svelte` com props `buscaId: string` e `keywords: string[]`
- [x] Importar e chamar `buscarNovidades({ buscaId, dias: 7 })` de `$lib/api.js` no `onMount` ou efeito reativo
- [x] Exibir loading state enquanto a request está em andamento
- [x] Renderizar produtos novos (`produtos_novos[]`) usando `ProductCard` com layout "compact"
- [x] Renderizar variações de preço (`variacoes[]`) com indicador de percentual (positivo/negativo)
- [x] Exibir empty state "Aguardando primeira coleta..." quando não há dados
- [x] Usar primitivos do design system (`Card`, `Badge`, `Loading` de `$lib/components/ui`)

## Task 3: Atualizar BuscaCard.svelte com cronLabel e ação de selecionar

- [x] Adicionar função `cronLabel(cron)` que converte expressões cron comuns em texto legível ("a cada 8h", "a cada 12h", "diária 9h")
- [x] Substituir badge estática `⏱ agendada` por `⏱ {cronLabel(busca.cron)}` quando cron presente
- [x] Adicionar prop `onselecionar` (callback) ao componente
- [x] Ao clicar no card (ou num botão "ver resultados"), chamar `onselecionar(busca)`
- [x] Indicar visualmente quando o card está selecionado (borda destacada ou bg-accent)

## Task 4: Integrar PainelNovidades em GerenciarBuscas.svelte

- [x] Adicionar estado `buscaSelecionada` ($state) em GerenciarBuscas
- [x] Passar `onselecionar` prop para cada `BuscaCard`, setando `buscaSelecionada` ao clicar
- [x] Renderizar `PainelNovidades` condicionalmente abaixo da lista quando `buscaSelecionada` não é null
- [x] Limpar `buscaSelecionada` quando a busca é removida ou o form de nova busca abre
- [x] Verificar que `npm run check` e `npm run build` passam sem erros no diretório `web/`
