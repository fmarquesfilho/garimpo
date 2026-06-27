//go:build gcp

package store

import (
	"context"

	"cloud.google.com/go/bigquery"
)

// EnsureSchema cria as tabelas do dataset se ainda não existirem. Idempotente —
// chamado no startup toda vez, de modo que o banco evolui automaticamente ao
// deploy sem passo manual de "criar tabelas".
func (s *BigQueryStore) EnsureSchema(ctx context.Context) error {
	ds := s.client.Dataset(s.dataset)

	// --- tabela eventos ---
	eSchema := bigquery.Schema{
		{Name: "tipo", Type: bigquery.StringFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "canal", Type: bigquery.StringFieldType},
		{Name: "sub_id", Type: bigquery.StringFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "vendas", Type: bigquery.IntegerFieldType},
		{Name: "score", Type: bigquery.FloatFieldType},
		{Name: "em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, s.tabela, eSchema, "em"); err != nil {
		return err
	}

	// --- tabela snapshots ---
	sSchema := bigquery.Schema{
		{Name: "coletado_em", Type: bigquery.TimestampFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "keyword", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "posicao", Type: bigquery.IntegerFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "vendas", Type: bigquery.IntegerFieldType},
		{Name: "nota", Type: bigquery.FloatFieldType},
		{Name: "score", Type: bigquery.FloatFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, s.tabelaSnap, sSchema, "coletado_em"); err != nil {
		return err
	}

	// --- tabela buscas ---
	bSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "keywords", Type: bigquery.StringFieldType},
		{Name: "shop_ids", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "comissao_min", Type: bigquery.FloatFieldType},
		{Name: "vendas_min", Type: bigquery.IntegerFieldType},
		{Name: "nota_min", Type: bigquery.FloatFieldType},
		{Name: "top", Type: bigquery.IntegerFieldType},
		{Name: "cron", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "owner_uid", Type: bigquery.StringFieldType},
		{Name: "origem_padrao", Type: bigquery.StringFieldType},
		{Name: "rotation_cursor", Type: bigquery.StringFieldType},
		{Name: "full_scan_at", Type: bigquery.StringFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "buscas", bSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela destinos (append-only) ---
	dSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "tipo", Type: bigquery.StringFieldType},
		{Name: "config", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "destinos", dSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela templates (append-only) ---
	tSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "corpo", Type: bigquery.StringFieldType},
		{Name: "com_foto", Type: bigquery.BooleanFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "templates", tSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela publicacoes ---
	pSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "link", Type: bigquery.StringFieldType},
		{Name: "imagem", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "destino_id", Type: bigquery.StringFieldType},
		{Name: "template_id", Type: bigquery.StringFieldType},
		{Name: "agendada_em", Type: bigquery.StringFieldType},
		{Name: "status", Type: bigquery.StringFieldType},
		{Name: "detalhe", Type: bigquery.StringFieldType},
		{Name: "criada_em", Type: bigquery.TimestampFieldType},
		{Name: "enviada_em", Type: bigquery.StringFieldType},
		{Name: "owner_uid", Type: bigquery.StringFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "publicacoes", pSchema, "criada_em"); err != nil {
		return err
	}

	// --- tabela favoritos (append-only) ---
	fSchema := bigquery.Schema{
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "link", Type: bigquery.StringFieldType},
		{Name: "imagem", Type: bigquery.StringFieldType},
		{Name: "loja", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "origem", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "owner_uid", Type: bigquery.StringFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	return criarSeNaoExistir(ctx, ds, "favoritos", fSchema, "salvo_em")
}

// criarSeNaoExistir cria a tabela particionada por dia se ainda não existir.
// Se já existir, tenta evoluir o schema adicionando colunas novas (sem remover existentes).
func criarSeNaoExistir(ctx context.Context, ds *bigquery.Dataset, nome string, schema bigquery.Schema, campoPartition string) error {
	t := ds.Table(nome)
	md, err := t.Metadata(ctx)
	if err == nil {
		return evoluirSchema(ctx, t, md, schema)
	}
	meta := &bigquery.TableMetadata{
		Schema: schema,
		TimePartitioning: &bigquery.TimePartitioning{
			Type:  bigquery.DayPartitioningType,
			Field: campoPartition,
		},
	}
	return t.Create(ctx, meta)
}

// evoluirSchema adiciona colunas que existem no schema desejado mas não na tabela.
func evoluirSchema(ctx context.Context, t *bigquery.Table, md *bigquery.TableMetadata, desejado bigquery.Schema) error {
	existentes := make(map[string]bool)
	for _, f := range md.Schema {
		existentes[f.Name] = true
	}

	var novas bigquery.Schema
	for _, f := range desejado {
		if !existentes[f.Name] {
			novas = append(novas, f)
		}
	}

	if len(novas) == 0 {
		return nil
	}

	atualizado := append(md.Schema, novas...)
	_, err := t.Update(ctx, bigquery.TableMetadataToUpdate{Schema: atualizado}, md.ETag)
	return err
}
