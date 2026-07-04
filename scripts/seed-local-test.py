#!/usr/bin/env python3
"""
seed-local-test.py — Popular BigQuery emulator + PostgreSQL com cenário fictício
para testar o fluxo completo de variação de preços.

Pré-requisitos:
  docker compose up -d postgres bigquery-emulator
  pip3 install google-cloud-bigquery psycopg2-binary

Uso:
  python3 scripts/seed-local-test.py

Cenário criado:
  - 1 busca "perfumes-importados" com shop_id 123456
  - 2 snapshots (3 dias atrás + hoje) com 5 produtos
  - 2 produtos com queda de preço (20% e 35%)
  - 1 produto com alta de preço (10%)
  - 2 produtos sem variação
  - 1 produto novo (só aparece no snapshot de hoje)
"""

import os
import uuid
from datetime import datetime, timedelta, timezone

# ──────────────────────────────────────────────────────────────────────────────
# Config
# ──────────────────────────────────────────────────────────────────────────────

BQ_PROJECT = os.getenv("BQ_PROJECT", "garimpei-dev")
BQ_DATASET = os.getenv("BQ_DATASET", "garimpo")
BQ_EMULATOR_HOST = os.getenv("BIGQUERY_EMULATOR_HOST", "localhost:9050")

PG_HOST = os.getenv("PG_HOST", "localhost")
PG_PORT = os.getenv("PG_PORT", "5432")
PG_DB = os.getenv("PG_DB", "garimpei")
PG_USER = os.getenv("PG_USER", "garimpei")
PG_PASS = os.getenv("PG_PASS", "garimpei_dev")

OWNER_UID = "dev-user-001"
BUSCA_KEYWORD = "perfumes-importados"
BUSCA_ID = str(uuid.UUID("11111111-1111-1111-1111-111111111111"))

# Timestamps
NOW = datetime.now(timezone.utc)
THREE_DAYS_AGO = NOW - timedelta(days=3)

# ──────────────────────────────────────────────────────────────────────────────
# Dados fictícios
# ──────────────────────────────────────────────────────────────────────────────

PRODUTOS_DIA_1 = [
    {
        "produto_id": "SP-001",
        "nome": "Perfume CK One 100ml EDT",
        "preco": 189.90,
        "comissao": 0.12,
        "vendas": 3420,
        "nota": 4.8,
        "imagem": "https://cf.shopee.com.br/file/ck-one-100ml.jpg",
        "link": "https://shopee.com.br/product/123456/SP-001",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-002",
        "nome": "Dolce & Gabbana Light Blue 75ml",
        "preco": 299.00,
        "comissao": 0.10,
        "vendas": 1850,
        "nota": 4.9,
        "imagem": "https://cf.shopee.com.br/file/dg-light-blue.jpg",
        "link": "https://shopee.com.br/product/123456/SP-002",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-003",
        "nome": "Versace Pour Homme 100ml",
        "preco": 245.00,
        "comissao": 0.15,
        "vendas": 980,
        "nota": 4.7,
        "imagem": "https://cf.shopee.com.br/file/versace-pour-homme.jpg",
        "link": "https://shopee.com.br/product/123456/SP-003",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-004",
        "nome": "Carolina Herrera Good Girl 80ml",
        "preco": 420.00,
        "comissao": 0.08,
        "vendas": 2100,
        "nota": 4.9,
        "imagem": "https://cf.shopee.com.br/file/ch-good-girl.jpg",
        "link": "https://shopee.com.br/product/123456/SP-004",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-005",
        "nome": "Paco Rabanne 1 Million 200ml",
        "preco": 350.00,
        "comissao": 0.11,
        "vendas": 5600,
        "nota": 4.6,
        "imagem": "https://cf.shopee.com.br/file/paco-1-million.jpg",
        "link": "https://shopee.com.br/product/123456/SP-005",
        "loja": "ImportsPerfumaria",
    },
]

