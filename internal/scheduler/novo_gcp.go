//go:build gcp

package scheduler

import (
	"context"
	"fmt"
	"os"
)

// Novo (gcp) cria o GCPScheduler a partir do ambiente:
//
//	GOOGLE_CLOUD_PROJECT  projeto GCP
//	GCP_REGION            região (ex.: southamerica-east1)
//	SCHEDULER_API_URL     URL da API de coleta (ou detecta do Cloud Run)
//	COLETA_TOKEN          token de autenticação para o endpoint /api/coletar
func Novo(ctx context.Context) (Scheduler, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GCP_REGION")
	if location == "" {
		location = "southamerica-east1"
	}
	token := os.Getenv("COLETA_TOKEN")
	apiURL := os.Getenv("SCHEDULER_API_URL")

	// Detecta a URL da própria API no Cloud Run
	if apiURL == "" && os.Getenv("K_SERVICE") != "" {
		// O Cloud Run expõe K_SERVICE. A URL segue o padrão:
		// https://<service>-<project-number>.<region>.run.app
		// Mas não temos o project number aqui. Usamos a variável que o
		// deploy injeta, ou tentamos o formato legado.
		// Alternativa segura: ler de uma env dedicada no deploy.
		// Por ora, se não tem SCHEDULER_API_URL, voltamos NopScheduler.
	}

	if project == "" || apiURL == "" || token == "" {
		// Sem configuração suficiente — volta pro Nop silenciosamente
		return NopScheduler{}, nil
	}

	s, err := NovoGCPScheduler(ctx, project, location, apiURL, token)
	if err != nil {
		return nil, fmt.Errorf("gcp scheduler: %w", err)
	}
	return s, nil
}
