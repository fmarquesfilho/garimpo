# Deploy na GCP (Cloud Run + BigQuery + Firebase Hosting)

Arquitetura analytics-first: a API Go roda no **Cloud Run**, os dados minerados
e as decisões de curadoria vão para o **BigQuery** (análise barata + Looker
Studio), e o front estático fica no **Firebase Hosting**, que faz rewrite de
`/api` para o Cloud Run — uma origem só, sem CORS.

```
navegador ──https──> Firebase Hosting ──┬─ /        -> site estático (web/build)
                                         └─ /api/**  -> Cloud Run (garimpo-api)
                                                          ├─ Secret Manager: SHOPEE_*
                                                          └─ grava eventos -> BigQuery
                                                                                  └─ Looker Studio / Python
```

Custo no seu volume: Cloud Run praticamente de graça (escala a zero), BigQuery
dentro do free tier (10 GB de storage + 1 TB de consulta/mês), Firebase Hosting
free tier. Deve ficar bem abaixo de R$50 — confirme no Billing.

## RUNBOOK: do zero ao ar (checklist ordenado)

Faça na ordem. Os detalhes de cada passo estão nas seções abaixo.

- [ ] **1.** Criar projeto GCP e instalar o `gcloud` (e `bq`). `gcloud auth login`.
- [ ] **2.** Definir `PROJECT_ID` e `REGION` e rodar o bloco de **Setup** (habilita
      APIs, cria Artifact Registry, cria os segredos `SHOPEE_*`, cria o dataset e
      as tabelas do BigQuery).
- [ ] **3.** Dar à service account do Cloud Run os papéis de BigQuery + Secret
      Manager (seção *Permissões*).
- [ ] **4.** Editar `.firebaserc` (`SEU_PROJECT_ID`) e `firebase.json` (trocar
      `REGIAO` pela sua `REGION`). Associar o projeto ao Firebase Hosting.
- [ ] **5.** Criar a service account de deploy do CI e gerar `key.json`
      (seção *Service account para o CI*).
- [ ] **6.** No GitHub, criar os secrets `GCP_PROJECT_ID` e `GCP_SA_KEY`.
      (Os segredos da Shopee **não** vão para o GitHub.)
- [ ] **7.** `git push` na `main` → o workflow `deploy-gcp.yml` testa, builda a
      imagem, sobe no Cloud Run e publica o front. Acompanhe na aba **Actions**.
- [ ] **8.** **Testar** (seção *Primeiro deploy → verificação*): abrir a URL do
      Hosting, conferir `/api/health`, fazer uma busca e uma publicação.
- [ ] **9.** *(Opcional)* Ligar a coleta periódica (seção *Coleta periódica*):
      criar `COLETA_TOKEN` e os jobs do Cloud Scheduler.

Pré-requisitos locais para o passo 8 de teste manual: `curl` e o `gcloud`.

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

# 3) Segredos da Shopee no Secret Manager (NÃO vão para o GitHub)
printf '%s' 'SEU_APP_ID'  | gcloud secrets create SHOPEE_APP_ID --data-file=-
printf '%s' 'SEU_SECRET'  | gcloud secrets create SHOPEE_SECRET --data-file=-

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
gcloud secrets add-iam-policy-binding SHOPEE_APP_ID --member="serviceAccount:$SA" --role="roles/secretmanager.secretAccessor"
gcloud secrets add-iam-policy-binding SHOPEE_SECRET --member="serviceAccount:$SA" --role="roles/secretmanager.secretAccessor"
```

### Firebase Hosting

```bash
# associe o projeto ao Firebase (se ainda não for) e edite os arquivos:
#  - .firebaserc        -> troque SEU_PROJECT_ID
#  - firebase.json      -> troque "REGIAO" pela sua REGION
```

> **Sobre o rewrite Hosting → Cloud Run:** nem toda região é suportada para esse
> rewrite. Se a sua não for, há plano B: em vez do rewrite, aponte o front
> direto para a URL do Cloud Run definindo `VITE_API_BASE` no build (o CORS já
> está liberado na API). Aí o `firebase.json` mantém só o fallback SPA.

## Service account para o CI

Crie uma SA de deploy e dê os papéis necessários; gere a chave JSON:

```bash
gcloud iam service-accounts create gh-deploy --display-name="GitHub Deploy"
DEPLOY_SA="gh-deploy@${PROJECT_ID}.iam.gserviceaccount.com"
for ROLE in run.admin cloudbuild.builds.editor artifactregistry.writer \
            firebasehosting.admin iam.serviceAccountUser storage.admin; do
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

Apague o `key.json` local depois. (Mais seguro ainda, e sem chave: configurar
Workload Identity Federation — fica como evolução.)

## Primeiro deploy

