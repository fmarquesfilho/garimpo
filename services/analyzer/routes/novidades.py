"""Novidades: produtos novos e variações de preço detectadas entre snapshots."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Novidades"])


@router.get("/novidades")
def get_novidades(
    busca_id: str = Query("", description="ID da busca/loja"),
    dias: int = Query(7, ge=1, le=90),
):
    if settings.mock_data:
        from mock_data import NOVIDADES_RESPONSE
        return {**NOVIDADES_RESPONSE, "busca_id": busca_id, "dias": dias}

    import bq_client
    """Compara snapshots da janela para detectar produtos novos e variações."""
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    WITH recentes AS (
      SELECT
        produto_id, nome, preco, comissao, vendas, nota, imagem, link, loja,
        DATE(coletado_em) AS dia,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn_recent,
        COUNT(*) OVER (PARTITION BY produto_id) AS aparicoes
      FROM {ds}.snapshots
      WHERE keyword LIKE @busca_id
        AND coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    primeiros AS (
      SELECT
        produto_id, nome, preco AS preco_primeiro, dia AS primeiro_dia
      FROM (
        SELECT *, ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY dia ASC) AS rn
        FROM recentes
      )
      WHERE rn = 1
    ),
    ultimos AS (
      SELECT produto_id, preco AS preco_atual, dia AS ultimo_dia
      FROM recentes
      WHERE rn_recent = 1
    )
    SELECT
      r.produto_id, r.nome, r.preco, r.comissao, r.vendas, r.nota,
      r.imagem, r.link, r.loja, r.aparicoes,
      p.preco_primeiro, u.preco_atual,
      SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) AS variacao
    FROM recentes r
    JOIN primeiros p ON r.produto_id = p.produto_id
    JOIN ultimos u ON r.produto_id = u.produto_id
    WHERE r.rn_recent = 1
    ORDER BY r.aparicoes ASC, variacao ASC
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("busca_id", "STRING", f"%{busca_id}%"),
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    novos = []
    variacoes = []

    for row in rows:
        if row.get("aparicoes", 0) == 1:
            novos.append({
                "produto_id": row["produto_id"],
                "nome": row["nome"],
                "preco": row["preco"],
                "comissao": row.get("comissao"),
                "vendas": row.get("vendas"),
                "nota": row.get("nota"),
                "imagem": row.get("imagem"),
                "link": row.get("link"),
                "loja": row.get("loja"),
            })
        elif row.get("variacao") is not None and abs(row["variacao"]) > 0.01:
            variacoes.append({
                "produto_id": row["produto_id"],
                "nome": row["nome"],
                "preco_anterior": row["preco_primeiro"],
                "preco_atual": row["preco_atual"],
                "variacao": round(row["variacao"], 4),
                "imagem": row.get("imagem"),
                "link": row.get("link"),
                "loja": row.get("loja"),
            })

    return {
        "busca_id": busca_id,
        "dias": dias,
        "produtos_novos": novos,
        "variacoes": variacoes,
        "total_novos": len(novos),
        "total_variacoes": len(variacoes),
    }
