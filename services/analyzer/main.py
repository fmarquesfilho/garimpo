"""Garimpei Analyzer — FastAPI service for analytics queries on BigQuery."""

from fastapi import FastAPI

from config import settings
from routes import novidades, quedas, evolucao, estatisticas

app = FastAPI(
    title="Garimpei Analyzer",
    description="Analytics service — novidades, quedas, evolução, estatísticas",
    version="1.0.0",
)

app.include_router(novidades.router)
app.include_router(quedas.router)
app.include_router(evolucao.router)
app.include_router(estatisticas.router)


@app.get("/health")
def health():
    return {
        "status": "ok",
        "service": "analyzer",
        "bq_project": settings.bq_project or "(not configured)",
        "bq_dataset": settings.bq_dataset,
    }
