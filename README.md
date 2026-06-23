# Garimpo

Plataforma de curadoria e publicação automatizada para afiliados Shopee.
Busca produtos, monitora lojas, rankeia por estratégia, e publica em canais
(Telegram, WhatsApp) com templates, fotos e agendamento — tudo com rastreamento
de conversão.

## Funcionalidades

- **Curadoria inteligente** — busca na Shopee por keyword ou categoria, rankeia
  por "teor" (comissão × demanda × avaliação), filtra produto-fantasma.
- **Monitoramento de lojas** — acompanha lojas específicas via `shopOfferV2`,
  detecta novos produtos e variações de preço.
- **Publicação rica** — templates customizáveis, editor WYSIWYG, foto do produto,
  botão inline "Comprar", envio para múltiplos destinos (Telegram, WhatsApp futuro).
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
go test ./...                    # todos os pacotes
go test ./... -cover             # com cobertura
cd web && npm run build          # verifica o frontend
```

Cobertura atual: publish 91%, scoring 90%, source 87%, strategy 86%, engine 85%.

## Deploy (GCP)

Push para `main` dispara o workflow `deploy-gcp.yml`:
1. Testes Go + build do frontend
2. Build da imagem Docker (com `-tags gcp` para BigQuery)
3. Deploy no Cloud Run + Firebase Hosting

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
  publish/           saída: Dispatcher, Sender (Telegram), Templates, Destinos
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

## Variáveis de ambiente

| Variável | Onde | Para quê |
|----------|------|----------|
| `SHOPEE_APP_ID` | Secret Manager | Credencial da API de afiliados |
| `SHOPEE_SECRET` | Secret Manager | Assinatura HMAC-SHA256 |
| `TELEGRAM_BOT_TOKEN` | Secret Manager | Bot do Telegram (@BotFather) |
| `COLETA_TOKEN` | Secret Manager | Protege endpoints de coleta/scheduler |
| `GOOGLE_CLOUD_PROJECT` | Cloud Run env | Projeto GCP |
| `BQ_DATASET` | Cloud Run env | Dataset BigQuery (default: `garimpo`) |

## Documentação

- `docs/DEPLOY_GCP.md` — runbook completo de deploy na GCP
- `docs/COLETA.md` — o que é coletado, onde fica, como ligar
- `docs/APIS.md` — referência das APIs Shopee e Instagram
- `docs/MODELO.md` — modelo de negócio e roadmap
- `docs/MANUAL.md` — manual de uso (teor, selos, fluxo)
- `docs/JORNADA.md` — jornada do usuário e pontos de decisão
- `docs/CIENCIA_DE_DADOS.md` — guia de análise por maturidade
