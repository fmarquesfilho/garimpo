# ADR 0014 — Serviço Analyzer em Python (FastAPI + BigQuery)

**Status:** aceite  
**Data:** 2026-07-01  

## Contexto

O monólito Go contém lógica analítica (novidades, quedas de preço, evolução,
estatísticas) que consulta BigQuery para detectar padrões em séries temporais de
snapshots. Esta lógica precisa ser migrada para descomissionar o monólito.

Opções avaliadas:
1. Portar para C# com SDK BigQuery — funcional mas perde ecossistema analítico
2. Extrair para microserviço Go — mínimo esforço mas limita evolução IA
3. **Novo serviço Python (FastAPI)** — ecossistema analítico superior, prepara IA

## Decisão

Criar um serviço `analyzer` em Python com FastAPI, responsável por:
- Queries analíticas no BigQuery (novidades, quedas, evolução, estatísticas)
- Detecção de padrões (produtos novos, variações significativas)
- (Futuro) Scoring ML e recomendação personalizada (T-0007)

### Comunicação

REST (FastAPI + Pydantic) em vez de gRPC. Justificativas:
- Ecossistema Python para gRPC é funcional mas verboso
- FastAPI + Pydantic é mais natural e produtivo
- Latência de REST vs gRPC é irrelevante para queries BQ (~200ms)
- OpenAPI gerado automaticamente (Swagger UI embutido)

### Responsabilidades por linguagem

| Stack | Responsabilidade |
|-------|-----------------|
| C# (ASP.NET Core) | CRUD, auth, multi-tenant, orquestração, frontend API |
| Go (gRPC) | I/O intensivo — coleta Shopee, publicação, alertas, scheduling |
| Python (FastAPI) | Analytics, detecção de padrões, IA/ML, queries BigQuery |

## Endpoints do analyzer

```
GET /novidades?busca_id=X&dias=7       → produtos novos + variações de preço
GET /quedas?dias=7&threshold=0.15      → produtos com queda significativa
GET /evolucao?dias=30                  → série temporal de preço por loja
GET /estatisticas?dias=30              → resumo por categoria
GET /health                            → health check
```

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Framework | FastAPI |
| Validação | Pydantic v2 |
| BigQuery | google-cloud-bigquery |
| Dados | pandas (manipulação de DataFrames) |
| Container | Python 3.13 slim (~50MB) |
| Testes | pytest + httpx (TestClient) |

## Consequências

### Se aceitar
- Decomissão do Go fica viável (analytics migra para Python)
- Ecossistema ML/IA pronto (pandas, scikit-learn, etc.)
- Cada linguagem no seu ponto forte
- Terceiro runtime para manter (mitigado: Docker + CI isolado)

### Custo
- +1 linguagem no mono-repo
- +1 sidecar no Cloud Run multi-container
- CI: pytest + ruff/mypy + docker build
