"""Estatísticas: resumo de mercado segmentado por fonte (loja vs keyword)."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Estatísticas"])


@router.get("/estatisticas")
def get_estatisticas(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        from mock_data import ESTATISTICAS_RESPONSE
        return {**ESTATISTICAS_RESPONSE, "dias_janela": dias}

    import bq_client
    from google.cloud.bigquery import ScalarQueryParameter

    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    # Query global (backward compat)
    sql = f"""
    WITH ultimos AS (
      SELECT
        produto_id, nome, preco, comissao, vendas, nota,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      COUNT(DISTINCT produto_id) AS total_produtos,
      AVG(preco) AS preco_medio,
      AVG(comissao) AS comissao_media,
      AVG(vendas) AS vendas_media,
      AVG(nota) AS nota_media,
      APPROX_QUANTILES(preco, 4)[OFFSET(2)] AS preco_mediana,
      APPROX_QUANTILES(comissao, 4)[OFFSET(2)] AS comissao_mediana
    FROM ultimos
    WHERE rn = 1
    """

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    if not rows:
        return {
            "dias": dias,
            "total_amostras": 0,
            "resumo": None,
            "por_fonte": _empty_por_fonte(),
        }

    row = rows[0]

    # Query segmentada por fonte (loja- prefix vs keyword)
    sql_fonte = f"""
    WITH ultimos AS (
      SELECT
        produto_id, preco, comissao, keyword,
        CASE WHEN keyword LIKE 'loja-%' THEN 'loja' ELSE 'keyword' END AS fonte,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      fonte,
      COUNT(DISTINCT produto_id) AS total_produtos,
      AVG(preco) AS preco_medio,
      AVG(comissao) AS comissao_media,
      COUNT(DISTINCT keyword) AS total_coletas
    FROM ultimos
    WHERE rn = 1
    GROUP BY fonte
    """

    fonte_rows = bq_client.query(sql_fonte, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    por_fonte = _empty_por_fonte()
    for fr in fonte_rows:
        key = "lojas" if fr["fonte"] == "loja" else "keywords"
        por_fonte[key] = {
            "total_produtos": fr.get("total_produtos", 0),
            "preco_medio": round(fr.get("preco_medio", 0) or 0, 2),
            "comissao_media": round(fr.get("comissao_media", 0) or 0, 4),
            "total_coletas": fr.get("total_coletas", 0),
        }

    return {
        "dias": dias,
        "total_amostras": row.get("total_produtos", 0),
        "resumo": {
            "total_produtos": row.get("total_produtos", 0),
            "preco_medio": round(row.get("preco_medio", 0) or 0, 2),
            "comissao_media": round(row.get("comissao_media", 0) or 0, 4),
            "vendas_media": round(row.get("vendas_media", 0) or 0, 1),
            "nota_media": round(row.get("nota_media", 0) or 0, 2),
            "preco_mediana": row.get("preco_mediana"),
            "comissao_mediana": row.get("comissao_mediana"),
        },
        "por_fonte": por_fonte,
    }


def _empty_por_fonte():
    """Retorna estrutura vazia para por_fonte."""
    empty = {"total_produtos": 0, "preco_medio": 0, "comissao_media": 0, "total_coletas": 0}
    return {"lojas": {**empty}, "keywords": {**empty}}