# Dia 2 (hoje): variações de preço + produto novo
PRODUTOS_DIA_2 = [
    {
        "produto_id": "SP-001",
        "nome": "Perfume CK One 100ml EDT",
        "preco": 151.90,  # ← QUEDA de ~20%
        "comissao": 0.12,
        "vendas": 3520,
        "nota": 4.8,
        "imagem": "https://cf.shopee.com.br/file/ck-one-100ml.jpg",
        "link": "https://shopee.com.br/product/123456/SP-001",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-002",
        "nome": "Dolce & Gabbana Light Blue 75ml",
        "preco": 194.00,  # ← QUEDA de ~35%
        "comissao": 0.10,
        "vendas": 1920,
        "nota": 4.9,
        "imagem": "https://cf.shopee.com.br/file/dg-light-blue.jpg",
        "link": "https://shopee.com.br/product/123456/SP-002",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-003",
        "nome": "Versace Pour Homme 100ml",
        "preco": 245.00,  # sem variação
        "comissao": 0.15,
        "vendas": 1010,
        "nota": 4.7,
        "imagem": "https://cf.shopee.com.br/file/versace-pour-homme.jpg",
        "link": "https://shopee.com.br/product/123456/SP-003",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-004",
        "nome": "Carolina Herrera Good Girl 80ml",
        "preco": 462.00,  # ← ALTA de ~10%
        "comissao": 0.08,
        "vendas": 2150,
        "nota": 4.9,
        "imagem": "https://cf.shopee.com.br/file/ch-good-girl.jpg",
        "link": "https://shopee.com.br/product/123456/SP-004",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-005",
        "nome": "Paco Rabanne 1 Million 200ml",
        "preco": 350.00,  # sem variação
        "comissao": 0.11,
        "vendas": 5700,
        "nota": 4.6,
        "imagem": "https://cf.shopee.com.br/file/paco-1-million.jpg",
        "link": "https://shopee.com.br/product/123456/SP-005",
        "loja": "ImportsPerfumaria",
    },
    {
        "produto_id": "SP-006",
        "nome": "Jean Paul Gaultier Le Male 125ml",  # ← PRODUTO NOVO
        "preco": 280.00,
        "comissao": 0.13,
        "vendas": 450,
        "nota": 4.5,
        "imagem": "https://cf.shopee.com.br/file/jpg-le-male.jpg",
        "link": "https://shopee.com.br/product/123456/SP-006",
        "loja": "ImportsPerfumaria",
    },
]


def seed_bigquery():
    """Cria tabela e insere snapshots no BigQuery emulator."""
    print("🔧 Configurando BigQuery emulator...")
    os.environ["BIGQUERY_EMULATOR_HOST"] = BQ_EMULATOR_HOST

    from google.cloud import bigquery

    client = bigquery.Client(project=BQ_PROJECT)

    # Criar dataset (pode já existir no emulator via --dataset flag)
    dataset_ref = f"{BQ_PROJECT}.{BQ_DATASET}"
    try:
        client.create_dataset(dataset_ref, exists_ok=True)
        print(f"  ✅ Dataset {dataset_ref} pronto")
    except Exception as e:
        print(f"  ⚠️  Dataset (pode já existir): {e}")

    # Criar tabela snapshots com schema que o analyzer espera
    table_ref = f"{dataset_ref}.snapshots"

    schema = [
        bigquery.SchemaField("coletado_em", "TIMESTAMP"),
        bigquery.SchemaField("keyword", "STRING"),
        bigquery.SchemaField("produto_id", "STRING"),
        bigquery.SchemaField("nome", "STRING"),
        bigquery.SchemaField("preco", "FLOAT64"),
        bigquery.SchemaField("comissao", "FLOAT64"),
        bigquery.SchemaField("vendas", "INT64"),
        bigquery.SchemaField("nota", "FLOAT64"),
        bigquery.SchemaField("imagem", "STRING"),
        bigquery.SchemaField("link", "STRING"),
        bigquery.SchemaField("loja", "STRING"),
    ]

    table = bigquery.Table(table_ref, schema=schema)
    try:
        client.delete_table(table_ref, not_found_ok=True)
        client.create_table(table)
        print(f"  ✅ Tabela {table_ref} criada")
    except Exception as e:
        print(f"  ❌ Erro criando tabela: {e}")
        raise

    # Inserir snapshots dia 1 (3 dias atrás)
    rows_dia_1 = [
        {
            "coletado_em": THREE_DAYS_AGO.isoformat(),
            "keyword": BUSCA_KEYWORD,
            **p,
        }
        for p in PRODUTOS_DIA_1
    ]

    # Inserir snapshots dia 2 (hoje)
    rows_dia_2 = [
        {
            "coletado_em": NOW.isoformat(),
            "keyword": BUSCA_KEYWORD,
            **p,
        }
        for p in PRODUTOS_DIA_2
    ]

    all_rows = rows_dia_1 + rows_dia_2
    errors = client.insert_rows_json(table_ref, all_rows)
    if errors:
        print(f"  ❌ Erros ao inserir: {errors}")
        raise RuntimeError(f"BigQuery insert errors: {errors}")

    print(f"  ✅ {len(rows_dia_1)} snapshots (dia 1: {THREE_DAYS_AGO.date()})")
    print(f"  ✅ {len(rows_dia_2)} snapshots (dia 2: {NOW.date()})")
    print(f"  📊 Total: {len(all_rows)} rows inseridas")


