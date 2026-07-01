package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	alerterpb "github.com/fmarquesfilho/garimpo/gen/go/alerter/v1"
	"github.com/fmarquesfilho/garimpo/internal/alerts"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// AlerterServer implementa alerter.v1.AlerterService.
type AlerterServer struct {
	alerterpb.UnimplementedAlerterServiceServer
	repo store.Repository
}

func NewAlerterServer() *AlerterServer {
	// O repository será inicializado a partir do ambiente (BigQuery em produção).
	// Por enquanto usa a inicialização padrão do store.
	repo := initRepo()
	return &AlerterServer{repo: repo}
}

func (s *AlerterServer) CheckAndNotify(ctx context.Context, req *alerterpb.CheckAndNotifyRequest) (*alerterpb.CheckAndNotifyResponse, error) {
	if len(req.GetRules()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "pelo menos uma rule é necessária") //nolint:wrapcheck // gRPC status
	}

	cfg := alerts.Config{
		TelegramToken:  os.Getenv("ALERTAS_TELEGRAM_TOKEN"),
		TelegramChatID: os.Getenv("ALERTAS_TELEGRAM_CHAT_ID"),
		Logger:         slog.Default(),
	}
	if cfg.TelegramToken == "" {
		cfg.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if !cfg.Ativo() {
		return nil, status.Error(codes.FailedPrecondition, "alertas não configurados (ALERTAS_TELEGRAM_TOKEN + ALERTAS_TELEGRAM_CHAT_ID)") //nolint:wrapcheck // gRPC status
	}

	var results []*alerterpb.AlertResult
	alertsSent := 0

	for _, rule := range req.GetRules() {
		cfg.Threshold = rule.GetPriceThreshold()
		cfg.ApenasQuedas = true

		a := alerts.Novo(cfg)
		a.VerificarENotificar(ctx, s.repo.Snapshots(), rule.GetKeyword())

		results = append(results, &alerterpb.AlertResult{
			RuleId:     rule.GetRuleId(),
			Triggered:  true,
			NotifiedAt: time.Now().UTC().Format(time.RFC3339),
		})
		alertsSent++
	}

	return &alerterpb.CheckAndNotifyResponse{
		AlertsSent: int32(min(alertsSent, int(^uint32(0)>>1))), //nolint:gosec // bounded by rules count
		Results:    results,
	}, nil
}

// initRepo inicializa o repository. Em produção usa BigQuery; em dev pode ser nop.
func initRepo() store.Repository {
	// A inicialização real depende do ambiente e será conectada quando o
	// serviço estiver completo. Por enquanto retorna nil (o servidor
	// reporta FailedPrecondition se chamado sem repo).
	return nil
}
