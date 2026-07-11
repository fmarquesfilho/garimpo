# Fluxo de Variação de Preços

Documento de referência para o fluxo end-to-end de monitoramento de preços:
coleta de snapshots → detecção de variações → exibição na UI → publicação de ofertas.

---

## Visão geral do fluxo

```
┌─────────────┐    gRPC Fetch     ┌─────────────┐    INSERT     ┌───────────┐
│  Scheduler  │ ─────────────────►│  Collector  │ ─────────────►│  BigQuery │
│  (Go:50054) │  keyword/shop_id  │  (Go:50051) │  snapshots    │  garimpo. │
│  cron BRT   │                   │  Shopee API │               │  snapshots│
└──────┬──────┘                   └─────────────┘               └─────┬─────┘
       │                                                              │
       │ jobs configurados                                            │ SELECT
       │ via SetSchedule gRPC                                         │ (janela N dias)
       │                                                              │
       │                          ┌─────────────┐    BigQuery    ┌────▼──────┐
       │                          │   API C#    │◄───────────────│  Analyzer │
       │                          │   (:8080)   │  HTTP proxy    │ (Py:8060) │
       │                          │             │  /novidades    │  FastAPI  │
       │                          └──────┬──────┘  /quedas       └───────────┘
       │                                 │         /evolucao
       │                                 │ JSON
       │                          ┌──────▼──────┐
       │                          │  Frontend   │
       │                          │  SvelteKit  │
       │                          │  /lojas     │
       │                          │  aba Preços │
       │                          └─────────────┘
```

---

## Componentes envolvidos

### 1. Scheduler (Go, gRPC :50054)

**Papel**: orquestra coletas periódicas via cron.

- Timezone: `America/Sao_Paulo` (BRT)
- Jobs são criados dinamicamente via `SetSchedule` gRPC
- Cada job tem: `job_id`, `cron_expression`, `params` (keyword, type, owner_uid)
- Ao disparar, chama `collector.Fetch()` com a keyword/shop_id configurada

**Onde fica o estado**: em memória (map de jobs). Jobs não persistem entre restarts — precisam ser reconfigurados.

**Código**: `services/scheduler/server.go`

### 2. Collector (Go, gRPC :50051)

**Papel**: busca produtos na Shopee e grava snapshot no BigQuery.

- Autenticação HMAC-SHA256 (AppID + Secret)
- Throttling: 200ms entre páginas, 60s entre lojas
- Suporta busca por keyword (`Fetch`) ou por loja (`FetchShop`)
- Grava cada produto coletado como uma linha em `garimpo.snapshots`

**Onde grava**: BigQuery tabela `garimpo.snapshots`

**Código**: `services/collector/server.go`, `services/collector/pipeline.go`

### 3. BigQuery — tabela `snapshots`

**Papel**: armazena a "foto" do mercado a cada coleta.

**Schema** (produção):

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `coletado_em` | TIMESTAMP | Quando foi coletado (partition key) |
| `categoria` | STRING | Categoria do produto |
| `keyword` | STRING | Keyword ou ID da loja (ex: `loja-920292999`) |
| `estrategia` | STRING | Estratégia de curadoria usada |
| `posicao` | INTEGER | Posição no ranking daquele dia |
| `produto_id` | STRING | ID do produto na Shopee |
| `nome` | STRING | Nome do produto |
| `preco` | FLOAT | Preço atual em BRL |
| `comissao` | FLOAT | Comissão de afiliado (fração: 0.12 = 12%) |
| `vendas` | INTEGER | Volume de vendas |
| `nota` | FLOAT | Nota do produto (0-5) |
| `score` | FLOAT | Score de curadoria calculado |
| `imagem` | STRING | URL da imagem do produto |
| `link` | STRING | Link de afiliado |
| `loja` | STRING | Nome da loja na Shopee |

**Particionamento**: por `DATE(coletado_em)` (queries baratas por janela temporal).

**Volume atual** (jul/2026): ~1576 rows, crescendo ~800/dia com 2 lojas monitoradas.

### 4. Analyzer (Python, REST :8060)

**Papel**: queries analíticas que comparam snapshots entre datas.

**Endpoint principal**: `GET /novidades?busca_id=<keyword>&dias=<N>`

Lógica:
1. Seleciona todos os snapshots da janela (últimos N dias) que casam com `keyword LIKE %busca_id%`
2. Para cada `produto_id`, compara primeiro preço vs último preço
3. Classifica:
   - **Produto novo**: aparece apenas 1 vez na janela (nunca visto antes)
   - **Variação**: `|preco_atual - preco_primeiro| / preco_primeiro > 1%`
4. Retorna listas separadas de `produtos_novos` e `variacoes`

**Outros endpoints**:
- `GET /quedas?dias=7&threshold=0.15` — só quedas acima de 15%
- `GET /evolucao?dias=30` — série temporal de preço médio por loja/dia
- `GET /coletas?dias=30` — histórico de coletas (snapshots agrupados por dia+keyword)
- `GET /estatisticas?dias=30` — resumo geral (mediana, média, total produtos)

