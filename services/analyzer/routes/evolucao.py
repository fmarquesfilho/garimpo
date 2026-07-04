"""Evolução: série temporal de preço médio por loja."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Evolução"])


@router.get("/evolucao")
def get_evolucao(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        from mock_data import EVOLUCAO_RESPONSE
        return {**EVOLUCAO_RESPONSE, "dias_janela": dias}

    """Série temporal de preço médio por dia, com resumo global e top variações."""
    import bq_client
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    SELECT
      DATE(coletado_em) AS dia,
      keyword AS loja,
      AVG(preco) AS preco_medio,
      COUNT(DISTINCT produto_id) AS produtos,
      MIN(preco) AS preco_min,
      MAX(preco) AS preco_max
    FROM {ds}.snapshots
    WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    GROUP BY dia, loja
    ORDER BY dia DESC, loja
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    # Agrupar por loja
    lojas: dict = {}
    for row in rows:
        loja = row["loja"]
        if loja not in lojas:
            lojas[loja] = []
        lojas[loja].append({
            "dia": str(row["dia"]),
            "preco_medio": round(row["preco_medio"], 2),
            "produtos": row["produtos"],
            "preco_min": row["preco_min"],
            "preco_max": row["preco_max"],
        })

    # Calcular resumo com contagem de quedas e altas
    resumo_sql = f"""
    WITH primeiros AS (
      SELECT produto_id, preco AS preco_primeiro,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    ultimos AS (
      SELECT produto_id, preco AS preco_atual,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      COUNTIF(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) < -0.01) AS total_quedas,
      COUNTIF(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) > 0.01) AS total_altas,
      COUNT(*) AS total_produtos,
      AVG(u.preco_atual) AS preco_medio_global,
      AVG(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro)) AS variacao_media_global_pct
    FROM primeiros p
    JOIN ultimos u ON p.produto_id = u.produto_id AND u.rn = 1
    WHERE p.rn = 1 AND p.preco_primeiro > 0
    """

    resumo_rows = bq_client.query(resumo_sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    resumo = {}
    if resumo_rows:
        r = resumo_rows[0]
        resumo = {
            "total_lojas": len(lojas),
            "total_produtos": r.get("total_produtos", 0),
            "preco_medio_global": round(r.get("preco_medio_global", 0) or 0, 2),
            "variacao_media_global_pct": round(r.get("variacao_media_global_pct", 0) or 0, 4),
            "total_quedas": r.get("total_quedas", 0),
            "total_altas": r.get("total_altas", 0),
        }

    # Formatar lojas como lista com variação média
    lojas_lista = []
    for loja_nome, serie in lojas.items():
        pontos = sorted(serie, key=lambda x: x["dia"])
        variacao_media = 0.0
        if len(pontos) >= 2 and pontos[0]["preco_medio"] > 0:
            variacao_media = (pontos[-1]["preco_medio"] - pontos[0]["preco_medio"]) / pontos[0]["preco_medio"]
        lojas_lista.append({
            "busca_id": loja_nome,
            "produtos": pontos[-1]["produtos"] if pontos else 0,
            "preco_medio": pontos[-1]["preco_medio"] if pontos else 0,
            "variacao_media_pct": round(variacao_media, 4),
            "pontos": [{"data": p["dia"], "preco_medio": p["preco_medio"]} for p in pontos],
        })

    return {
        "dias": dias,
        "lojas": lojas_lista,
        "resumo": resumo,
        "total_lojas": len(lojas),
    }