def seed_postgresql():
    """Cria uma busca no PostgreSQL para o cenário."""
    print("\n🔧 Configurando PostgreSQL...")

    try:
        import psycopg2
    except ImportError:
        print("  ⚠️  psycopg2 não instalado. Instale com: pip3 install psycopg2-binary")
        print("  ⏭️  Pulando seed PostgreSQL (pode criar via API depois)")
        return

    conn = psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        dbname=PG_DB,
        user=PG_USER,
        password=PG_PASS,
    )
    conn.autocommit = True
    cur = conn.cursor()

    # Inserir busca (upsert)
    cur.execute("""
        INSERT INTO "Buscas" ("Id", "Keyword", "SortBy", "Limit", "Active", "OwnerUid", "CreatedAt", "UpdatedAt", "Marketplaces")
        VALUES (%s, %s, 'relevance', 50, TRUE, %s, NOW(), NOW(), '["shopee"]')
        ON CONFLICT ("Id") DO UPDATE SET "Active" = TRUE, "UpdatedAt" = NOW()
    """, (BUSCA_ID, BUSCA_KEYWORD, OWNER_UID))

    print(f"  ✅ Busca criada: {BUSCA_KEYWORD} (id={BUSCA_ID})")
    print(f"  👤 Owner: {OWNER_UID}")

    cur.close()
    conn.close()


def print_summary():
    """Imprime resumo do cenário e como testar."""
    print("\n" + "=" * 70)
    print("🎯 CENÁRIO DE TESTE PRONTO")
    print("=" * 70)
    print(f"""
Busca: "{BUSCA_KEYWORD}" (ID: {BUSCA_ID})
Owner: {OWNER_UID}

📦 Produtos no cenário:
  SP-001  CK One 100ml         R$189,90 → R$151,90  📉 -20%
  SP-002  D&G Light Blue 75ml  R$299,00 → R$194,00  📉 -35%
  SP-003  Versace Pour Homme   R$245,00 → R$245,00  ─── 0%
  SP-004  CH Good Girl 80ml    R$420,00 → R$462,00  📈 +10%
  SP-005  Paco 1 Million 200ml R$350,00 → R$350,00  ─── 0%
  SP-006  JPG Le Male 125ml    (novo)   → R$280,00  🆕 novo

🧪 Como testar:

  1. Subir os serviços:
     docker compose up -d postgres bigquery-emulator analyzer

  2. Testar analyzer diretamente:
     curl http://localhost:8060/health
     curl "http://localhost:8060/novidades?busca_id={BUSCA_KEYWORD}&dias=7"
     curl "http://localhost:8060/quedas?dias=7&threshold=0.15"
     curl "http://localhost:8060/coletas?dias=30"

  3. Testar via API C# (requer api rodando):
     curl http://localhost:8090/api/lojas
     curl "http://localhost:8090/api/lojas/novidades?busca_id={BUSCA_KEYWORD}&dias=7"

  4. Testar via frontend:
     cd web && VITE_API_BASE=http://localhost:8090 npm run dev
     → Navegar para /lojas → aba "📉 Preços"
""")


if __name__ == "__main__":
    print("🌱 Seed de cenário local — Fluxo de Variação de Preços\n")

    seed_bigquery()
    seed_postgresql()
    print_summary()
