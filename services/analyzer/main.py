"""Garimpei Analyzer — FastAPI service for analytics queries on BigQuery."""

from fastapi import FastAPI

from config import settings
from routes import novidades, quedas, evolucao, estatisticas, coletas, conversoes, cupons

# Initialize OTel AFTER imports (ruff E402 compliance).
# FastAPIInstrumentor.instrument() patches globally — works even after app creation.
from otel_setup import init_otel
init_otel("analyzer")

app = FastAPI(
    title="Garimpei Analyzer",
    description="Analytics service — novidades, quedas, evolução, estatísticas, coletas, conversões, cupons",
    version="1.0.0",
)

app.include_router(novidades.router)
app.include_router(quedas.router)
app.include_router(evolucao.router)
app.include_router(estatisticas.router)
app.include_router(coletas.router)
app.include_router(conversoes.router)
app.include_router(cupons.router)


@app.get("/health")
def health():
    return {
        "status": "ok",
        "service": "analyzer",
        "bq_project": settings.bq_project or "(not configured)",
        "bq_dataset": settings.bq_dataset,
    }