**Modo mock**: com `MOCK_DATA=true`, retorna dados fictícios sem BigQuery (para dev local).

**Código**: `services/analyzer/routes/novidades.py`, `services/analyzer/routes/quedas.py`

### 5. API C# (ASP.NET Core, :8080)

**Papel**: proxy autenticado entre frontend e analyzer.

O endpoint `GET /api/lojas/novidades` no C# faz:
1. Recebe request autenticado (Firebase JWT)
2. Lê `Analyzer:BaseUrl` da config (default: `http://localhost:8060`)
3. Faz `GET {analyzerUrl}/novidades?busca_id={busca_id}&dias={dias}`
4. Retorna JSON para o frontend
5. Se o analyzer estiver offline, retorna fallback vazio (sem erro pro usuário)

**Código**: `src/Garimpei.Api/Endpoints/LojasEndpoints.cs`

### 6. Frontend (SvelteKit)

**Papel**: exibe variações na aba "📉 Preços" da página /lojas.

Fluxo na UI:
1. Usuário navega para `/lojas`
2. Seleciona uma loja monitorada
3. Aba "📉 Preços" chama `GET /api/lojas/novidades?busca_id=<keyword>&dias=7`
4. Exibe tabela: Produto | Antes | Agora | Variação% | Detectado em
5. Botão "📤 Publicar" navega para `/publicar` com dados pré-preenchidos

---

## Onde os dados ficam (resumo)

| Dado | Storage | Quem escreve | Quem lê |
|------|---------|-------------|---------|
| Buscas/lojas monitoradas | PostgreSQL (`Buscas`) | C# API (CRUD) | C# API, Scheduler |
| Snapshots de preço | BigQuery (`snapshots`) | Collector (Go) | Analyzer (Python) |
| Variações detectadas | Calculado em runtime | — | Analyzer (query BQ) → C# → Frontend |
| Jobs do scheduler | Memória (scheduler) | gRPC `SetSchedule` | gRPC `ListJobs` |
| Publicações | PostgreSQL (`Publicacoes`) | C# API | C# API, Frontend |

---

## Como testar localmente

### Modo mock (sem BigQuery)

```bash
# Terminal 1: Analyzer com dados fictícios
cd services/analyzer
MOCK_DATA=true python3 -m uvicorn main:app --host 127.0.0.1 --port 8060

# Terminal 2: API C# apontando para analyzer local
cd src
ASPNETCORE_ENVIRONMENT=Development Analyzer__BaseUrl=http://localhost:8060 dotnet run --project Garimpei.Api

# Terminal 3: Frontend
cd web && VITE_API_BASE=http://localhost:5000 npm run dev

# Testar diretamente
curl "http://localhost:8060/novidades?busca_id=perfumes-importados&dias=7"
curl "http://localhost:5000/api/lojas/novidades?busca_id=perfumes-importados&dias=7" \
  -H "X-Dev-User: dev-user-001"
```

### Com BigQuery emulator (dados realistas)

```bash
# Subir dependências
docker compose up -d postgres bigquery-emulator

# Popular com dados fictícios
python3 scripts/seed-local-test.py

# Subir analyzer (sem mock, com emulator)
cd services/analyzer
BQ_PROJECT=garimpei-dev BIGQUERY_EMULATOR_HOST=localhost:9050 \
  python3 -m uvicorn main:app --host 127.0.0.1 --port 8060
```

### Testar em produção

```bash
# Listar lojas monitoradas
./scripts/prod-api.sh GET /api/lojas

# Verificar novidades/variações
./scripts/prod-api.sh GET '/api/lojas/novidades?busca_id=loja-920292999&dias=7'

# Verificar quedas significativas (via analyzer direto no BQ)
bq query --use_legacy_sql=false --project_id=garimpo-500114 \
  "SELECT keyword, COUNT(*) FROM \`garimpo-500114.garimpo.snapshots\` GROUP BY keyword"
```

---

## Problemas conhecidos e soluções

| Problema | Causa | Solução aplicada |
|----------|-------|-----------------|
| `/api/lojas/novidades` retorna vazio | Analyzer crashava: coluna `em` não existe (é `coletado_em`) | Fix em `f2fab6e` — renomear para `coletado_em` em todas as queries |
| Scheduler sem jobs após restart | Jobs ficam só em memória | Pendente: T-0028 (configurar coleta no scheduler) |
| Zero snapshots = zero variações | Precisa de pelo menos 2 coletas | Scheduler precisa rodar >1x para ter diff |
| Analyzer offline = fallback vazio | C# retorna arrays vazios (sem erro) | Por design — graceful degradation |

---

## Próximos passos

1. **T-0028**: Criar mecanismo de auto-sync (buscas ativas no PG → jobs no scheduler)
2. **T-0007**: Scoring ML no analyzer (substituir rule-based por modelo treinado)
3. **Alertas de preço**: scheduler → Cloud Tasks → C# → publisher (Telegram/WhatsApp) — ADR-0023
4. **Histórico de variações**: armazenar detecções em tabela dedicada no BQ
