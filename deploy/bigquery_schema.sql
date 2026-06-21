-- Schema do armazém analítico do Garimpo (BigQuery).
-- Rode no console do BigQuery ou via `bq query --use_legacy_sql=false < este.sql`.
-- Troque SEU_PROJECT pelo id do projeto. Dataset sugerido na região da sua VM/Run
-- (ex.: southamerica-east1 para menor latência no Brasil).

-- 1) Dataset
CREATE SCHEMA IF NOT EXISTS `SEU_PROJECT.garimpo`
  OPTIONS (location = 'southamerica-east1');

-- 2) Eventos de curadoria (o que a API grava quando ela seleciona um produto).
--    Particionado por dia para consultas baratas e rápidas.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.eventos` (
  tipo        STRING,      -- "selecao" | "publicacao"
  produto_id  STRING,
  nome        STRING,
  categoria   STRING,
  estrategia  STRING,      -- "nicho" | "diversificada"
  canal       STRING,      -- preenchido em publicações: "telegram", ...
  sub_id      STRING,      -- atribuição: canal_estrategia_AAAAMMDD (conversionReport)
  comissao    FLOAT64,     -- fração (0.15 = 15%)
  preco       FLOAT64,
  vendas      INT64,
  score       FLOAT64,     -- o "teor"
  em          TIMESTAMP
)
PARTITION BY DATE(em);

-- 2b) Snapshots de mercado: a foto periódica dos top N de uma categoria.
--     É o pano de fundo para medir impacto de campanha (o mercado mudou em volta?).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.snapshots` (
  coletado_em  TIMESTAMP,
  categoria    STRING,
  keyword      STRING,
  estrategia   STRING,
  posicao      INT64,       -- ranking no dia (1 = topo)
  produto_id   STRING,
  nome         STRING,
  preco        FLOAT64,
  comissao     FLOAT64,
  vendas       INT64,
  nota         FLOAT64,
  score        FLOAT64
)
PARTITION BY DATE(coletado_em);

-- 2c) Buscas salvas (perfis de coleta), append-only/versionadas. O estado atual
--     é o último registro por nome (ativo = TRUE). Alimenta a coleta agendada.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.buscas` (
  nome         STRING,
  keyword      STRING,
  categoria    STRING,
  estrategia   STRING,
  comissao_min FLOAT64,
  vendas_min   INT64,
  nota_min     FLOAT64,
  top          INT64,
  cron         STRING,       -- periodicidade da coleta (ex.: "0 8 * * *")
  ativo        BOOL,         -- FALSE = removida (tombstone)
  salvo_em     TIMESTAMP
)
PARTITION BY DATE(salvo_em);

-- 3) (Opcional, Incremento futuro) Conversões vindas do conversionReport da Shopee.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.conversoes` (
  conversion_id   STRING,
  produto_id      STRING,
  estrategia      STRING,   -- recuperada do subId/utmContent
  comissao_total  FLOAT64,
  status          STRING,   -- PENDING | COMPLETED | CANCELLED
  clique_em       TIMESTAMP,
  compra_em       TIMESTAMP
)
PARTITION BY DATE(compra_em);

-- Exemplo de análise: seleções por estratégia na última semana.
--   SELECT estrategia, COUNT(*) selecoes, AVG(comissao) comissao_media, AVG(score) teor_medio
--   FROM `SEU_PROJECT.garimpo.eventos`
--   WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
--   GROUP BY estrategia;
