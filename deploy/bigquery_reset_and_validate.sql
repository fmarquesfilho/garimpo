-- =============================================================================
-- BigQuery: Validação de Schema + Reset para nova fase de testes
-- Projeto: garimpo-500114 | Dataset: garimpo
-- =============================================================================
-- EXECUTAR NO CONSOLE DO BIGQUERY OU VIA:
--   bq query --use_legacy_sql=false --project_id=garimpo-500114 < deploy/bigquery_reset_and_validate.sql
-- =============================================================================

-- ═══════════════════════════════════════════════════════════════════════════════
-- PARTE 1: TRUNCAR DADOS ANTIGOS (reset para fase nova de testes)
-- ═══════════════════════════════════════════════════════════════════════════════

-- Snapshots: dados corrompidos pelo monólito Go (bug de IDs inconsistentes)
TRUNCATE TABLE `garimpo-500114.garimpo.snapshots`;

-- Eventos: limpar histórico antigo para recomeçar
TRUNCATE TABLE `garimpo-500114.garimpo.eventos`;

-- Buscas: limpar — agora gerenciadas pelo PostgreSQL (C# API)
TRUNCATE TABLE `garimpo-500114.garimpo.buscas`;

-- Publicações: limpar — agora gerenciadas pelo PostgreSQL
TRUNCATE TABLE `garimpo-500114.garimpo.publicacoes`;

-- Destinos: limpar — agora gerenciados pelo PostgreSQL
TRUNCATE TABLE `garimpo-500114.garimpo.destinos`;

-- Templates: limpar — agora gerenciados pelo PostgreSQL
TRUNCATE TABLE `garimpo-500114.garimpo.templates`;

-- Favoritos: limpar — agora gerenciados pelo PostgreSQL
TRUNCATE TABLE `garimpo-500114.garimpo.favoritos`;

-- Conversões: MANTER (dados reais da Shopee, se houver)
-- TRUNCATE TABLE `garimpo-500114.garimpo.conversoes`;

-- ═══════════════════════════════════════════════════════════════════════════════
-- PARTE 2: VALIDAR SCHEMA (criar colunas faltantes)
-- ═══════════════════════════════════════════════════════════════════════════════

-- A tabela snapshots precisa dos campos imagem, link e loja (adicionados pelo Go
-- EnsureSchema, mas ausentes no SQL original). Caso não existam:
-- BigQuery não suporta ALTER TABLE ADD COLUMN IF NOT EXISTS, então usamos
-- um CTAS se necessário. Na prática, o Go EnsureSchema cuida disso.
-- Mas aqui documentamos o schema esperado para referência:

-- Schema esperado: snapshots
-- coletado_em  TIMESTAMP (partition key)
-- categoria    STRING
-- keyword      STRING
-- estrategia   STRING
-- posicao      INT64
-- produto_id   STRING
-- nome         STRING
-- preco        FLOAT64
-- comissao     FLOAT64
-- vendas       INT64
-- nota         FLOAT64
-- score        FLOAT64
-- imagem       STRING   ← adicionado (evolução)
-- link         STRING   ← adicionado (evolução)
-- loja         STRING   ← adicionado (evolução)

-- ═══════════════════════════════════════════════════════════════════════════════
-- PARTE 3: VALIDAÇÃO — queries de smoke test (rodar após primeira coleta)
-- ═══════════════════════════════════════════════════════════════════════════════

-- [VALIDAR] Após primeira coleta: existem snapshots com produto_id repetido?
-- Se esta query retornar linhas com aparicoes > 1, o pipeline está correto.
-- SELECT produto_id, COUNT(*) AS aparicoes
-- FROM `garimpo-500114.garimpo.snapshots`
-- WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
-- GROUP BY produto_id
-- HAVING aparicoes > 1
-- LIMIT 10;

-- [VALIDAR] Variações de preço detectáveis?
-- SELECT a.produto_id, a.preco AS preco_antigo, b.preco AS preco_novo,
--        (b.preco - a.preco) / a.preco AS variacao_pct
-- FROM `garimpo-500114.garimpo.snapshots` a
-- JOIN `garimpo-500114.garimpo.snapshots` b
--   ON a.produto_id = b.produto_id AND b.coletado_em > a.coletado_em
-- WHERE a.preco > 0 AND b.preco > 0 AND a.preco != b.preco
-- LIMIT 10;
