-- Migração: Adicionar campos de origem do produto
-- Executar manualmente no BigQuery Console antes do deploy.
-- Ambos os ALTER TABLE são idempotentes (erro se já existir, mas não prejudica).

-- Campo "nome" na tabela de buscas (nome amigável da loja)
ALTER TABLE `garimpo.buscas` ADD COLUMN IF NOT EXISTS nome STRING;

-- Campo "origem_padrao" na tabela de buscas (origem padrão dos produtos da loja)
ALTER TABLE `garimpo.buscas` ADD COLUMN IF NOT EXISTS origem_padrao STRING;

-- Campo "origin" na tabela de snapshots (país de origem do produto no momento da coleta)
ALTER TABLE `garimpo.snapshots` ADD COLUMN IF NOT EXISTS origin STRING;
