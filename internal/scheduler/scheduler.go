// Package scheduler gerencia os jobs do Cloud Scheduler de forma programática.
// Cada busca salva com cron gera N jobs (um por keyword). Jobs são criados ou
// atualizados quando a busca é salva, e deletados quando a busca é removida.
//
// A interface permite usar um NopScheduler em dev (sem GCP).
package scheduler

import (
	"context"
)

// Job representa um agendamento de coleta para uma keyword.
type Job struct {
	ID       string // ex.: "coleta-shiseido-shiseido"
	Cron     string // ex.: "0 8 * * *"
	Keyword  string
	Params   ColetaParams
}

// ColetaParams são os parâmetros de uma coleta agendada.
type ColetaParams struct {
	Categoria   string
	Estrategia  string
	Top         int
	VendasMin   int
	NotaMin     float64
}

// Scheduler cria/atualiza/deleta jobs de coleta periódica.
type Scheduler interface {
	// SyncBusca cria ou atualiza os jobs para uma busca (um por keyword).
	// Se cron estiver vazio, deleta os jobs existentes dessa busca.
	SyncBusca(ctx context.Context, buscaID string, keywords []string, cron string, params ColetaParams) error
	// DeletarBusca remove todos os jobs associados a uma busca.
	DeletarBusca(ctx context.Context, buscaID string, keywords []string) error
	Nome() string
}

// NopScheduler não faz nada — usado em dev/local.
type NopScheduler struct{}

func (NopScheduler) SyncBusca(context.Context, string, []string, string, ColetaParams) error {
	return nil
}
func (NopScheduler) DeletarBusca(context.Context, string, []string) error { return nil }
func (NopScheduler) Nome() string                                         { return "nop" }
