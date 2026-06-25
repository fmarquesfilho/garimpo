//go:build gcp

package scheduler

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	scheduler "cloud.google.com/go/scheduler/apiv1"
	schedulerpb "cloud.google.com/go/scheduler/apiv1/schedulerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GCPScheduler cria/atualiza/deleta jobs no Cloud Scheduler.
type GCPScheduler struct {
	client   *scheduler.CloudSchedulerClient
	project  string
	location string
	timezone string
	apiURL   string // URL base da API (ex.: https://garimpo-api-xxx.run.app)
	token    string // COLETA_TOKEN para autenticar as chamadas
}

// NovoGCPScheduler cria um scheduler a partir dos parâmetros de ambiente.
func NovoGCPScheduler(ctx context.Context, project, location, apiURL, token string) (*GCPScheduler, error) {
	c, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduler client: %w", err)
	}
	return &GCPScheduler{
		client:   c,
		project:  project,
		location: location,
		timezone: "America/Sao_Paulo",
		apiURL:   strings.TrimRight(apiURL, "/"),
		token:    token,
	}, nil
}

func (s *GCPScheduler) Nome() string { return "gcp-scheduler" }

func (s *GCPScheduler) SyncBusca(ctx context.Context, buscaID string, keywords []string, cron string, params ColetaParams) error {
	if cron == "" {
		return s.DeletarBusca(ctx, buscaID, keywords)
	}

	// Se tem shop_ids, cria um único job para a busca inteira (não por keyword)
	if len(params.ShopIDs) > 0 && len(keywords) == 0 {
		keywords = []string{"all"} // placeholder — buildURI ignora keyword quando tem shop_ids
	}

	for _, kw := range keywords {
		jobID := jobName(buscaID, kw)
		uri := s.buildURI(kw, params)
		if err := s.criarOuAtualizar(ctx, jobID, cron, uri); err != nil {
			return fmt.Errorf("job %s: %w", jobID, err)
		}
	}
	return nil
}

func (s *GCPScheduler) DeletarBusca(ctx context.Context, buscaID string, keywords []string) error {
	for _, kw := range keywords {
		jobID := jobName(buscaID, kw)
		name := s.jobPath(jobID)
		err := s.client.DeleteJob(ctx, &schedulerpb.DeleteJobRequest{Name: name})
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("deletar job %s: %w", jobID, err)
		}
	}
	return nil
}

func (s *GCPScheduler) criarOuAtualizar(ctx context.Context, jobID, cron, uri string) error {
	name := s.jobPath(jobID)
	parent := fmt.Sprintf("projects/%s/locations/%s", s.project, s.location)

	job := &schedulerpb.Job{
		Name:     name,
		Schedule: cron,
		TimeZone: s.timezone,
		Target: &schedulerpb.Job_HttpTarget{
			HttpTarget: &schedulerpb.HttpTarget{
				Uri:        uri,
				HttpMethod: schedulerpb.HttpMethod_POST,
				Headers:    map[string]string{"X-Garimpo-Token": s.token},
			},
		},
	}

	// Tenta criar; se já existe, atualiza.
	_, err := s.client.CreateJob(ctx, &schedulerpb.CreateJobRequest{
		Parent: parent,
		Job:    job,
	})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.AlreadyExists {
		return err
	}

	// Atualiza o job existente
	_, err = s.client.UpdateJob(ctx, &schedulerpb.UpdateJobRequest{Job: job})
	return err
}

func (s *GCPScheduler) buildURI(keyword string, p ColetaParams) string {
	q := url.Values{}
	if p.BuscaID != "" {
		q.Set("busca_id", p.BuscaID)
	}
	if len(p.ShopIDs) > 0 {
		q.Set("fonte", "shopee-shop")
		ids := make([]string, 0, len(p.ShopIDs))
		for _, id := range p.ShopIDs {
			ids = append(ids, fmt.Sprintf("%d", id))
		}
		q.Set("shop_ids", strings.Join(ids, ","))
		if keyword != "" {
			q.Set("keyword", keyword) // filtra dentro das lojas
		}
	} else {
		q.Set("fonte", "shopee")
		q.Set("keyword", keyword)
	}
	if p.Categoria != "" {
		q.Set("categoria", p.Categoria)
	}
	if p.Estrategia != "" {
		q.Set("estrategia", p.Estrategia)
	}
	if p.Top > 0 {
		q.Set("top", fmt.Sprintf("%d", p.Top))
	}
	if p.VendasMin > 0 {
		q.Set("vendas_min", fmt.Sprintf("%d", p.VendasMin))
	}
	if p.NotaMin > 0 {
		q.Set("nota_min", fmt.Sprintf("%.1f", p.NotaMin))
	}
	return s.apiURL + "/api/coletar?" + q.Encode()
}

func (s *GCPScheduler) jobPath(jobID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.project, s.location, jobID)
}

// jobName gera um nome de job sanitizado (Cloud Scheduler aceita [a-z0-9-], máx 63).
func jobName(buscaID, keyword string) string {
	suffix := sanitize(keyword)
	name := "coleta-" + sanitize(buscaID) + "-" + suffix
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.TrimRight(name, "-")
}

func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var out []byte
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			out = append(out, c)
		case c == ' ' || c == '_':
			out = append(out, '-')
		}
	}
	return strings.Trim(string(out), "-")
}
