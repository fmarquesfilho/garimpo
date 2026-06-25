-- Migração: adiciona colunas de monitoramento de lojas à tabela buscas.
-- Rode uma vez no console do BigQuery ou via:
--   bq query --use_legacy_sql=false < deploy/migrate_buscas_v2.sql
--
-- BigQuery aceita ALTER TABLE ADD COLUMN em tabelas existentes (colunas ficam NULL
-- para linhas antigas). Isso é seguro e não requer recriar a tabela.

-- Substitua SEU_PROJECT pelo ID do projeto.

ALTER TABLE `SEU_PROJECT.garimpo.buscas`
ADD COLUMN IF NOT EXISTS shop_ids STRING;

ALTER TABLE `SEU_PROJECT.garimpo.buscas`
ADD COLUMN IF NOT EXISTS owner_uid STRING;

ALTER TABLE `SEU_PROJECT.garimpo.buscas`
ADD COLUMN IF NOT EXISTS rotation_cursor STRING;

ALTER TABLE `SEU_PROJECT.garimpo.buscas`
ADD COLUMN IF NOT EXISTS full_scan_at STRING;