`git push` na main dispara o `deploy-gcp.yml`: testa, builda a imagem (com a tag
`gcp`, que inclui a gravação no BigQuery), sobe no Cloud Run e publica o front.
Acompanhe na aba Actions. Ao fim, abra a URL do Firebase Hosting.

### Verificação (passo 8 do runbook)

```bash
# URLs
HOSTING="https://$PROJECT_ID.web.app"        # ou o domínio do Firebase Hosting
RUN_URL=$(gcloud run services describe garimpo-api --region $REGION --format='value(status.url)')

# 1) a API está viva?
curl -s "$RUN_URL/api/health"                 # espera {"status":"ok"} (ou similar)

# 2) a busca real funciona? (precisa dos SHOPEE_* no Secret Manager)
curl -s "$RUN_URL/api/candidatos?fonte=shopee&keyword=perfume&top=3" | head -c 400

# 3) o front carrega e fala com a API?
echo "Abra no navegador: $HOSTING  -> busque 'perfume' -> clique Publicar"
```

Checklist do teste manual no navegador:
- A página abre e a barra de filtros aparece (busca, comissão, vendas, nota, explorar).
- Buscar "perfume" traz produtos com o **teor** preenchido; o botão **?** explica o teor.
- Produtos de comissão alta sem vendas aparecem com o selo **⚠ suspeito**.
- **Publicar** mostra a mensagem montada e o **sub_id** de atribuição.
- A aba **Quadro** persiste os cards entre recarregamentos.

Se a página subir mas a busca vier vazia/erro:
```bash
gcloud run services logs read garimpo-api --region $REGION --limit 50
```
Causas comuns: `SHOPEE_*` não acessíveis (papel `secretmanager.secretAccessor`
faltando na SA do Run), ou o rewrite Hosting→Run na região não suportado — nesse
caso use o plano B do `VITE_API_BASE` (ver nota na seção Setup → Firebase).

## Coleta periódica (Cloud Scheduler)

Para acumular a série temporal que sustenta a análise de impacto, o
`POST /api/coletar?categoria=...&keyword=...` roda a busca e grava um snapshot
(top N do momento) na tabela `snapshots`. O **Cloud Scheduler** chama esse
endpoint em cron — sem infra a mais.

1. Proteja o endpoint com um token (senão qualquer um dispara coletas e queima
   seu rate limit da Shopee):
   ```bash
   printf '%s' "$(openssl rand -hex 24)" | gcloud secrets create COLETA_TOKEN --data-file=-
   gcloud run services update garimpo-api --region $REGION \
     --update-secrets COLETA_TOKEN=COLETA_TOKEN:latest
   ```
   Com `COLETA_TOKEN` definido, a API exige o header `X-Garimpo-Token`. Sem ele
   (local/dev), o endpoint fica liberado.

2. Crie os jobs (um por categoria, com horários espaçados): veja
   `deploy/scheduler-exemplo.sh`. Teste manual:
   ```bash
   TOKEN=$(gcloud secrets versions access latest --secret=COLETA_TOKEN)
   URL=$(gcloud run services describe garimpo-api --region $REGION --format='value(status.url)')
   curl -X POST -H "X-Garimpo-Token: $TOKEN" \
     "$URL/api/coletar?fonte=shopee&categoria=perfumaria&keyword=perfume&top=20&vendas_min=5"
   ```

Exemplo de análise de impacto (preço/comissão de um produto ao longo do tempo):
```sql
SELECT DATE(coletado_em) dia, AVG(comissao) comissao_media, AVG(preco) preco_medio
FROM `SEU_PROJECT.garimpo.snapshots`
WHERE categoria = 'perfumaria'
GROUP BY dia ORDER BY dia;
```

## Análises

- **Looker Studio** (grátis): conecte ao dataset `garimpo` e monte os painéis
  (seleções por estratégia, comissão média, "teor" médio, evolução semanal).
- **Python**: `pip install pandas-gbq` e
  `pandas_gbq.read_gbq("SELECT * FROM garimpo.eventos", project_id="...")`
  para análises mais pesadas no seu ambiente.
- Quando o Incremento 3 (atribuição via `conversionReport`) chegar, a tabela
  `conversoes` fecha o laço: dá para cruzar seleção → venda e medir, enfim,
  receita por estratégia e por hora de esforço.

## Build local com BigQuery

O build padrão (e o CI) é "verde" sem a dependência do BigQuery. Para compilar a
versão que grava no BigQuery:

```bash
go get cloud.google.com/go/bigquery
GOOGLE_CLOUD_PROJECT=$PROJECT_ID BQ_DATASET=garimpo BQ_TABELA=eventos \
  go run -tags gcp ./cmd/garimpo-api -fonte shopee
```
