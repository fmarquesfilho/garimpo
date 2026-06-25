-- Migração: adiciona colunas de monitoramento de lojas à tabela buscas.
-- BigQuery não suporta múltiplos statements num único bq query.
-- Rode cada ALTER separadamente:
--
--   bq query --use_legacy_sql=false 'ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS shop_ids STRING'
--   bq query --use_legacy_sql=false 'ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS owner_uid STRING'
--   bq query --use_legacy_sql=false 'ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS rotation_cursor STRING'
--   bq query --use_legacy_sql=false 'ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS full_scan_at STRING'
--
-- Status: EXECUTADO em 25/06/2026.

ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS shop_ids STRING;
ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS owner_uid STRING;
ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS rotation_cursor STRING;
ALTER TABLE `garimpo-500114.garimpo.buscas` ADD COLUMN IF NOT EXISTS full_scan_at STRING;
