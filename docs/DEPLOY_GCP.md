# Deploy na GCP (Cloud Run + BigQuery)

Arquitetura serverless: a API Go roda no **Cloud Run** e serve tanto a API quanto
o frontend estático (SPA). Os dados minerados e as decisões de curadoria vão para
o **BigQuery** (análise barata + Looker Studio).

```
navegador ──https──> Cloud Run (garimpo-api)
                       ├─ /           -> frontend estático (SPA, web/build)
                       ├─ /api/**     -> API JSON (curadoria, publicação, coleta)
                       ├─ Secret Manager: SHOPEE_*, TELEGRAM_*, WHATSAPP_*
                       └─ grava eventos -> BigQuery
                                              └─ Looker Studio / Python
```

Custo no seu volume: Cloud Run praticamente de graça (escala a zero), BigQuery
dentro do free tier (10 GB de storage + 1 TB de consulta/mês).
Deve ficar bem abaixo de R$50 — confirme no Billing.

## RUNBOOK: do zero ao ar (checklist ordenado)

Faça na ordem. Os detalhes de cada passo estão nas seções abaixo.

- [ ] **1.** Criar projeto GCP e instalar o `gcloud` (e `bq`). `gcloud auth login`.
- [ ] **2.** Definir `PROJECT_ID` e `REGION` e rodar o bloco de **Setup** (habilita
      APIs, cria Artifact Registry, cria os segredos, cria o dataset e
      as tabelas do BigQuery).
- [ ] **3.** Dar à service account do Cloud Run os papéis de BigQuery + Secret
      Manager (seção *Permissões*).
- [ ] **4.** Criar a service account de deploy do CI e gerar `key.json`
      (seção *Service account para o CI*).
- [ ] **5.** No GitHub, criar os secrets `GCP_PROJECT_ID` e `GCP_SA_KEY`.
      (Os segredos da Shopee/WhatsApp **não** vão para o GitHub — ficam no Secret Manager.)
- [ ] **6.** `git push` na `main` → o workflow `deploy-gcp.yml` testa, builda a
      imagem, sobe no Cloud Run. Acompanhe na aba **Actions**.
- [ ] **7.** **Testar** (seção *Primeiro deploy → verificação*): abrir a URL do
      Cloud Run, conferir `/api/health`, fazer uma busca e uma publicação.
- [ ] **8.** *(Opcional)* Ligar a coleta periódica (seção *Coleta periódica*):
      criar `COLETA_TOKEN` e os jobs do Cloud Scheduler.
- [ ] **9.** Adicionar o domínio do Cloud Run nos **Authorized domains** do
      Firebase Auth (Console → Authentication → Settings).

Pré-requisitos locais: `curl` e o `gcloud`.

## Setup (uma vez)

Defina o projeto e a região (use a mesma em tudo):

```bash
export PROJECT_ID=seu-projeto
export REGION=southamerica-east1     # menor latência no Brasil
gcloud config set project $PROJECT_ID

# 1) APIs
gcloud services enable run.googleapis.com cloudbuild.googleapis.com \
  artifactregistry.googleapis.com bigquery.googleapis.com secretmanager.googleapis.com

# 2) Artifact Registry (onde a imagem vai morar)
gcloud artifacts repositories create garimpo \
  --repository-format=docker --location=$REGION

# 3) Segredos no Secret Manager (NÃO vão para o GitHub)
printf '%s' 'SEU_APP_ID'  | gcloud secrets create SHOPEE_APP_ID --data-file=-
printf '%s' 'SEU_SECRET'  | gcloud secrets create SHOPEE_SECRET --data-file=-
printf '%s' 'TOKEN_BOT'   | gcloud secrets create TELEGRAM_BOT_TOKEN --data-file=-
printf '%s' 'PRODUCT_ID'  | gcloud secrets create WHATSAPP_PRODUCT_ID --data-file=-
printf '%s' 'PHONE_ID'    | gcloud secrets create WHATSAPP_PHONE_ID --data-file=-
printf '%s' 'API_KEY'     | gcloud secrets create WHATSAPP_API_KEY --data-file=-

# 4) BigQuery: dataset + tabelas
#    edite deploy/bigquery_schema.sql trocando SEU_PROJECT, depois:
bq query --use_legacy_sql=false < deploy/bigquery_schema.sql
```

### Permissões da service account do Cloud Run

A revisão do Cloud Run roda como uma service account (por padrão a Compute
default, ou crie uma dedicada). Ela precisa:

```bash
SA="$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')-compute@developer.gserviceaccount.com"

# escrever no BigQuery
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$SA" --role="roles/bigquery.dataEditor"

# ler os segredos
for SECRET in SHOPEE_APP_ID SHOPEE_SECRET TELEGRAM_BOT_TOKEN COLETA_TOKEN \
              WHATSAPP_PRODUCT_ID WHATSAPP_PHONE_ID WHATSAPP_API_KEY; do
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:$SA" --role="roles/secretmanager.secretAccessor" 2>/dev/null
done
```

## Service account para o CI

Crie uma SA de deploy e dê os papéis necessários; gere a chave JSON:

```bash
gcloud iam service-accounts create gh-deploy --display-name="GitHub Deploy"
DEPLOY_SA="gh-deploy@${PROJECT_ID}.iam.gserviceaccount.com"
for ROLE in run.admin cloudbuild.builds.editor artifactregistry.writer \
            iam.serviceAccountUser storage.admin; do
  gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$DEPLOY_SA" --role="roles/$ROLE"
done
gcloud iam service-accounts keys create key.json --iam-account=$DEPLOY_SA
```

No GitHub, em **Settings → Secrets and variables → Actions**:

| Secret | Valor |
|--------|-------|
| `GCP_PROJECT_ID` | id do projeto |
| `GCP_SA_KEY` | conteúdo do `key.json` |

Apague o `key.json` local depois. (Mais seguro: Workload Identity Federation.)

## Primeiro deploy

`git push` na main dispara o `deploy-gcp.yml`:
1. Testes Go (gate)
2. Build Docker multi-stage (Node build do frontend + Go build da API)
3. Push da imagem para o Artifact Registry
4. Deploy no Cloud Run
5. Atualiza Cloud Scheduler

Acompanhe na aba Actions. Ao fim, abra a URL do Cloud Run.

### Verificação

```bash
RUN_URL=$(gcloud run services describe garimpo-api --region $REGION --format='value(status.url)')

# 1) a API está viva?
curl -s "$RUN_URL/api/health"

# 2) a busca real funciona?
curl -s "$RUN_URL/api/candidatos?fonte=shopee&keyword=perfume&top=3" | head -c 400

# 3) o front carrega?
echo "Abra no navegador: $RUN_URL"
```

Checklist do teste manual:
- A página abre e pede login com Google.
- Após login, a barra de busca aparece.
- Buscar "perfume" traz produtos com o **teor** preenchido.
- **Publicar** mostra a mensagem montada e o destino.
- A aba **Quadro** persiste os cards entre recarregamentos.

Se a busca vier vazia/erro:
```bash
gcloud run services logs read garimpo-api --region $REGION --limit 50
```
Causas comuns: `SHOPEE_*` não acessíveis (papel `secretmanager.secretAccessor`
faltando na SA do Run).

## Coleta periódica (Cloud Scheduler)

O `POST /api/coletar?categoria=...&keyword=...` roda a busca e grava um snapshot
na tabela `snapshots`. O **Cloud Scheduler** chama esse endpoint em cron.

1. Proteja o endpoint com um token:
   ```bash
   printf '%s' "$(openssl rand -hex 24)" | gcloud secrets create COLETA_TOKEN --data-file=-
   gcloud run services update garimpo-api --region $REGION \
     --update-secrets COLETA_TOKEN=COLETA_TOKEN:latest
   ```

2. Crie os jobs (um por busca, com horários espaçados): veja
   `deploy/scheduler-exemplo.sh`. Teste manual:
   ```bash
   TOKEN=$(gcloud secrets versions access latest --secret=COLETA_TOKEN)
   URL=$(gcloud run services describe garimpo-api --region $REGION --format='value(status.url)')
   curl -X POST -H "X-Garimpo-Token: $TOKEN" \
     "$URL/api/coletar?fonte=shopee&categoria=perfumaria&keyword=perfume&top=20&vendas_min=5"
   ```

## Análises

- **Looker Studio** (grátis): conecte ao dataset `garimpo` e monte painéis.
- **Python**: `pip install pandas-gbq` e
  `pandas_gbq.read_gbq("SELECT * FROM garimpo.eventos", project_id="...")`

Exemplo de análise (preço/comissão ao longo do tempo):
```sql
SELECT DATE(coletado_em) dia, AVG(comissao) comissao_media, AVG(preco) preco_medio
FROM `SEU_PROJECT.garimpo.snapshots`
WHERE categoria = 'perfumaria'
GROUP BY dia ORDER BY dia;
```

## Build local com BigQuery

O build padrão é "verde" sem BigQuery. Para compilar com:

```bash
go get cloud.google.com/go/bigquery
GOOGLE_CLOUD_PROJECT=$PROJECT_ID BQ_DATASET=garimpo BQ_TABELA=eventos \
  go run -tags gcp ./cmd/garimpo-api -fonte shopee
```
