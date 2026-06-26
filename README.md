# Garimpei

Plataforma de curadoria e publicação automatizada para afiliados Shopee.
Busca produtos, monitora lojas, rankeia por estratégia, e publica em canais
(Telegram, WhatsApp) com templates, fotos e agendamento — tudo com rastreamento
de conversão.

## Funcionalidades

- **Curadoria inteligente** — busca na Shopee por keyword, rankeia
  por potencial de retorno (comissão × demanda × avaliação), filtra produto-fantasma.
  Interface simplificada com busca proeminente e filtros colapsáveis.
- **Monitoramento de lojas** — acompanha lojas específicas via `productOfferV2`,
  detecta novos produtos e variações de preço. Aceita links curtos, slugs e IDs.
- **Oportunidades** — feed unificado de todas as lojas com quedas de preço,
  produtos novos e altas, filtráveis por período.
- **Alertas automáticos** — notificações de preço via Telegram quando variações
  significativas são detectadas (bot @AlertaGarimpeiBot).
- **Publicação rica** — templates customizáveis, editor WYSIWYG, foto do produto,
  botão inline "Comprar", envio para múltiplos destinos (Telegram, WhatsApp).
  WhatsApp suporta até 5 grupos por destino.
- **Agendamento** — publica no horário configurado via Cloud Scheduler.
- **Rastreamento** — cada publicação gera um `sub_id` que identifica canal +
  estratégia + data, cruzável com o `validatedReport` da Shopee.
- **Pipeline de filtros modular** — Chain of Responsibility extensível; modo
  "sem filtro" para exploração de lojas, modo curadoria com pisos configuráveis.
- **Landing page** — protege a aplicação; exige login Google para acessar.

## Rodar localmente

```bash
# API (fonte CSV para teste rápido)
go run ./cmd/garimpo-api

# API com Shopee ao vivo
export SHOPEE_APP_ID=... SHOPEE_SECRET=...
go run ./cmd/garimpo-api -fonte shopee -keyword "skincare"

# Frontend (em outro terminal)
cd web && npm install && npm run dev
```

A API escuta na porta 8080; o front aponta para ela automaticamente em dev.

## Testes

```bash
go test ./...                    # todos os pacotes Go
cd web && npx vitest run         # testes de componente Svelte (Vitest)
cd web && npm test               # testes E2E (Playwright)
```

- **Go**: cobertura ~90% nos pacotes publish, scoring, source, strategy
- **Vitest**: testa componentes Svelte reais (SeletorGrupo, etc.) com jsdom
- **Playwright**: smoke tests, validação de rotas, mocks de API

## Deploy (GCP)

Push para `main` dispara o workflow `deploy-gcp.yml`:
1. Testes Go (gate de qualidade)
2. Build da imagem Docker multi-stage (Node + Go, com `-tags gcp`)
3. Deploy no Cloud Run (API + frontend numa única imagem)
4. Testes frontend rodam em paralelo (Vitest + Playwright)

O frontend é servido pelo próprio Go no Cloud Run (SPA handler com fallback).
Não usa Firebase Hosting.

URL de produção: `https://garimpei.app.br`

Detalhes em `docs/DEPLOY_GCP.md`.

## Estrutura

```
cmd/garimpo          CLI (composição: fonte + estratégia)
cmd/garimpo-api      servidor HTTP (API JSON para o frontend)
internal/
  domain/            núcleo: Product, Scored
  engine/            orquestra: fonte → filtros → scoring → ranking
  strategy/          estratégias (Niche, Diversified) + Pipeline de filtros
  scoring/           matemática: valor esperado, normalização, suspeito
  source/            adaptadores de entrada (CSV, Shopee keyword, Shopee shop)
  publish/           saída: Dispatcher, Sender (Telegram, WhatsApp/Maytapi), Templates, Destinos
  httpapi/           handlers HTTP (rotas com método explícito, Go 1.22+)
  store/             persistência (NopStore / BigQueryStore com -tags gcp)
  scheduler/         Cloud Scheduler (cria jobs de coleta e publicação)
  auth/              Firebase Auth (verifica tokens)
  logs/              logging estruturado (slog)
web/                 frontend SvelteKit 5 + Vite 6
  src/lib/components/  componentes reutilizáveis (TagInput, FilterBar, BuscaCard,
                       CandidateCard, RichEditor, ScoreMeter, StrategyToggle)
  src/routes/          páginas: /, /lojas, /publicar, /publicacoes,
                       /coletas, /canais, /quadro, /estatisticas
deploy/              systemd/nginx (VPS), schema BigQuery, scheduler
docs/                documentação detalhada
```

## API — endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/health` | Status da API |
| GET | `/api/candidatos` | Busca + ranking (aceita `fonte`, `sem_filtro`, `shop_ids`) |
| GET | `/api/comparar` | Nicho vs. diversificada lado a lado |
| POST | `/api/eventos` | Registra decisão de curadoria |
| POST | `/api/publicar` | Publica oferta (destino + template + imagem) |
| POST | `/api/coletar` | Coleta agendada (Cloud Scheduler) |
| GET | `/api/estatisticas` | Resumo descritivo dos snapshots |
| GET | `/api/coletas` | Histórico de coletas executadas |
| GET | `/api/conversoes` | Relatório de publicações por canal |
| GET | `/api/lojas/novidades` | Novos produtos + variações de preço |
| GET/POST/DELETE | `/api/buscas` | Perfis de busca (keywords + shop_ids + cron) |
| GET/POST/DELETE | `/api/destinos` | Canais de publicação (Telegram, WhatsApp) |
| GET/POST/DELETE | `/api/templates` | Modelos de mensagem |
| POST | `/api/templates/preview` | Renderiza preview de template |
| GET/POST | `/api/publicacoes` | Publicações agendadas/enviadas/erros |
| POST | `/api/publicar-pendentes` | Executa publicações agendadas vencidas |
| GET | `/api/whatsapp/grupos` | Lista grupos WhatsApp disponíveis (Maytapi) |
| POST | `/api/resolver-link` | Resolve link curto da Shopee |

## Variáveis de ambiente

| Variável | Onde | Para quê |
|----------|------|----------|
| `SHOPEE_APP_ID` | Secret Manager | Credencial da API de afiliados |
| `SHOPEE_SECRET` | Secret Manager | Assinatura HMAC-SHA256 |
| `TELEGRAM_BOT_TOKEN` | Secret Manager | Bot do Telegram (@BotFather) |
| `WHATSAPP_PRODUCT_ID` | Secret Manager | Maytapi Product ID |
| `WHATSAPP_PHONE_ID` | Secret Manager | Maytapi Phone ID |
| `WHATSAPP_API_KEY` | Secret Manager | Maytapi API Token |
| `COLETA_TOKEN` | Secret Manager | Protege endpoints de coleta/scheduler |
| `GOOGLE_CLOUD_PROJECT` | Cloud Run env | Projeto GCP |
| `BQ_DATASET` | Cloud Run env | Dataset BigQuery (default: `garimpo`) |
| `WEB_DIR` | Cloud Run env | Diretório do frontend (default: `/web`) |

## Documentação

- `docs/DEPLOY_GCP.md` — runbook completo de deploy na GCP
- `docs/COLETA.md` — o que é coletado, onde fica, como ligar
- `docs/APIS.md` — referência das APIs Shopee e Instagram
- `docs/MODELO.md` — modelo de negócio e roadmap
- `docs/MANUAL.md` — manual de uso (teor, selos, fluxo)
- `docs/JORNADA.md` — jornada do usuário e pontos de decisão
- `docs/CIENCIA_DE_DADOS.md` — guia de análise por maturidade
