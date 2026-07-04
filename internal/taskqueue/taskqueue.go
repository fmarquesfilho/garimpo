// Package taskqueue encapsula a criação de Cloud Tasks para alertas de preço.
//
// Design: cada alerta é uma task HTTP que chama POST /internal/alerts/check
// na API C# (ingress). O Cloud Tasks cuida de:
//   - Rate limiting (max_dispatches_per_second na queue)
//   - Retries com backoff exponencial
//   - Deduplicação por task name (keyword + dia)
//   - Durabilidade (sobrevive restarts e deploys)
package taskqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Config para o enqueue de alertas via Cloud Tasks.
type Config struct {
	// ProjectID do GCP (ex: "garimpo-500114")
	ProjectID string
	// Location da queue (ex: "southamerica-east1")
	Location string
	// QueueID (ex: "price-alerts")
	QueueID string
	// TargetURL é o endpoint HTTP que receberá a task
	// (ex: "https://garimpei-v2-xxxxx.run.app/internal/alerts/check")
	TargetURL string
	// ServiceAccountEmail para OIDC auth na task
	ServiceAccountEmail string
	// Logger
	Logger *slog.Logger
}

// AlertPayload é o corpo da task enviada ao endpoint de alertas.
type AlertPayload struct {
	OwnerUID string `json:"owner_uid"`
	Keyword  string `json:"keyword"`
	// Threshold mínimo de variação para disparar (0.15 = 15%)
	Threshold float64 `json:"threshold"`
	// ChatID do Telegram para enviar o alerta (resolvido pelo scheduler antes de enfileirar)
	ChatID string `json:"chat_id"`
}

// Client encapsula o Cloud Tasks client.
type Client struct {
	cfg    Config
	client *cloudtasks.Client
}

// New cria um Client de Cloud Tasks. Fecha com Close().
func New(ctx context.Context, cfg Config) (*Client, error) {
	c, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("taskqueue: criar client: %w", err)
	}
	return &Client{cfg: cfg, client: c}, nil
}

// Close libera recursos do client.
func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("taskqueue: close: %w", err)
	}
	return nil
}

// EnqueueAlert cria uma task para verificar alertas de preço.
// O delay escala com o índice para evitar flood (1s entre tasks).
func (c *Client) EnqueueAlert(ctx context.Context, payload AlertPayload, delaySeconds int) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("taskqueue: marshal payload: %w", err)
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		c.cfg.ProjectID, c.cfg.Location, c.cfg.QueueID)

	// Task name para deduplicação: keyword + dia (evita enviar o mesmo alerta 2x no dia)
	today := time.Now().Format("2006-01-02")
	taskName := fmt.Sprintf("%s/tasks/alert-%s-%s", queuePath, sanitize(payload.Keyword), today)

	scheduleTime := time.Now().Add(time.Duration(delaySeconds) * time.Second)

	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			Name: taskName,
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        c.cfg.TargetURL,
					Body:       body,
					Headers:    map[string]string{"Content-Type": "application/json"},
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: c.cfg.ServiceAccountEmail,
						},
					},
				},
			},
			ScheduleTime: timestamppb.New(scheduleTime),
		},
	}

	_, err = c.client.CreateTask(ctx, req)
	if err != nil {
		// ALREADY_EXISTS = task já criada hoje (deduplicação OK)
		if isAlreadyExists(err) {
			c.cfg.Logger.Debug("alert task already exists (dedup)",
				slog.String("keyword", payload.Keyword))
			return nil
		}
		return fmt.Errorf("taskqueue: criar task: %w", err)
	}

	c.cfg.Logger.Info("alert task enqueued",
		slog.String("keyword", payload.Keyword),
		slog.Int("delay_s", delaySeconds),
		slog.String("task", taskName))

	return nil
}

// sanitize limpa o keyword para uso como task name (alfanumérico + hifens).
func sanitize(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		} else {
			result = append(result, '-')
		}
	}
	return string(result)
}

func isAlreadyExists(err error) bool {
	return err != nil && (fmt.Sprintf("%v", err) == "rpc error: code = AlreadyExists" ||
		contains(err.Error(), "ALREADY_EXISTS") ||
		contains(err.Error(), "AlreadyExists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
