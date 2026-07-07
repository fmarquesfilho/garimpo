# Sessão 07/Julho 2026 — Lojas: integração de keywords + agendamento na UI e reformulação de layout

## Objetivo

Tornar alcançável pela interface o fluxo de monitoramento de lojas que o backend já
suportava desde a sessão 06/07: cadastrar **palavras-chave** (conjunto vazio ou não) e um
**agendamento** de coleta no momento em que a loja é adicionada. Unificar o agendamento
de todas as buscas (com ou sem loja) pelo Scheduler e reformular o layout da página
`/lojas` alinhando-a ao design system.

## Contexto

`POST /api/lojas` já aceitava `keywords[]` e `cron`, persistia em `Busca.Keywords` /
`Busca.CronExpression` e registrava o job periódico via `scheduler.SetSchedule`. Porém o
formulário de adicionar loja (`FormAdicionarLoja.svelte`) só coletava a URL/ID e a origem
— as palavras-chave e o agendamento eram **inalcançáveis pela UI**. Os componentes
reutilizáveis (`AgendadorBusca`, `TagInput`) já existiam mas não estavam conectados a esse
fluxo.

## Modelo conceitual

Uma **busca agendada** é o conceito guarda-chuva — pode ou não ter loja (`shop_ids`), pode
ou não ter palavras-chave. Loja e busca agendada se complementam:

| keywords | loja (shop_ids) | significado | entrada na UI |
|---|---|---|---|
| ✓ | ✗ | keywords na Shopee inteira | "Buscas por palavra-chave" |
| ✓ | ✓ | keywords dentro de uma loja | "Adicionar loja" |
| ✗ | ✓ | garimpar todos os produtos de uma loja | "Adicionar loja" |
| ✗ | ✗ | inválido — exige pelo menos um | — |

## Entregas

### 1. Integração keywords + agendamento no formulário de loja

- `FormAdicionarLoja.svelte`: campos de **palavras-chave** (`TagInput`, opcional — vazio
  permitido) e **agendamento** (`AgendadorBusca`, padrão *a cada 8h*, editável), passados a
  `adicionarLoja({ input, keywords, cron, origemPadrao })`.
- Correção de bug: a mensagem de sucesso usava `r.shop_id` (inexistente na resposta) →
  passou a usar o nome resolvido da loja (`r.keyword`).
- `AgendadorBusca.svelte`: novo preset **"A cada 8h"** (`0 */8 * * *`) e prop
  `permitirNunca` (loja monitorada sempre coleta periodicamente, então não oferece "Nunca").

### 2. Buscas por palavra-chave (renomeação + agendamento efetivo)

- `GerenciarBuscas.svelte`: seção renomeada de "Buscas Agendadas" para **"Buscas por
  palavra-chave"**, com texto explicando que rodam na Shopee inteira e que o monitoramento
  por loja é feito no formulário de loja (complementaridade explícita).
- O `cron` escolhido agora chega ao backend: `POST /api/buscas` passou a persistir
  `CronExpression` e a registrar/pausar o job no Scheduler (antes o cron era descartado).

### 3. Backend — agendamento unificado pelo Scheduler (ownership preservado)

- Novo helper compartilhado `SchedulerJobs` (`src/Garimpei.Api/Endpoints/SchedulerJobs.cs`)
  que monta/registra/pausa o job (`SetSchedule`) a partir de uma `Busca`:
  - com `ShopIds` → job `shop_collection`;
  - sem loja → job `keyword_search` (Fetch por keyword);
  - cron padrão `0 */8 * * *` quando ausente.
- Usado por `/api/lojas` **e** `/api/buscas` — **todo agendamento passa pelo Scheduler**
  (ADR-0023; o Scheduler enfileira os alertas via Cloud Tasks).
- `GET /api/buscas` passou a devolver as palavras-chave de filtro reais (`b.Keywords`) e o
  `cron`, em vez de conflar o nome da loja com as keywords.

### 4. Reformulação de layout de `/lojas`

- `PageHeader` (eyebrow + título display + subtítulo) no lugar do `<h1>` cru.
- Seção **"Suas lojas"** promovida ao topo: grid responsivo de cards de altura uniforme
  com chips de palavras-chave, estado de coleta e contador; card selecionado destacado e
  painel de detalhes (abas Produtos/Novidades/Preços) logo abaixo.
- Formulário de loja, "Buscas por palavra-chave" e "Alertas" reordenados como configuração
  secundária, após a lista de lojas.
- `EmptyState` reutilizado nos estados vazios; tokens semânticos
  (`muted-foreground`/`border`/`card`) no lugar de cores cruas.

## Regras arquiteturais respeitadas

- **Ownership:** C# continua dono exclusivo do PostgreSQL (`Buscas`); Collector Go faz o
  I/O externo (`ResolveShop`); Scheduler é dono dos jobs. Nada disso mudou.
- **Agendamento sempre via Scheduler** (`SetSchedule`) → Cloud Tasks para alertas
  (ADR-0023). Nenhuma chamada direta a Cloud Tasks no C#.
- **Componentização shadcn** (sessão 03/07): reuso de `Card`, `Button`, `Input`, `Select`,
  `Badge`, `TagInput`, `AgendadorBusca`, `PageHeader`, `EmptyState`, `Tabs` — sem CSS
  scoped novo.

## Arquivos alterados

| Arquivo | Mudança |
|---|---|
| `web/src/lib/components/FormAdicionarLoja.svelte` | keywords + agendamento + fix mensagem |
| `web/src/lib/components/AgendadorBusca.svelte` | preset "A cada 8h" + prop `permitirNunca` |
| `web/src/lib/components/GerenciarBuscas.svelte` | renomeação + copy de complementaridade |
| `web/src/routes/lojas/+page.svelte` | reformulação de layout |
| `src/Garimpei.Api/Endpoints/SchedulerJobs.cs` | **novo** — helper compartilhado do Scheduler |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | usa o helper compartilhado |
| `src/Garimpei.Api/Endpoints/BuscasEndpoints.cs` | GET com keywords/cron; POST agenda no Scheduler |
| `src/Garimpei.Tests/Integration/BuscasAgendadasTests.cs` | testes do helper `SchedulerJobs` |
| `web/tests/lojas-cadastro.spec.js`, `web/tests/buscas-agendadas.spec.js` | contrato/UI atualizados |

## Verificação

- Frontend: `npm run check` (0/0), `lint:js`, `format:check`, `test:unit` (141) e `build` — verdes.
- Backend: cobertura do helper `SchedulerJobs` em `BuscasAgendadasTests`. `dotnet build/test`
  e os E2E (`mise run test:e2e:lojas`, `test:e2e:buscas-agendadas`) dependem do CI/serviços.
