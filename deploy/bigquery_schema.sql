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
--     é o último registro por id (ativo = TRUE). Alimenta a coleta agendada.
--     Uma busca pode ter várias keywords (armazenadas como JSON array em `keywords`).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.buscas` (
  id           STRING,       -- slug da keyword principal (ex.: "perfumaria-japonesa")
  keywords     STRING,       -- JSON array de termos (ex.: '["kenzo","shiseido"]')
  shop_ids     STRING,       -- JSON array de IDs de lojas (ex.: '[12345,67890]')
  categoria    STRING,
  estrategia   STRING,       -- "nicho" | "diversificada" | "ambas"
  comissao_min FLOAT64,
  vendas_min   INT64,
  nota_min     FLOAT64,
  top          INT64,
  cron         STRING,       -- periodicidade da coleta (ex.: "0 8 * * *")
  ativo        BOOL,         -- FALSE = removida (tombstone)
  owner_uid    STRING,       -- uid do Firebase Auth
  rotation_cursor STRING,    -- JSON map shopID→próxima página (rotação de catálogo)
  full_scan_at STRING,       -- JSON map shopID→timestamp da última varredura completa
  salvo_em     TIMESTAMP
)
PARTITION BY DATE(salvo_em);

-- 3) (Opcional, Incremento futuro) Conversões vindas do conversionReport da Shopee.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.conversoes` (
  conversion_id   STRING,
  produto_id      STRING,
  nome            STRING,
  canal           STRING,
  estrategia      STRING,   -- recuperada do subId/utmContent
  comissao_total  FLOAT64,
  preco           FLOAT64,
  status          STRING,   -- PENDING | COMPLETED | CANCELLED
  clique_em       TIMESTAMP,
  compra_em       TIMESTAMP,
  convertido_em   TIMESTAMP
)
PARTITION BY DATE(compra_em);

-- 4) Destinos de publicação (append-only, versionado por salvo_em).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.destinos` (
  id        STRING,
  nome      STRING,
  tipo      STRING,       -- "telegram" | "whatsapp"
  config    STRING,       -- chat_id, telefone, etc.
  ativo     BOOL,
  salvo_em  TIMESTAMP
)
PARTITION BY DATE(salvo_em);

-- 5) Templates de mensagem (append-only).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.templates` (
  id        STRING,
  nome      STRING,
  corpo     STRING,       -- corpo com placeholders {{nome}}, {{preco}}, etc.
  com_foto  BOOL,
  ativo     BOOL,
  salvo_em  TIMESTAMP
)
PARTITION BY DATE(salvo_em);

-- 6) Publicações agendadas/enviadas.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.publicacoes` (
  id           STRING,
  produto_id   STRING,
  nome         STRING,
  categoria    STRING,
  preco        FLOAT64,
  comissao     FLOAT64,
  link         STRING,
  imagem       STRING,
  estrategia   STRING,
  destino_id   STRING,
  template_id  STRING,
  agendada_em  STRING,
  status       STRING,      -- "pendente" | "enviada" | "erro"
  detalhe      STRING,
  criada_em    TIMESTAMP,
  enviada_em   STRING,
  owner_uid    STRING
)
PARTITION BY DATE(criada_em);

-- 7) Favoritos (append-only).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.favoritos` (
  produto_id  STRING,
  nome        STRING,
  preco       FLOAT64,
  comissao    FLOAT64,
  link        STRING,
  imagem      STRING,
  loja        STRING,
  categoria   STRING,
  origem      STRING,
  ativo       BOOL,
  owner_uid   STRING,
  salvo_em    TIMESTAMP
)
PARTITION BY DATE(salvo_em);

-- Exemplo de análise: seleções por estratégia na última semana.
--   SELECT estrategia, COUNT(*) selecoes, AVG(comissao) comissao_media, AVG(score) teor_medio
--   FROM `SEU_PROJECT.garimpo.eventos`
--   WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
--   GROUP BY estrategia;

-- 8) Coupon snapshots: append-only log of collected coupons.
-- Partitioned by collected_at for cost-efficient time-range queries.
-- Partition expiration: 90 days (automatic cleanup by BigQuery).
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.coupon_snapshots` (
  coupon_id             STRING    NOT NULL,
  marketplace           STRING    NOT NULL,
  code                  STRING,
  discount_type         STRING    NOT NULL,
  discount_value        FLOAT64   NOT NULL,
  min_spend             FLOAT64,
  start_time            TIMESTAMP,
  end_time              TIMESTAMP,
  applicable_categories STRING,
  status                STRING    NOT NULL,
  detection_status      STRING,
  owner_uid             STRING    NOT NULL,
  collected_at          TIMESTAMP NOT NULL
)
PARTITION BY DATE(collected_at)
OPTIONS (
  partition_expiration_days = 90,
  description = "Append-only coupon snapshots for detection and analytics"
);
