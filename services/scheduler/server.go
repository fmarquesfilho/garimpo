package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	publisherpb "github.com/fmarquesfilho/garimpo/gen/go/publisher/v1"
	schedulerpb "github.com/fmarquesfilho/garimpo/gen/go/scheduler/v1"
	garimpotel "github.com/fmarquesfilho/garimpo/internal/otel"
	"github.com/fmarquesfilho/garimpo/internal/taskqueue"
)

// SchedulerServer implementa scheduler.v1.SchedulerService.
type SchedulerServer struct {
	schedulerpb.UnimplementedSchedulerServiceServer

	cron            *cron.Cron
	logger          *slog.Logger
	collector       collectorpb.CollectorServiceClient
	couponCollector couponpb.CouponCollectorServiceClient
	publisher       publisherpb.PublisherServiceClient
	analyzerURL     string
	alertChatID     string // fallback chat for alerts
	alertQueue      *taskqueue.Client

	mu   sync.RWMutex
	jobs map[string]*registeredJob
}

type registeredJob struct {
	id       cron.EntryID
	name     string
	cronExpr string
	status   string // active, paused, running
	lastRun  time.Time
	params   map[string]string
}

func NewSchedulerServer(collectorAddr, publisherAddr string, logger *slog.Logger) (*SchedulerServer, error) {
	// Connect to unified collector with OTel client interceptors
	collConn, err := grpc.NewClient(collectorAddr,
		append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			garimpotel.GRPCDialOptions()...)...)
	if err != nil {
		return nil, fmt.Errorf("conectar ao collector %s: %w", collectorAddr, err)
	}

	// Connect to publisher with OTel client interceptors
	pubConn, err := grpc.NewClient(publisherAddr,
		append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			garimpotel.GRPCDialOptions()...)...)
	if err != nil {
		return nil, fmt.Errorf("conectar ao publisher %s: %w", publisherAddr, err)
	}

	analyzerURL := envOrDefault("ANALYZER_URL", "http://localhost:8060")
	alertChatID := envOrDefault("ALERT_CHAT_ID", "")

	// Cloud Tasks client for durable alert dispatch
	var alertQ *taskqueue.Client
	projectID := envOrDefault("GCP_PROJECT_ID", "garimpo-500114")
	queueLocation := envOrDefault("CLOUD_TASKS_LOCATION", "southamerica-east1")
	alertQueueID := envOrDefault("ALERT_QUEUE_ID", "price-alerts")
	alertTargetURL := envOrDefault("ALERT_TARGET_URL", "")
	alertSA := envOrDefault("ALERT_SA_EMAIL", "")

	if alertTargetURL != "" {
		alertQ, err = taskqueue.New(context.Background(), taskqueue.Config{
			ProjectID:           projectID,
			Location:            queueLocation,
			QueueID:             alertQueueID,
			TargetURL:           alertTargetURL,
			ServiceAccountEmail: alertSA,
			Logger:              logger,
		})
		if err != nil {
			logger.Warn("cloud tasks client failed, alerts disabled", slog.String("error", err.Error()))
		} else {
			logger.Info("cloud tasks alert queue connected",
				slog.String("queue", alertQueueID),
				slog.String("target", alertTargetURL))
		}
	} else {
		logger.Info("ALERT_TARGET_URL not set, Cloud Tasks alerts disabled")
	}

	return &SchedulerServer{
		cron:            cron.New(cron.WithLocation(time.FixedZone("BRT", -3*60*60))),
		logger:          logger,
		collector:       collectorpb.NewCollectorServiceClient(collConn),
		couponCollector: couponpb.NewCouponCollectorServiceClient(collConn),
		publisher:       publisherpb.NewPublisherServiceClient(pubConn),
		analyzerURL:     analyzerURL,
		alertChatID:     alertChatID,
		alertQueue:      alertQ,
		jobs:            make(map[string]*registeredJob),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if v := lookupEnv(key); v != "" {
		return v
	}
	return fallback
}

// lookupEnv is a thin wrapper for testing.
var lookupEnv = os.Getenv

func (s *SchedulerServer) Start() {
	s.cron.Start()
	s.logger.Info("cron scheduler started")
}

func (s *SchedulerServer) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("cron scheduler stopped")
}

func (s *SchedulerServer) TriggerJob(ctx context.Context, req *schedulerpb.TriggerJobRequest) (*schedulerpb.TriggerJobResponse, error) {
	s.mu.RLock()
	job, exists := s.jobs[req.GetJobId()]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "job %q não encontrado", req.GetJobId())
	}

	// Execute job inline
	go s.dispatchJob(job, req.GetParams())

	return &schedulerpb.TriggerJobResponse{
		Accepted:    true,
		ExecutionId: fmt.Sprintf("%s-%d", req.GetJobId(), time.Now().UnixMilli()),
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *SchedulerServer) ListJobs(ctx context.Context, req *schedulerpb.ListJobsRequest) (*schedulerpb.ListJobsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*schedulerpb.Job
	for id, j := range s.jobs {
		if req.GetStatusFilter() != "" && req.GetStatusFilter() != "all" && j.status != req.GetStatusFilter() {
			continue
		}

		entry := s.cron.Entry(j.id)
		var nextRun string
		if !entry.Next.IsZero() {
			nextRun = entry.Next.Format(time.RFC3339)
		}
		var lastRun string
		if !j.lastRun.IsZero() {
			lastRun = j.lastRun.Format(time.RFC3339)
		}

		result = append(result, &schedulerpb.Job{
			Id:             id,
			Name:           j.name,
			CronExpression: j.cronExpr,
			Status:         j.status,
			LastRunAt:      lastRun,
			NextRunAt:      nextRun,
		})
	}

	return &schedulerpb.ListJobsResponse{Jobs: result}, nil
}

func (s *SchedulerServer) SetSchedule(ctx context.Context, req *schedulerpb.SetScheduleRequest) (*schedulerpb.SetScheduleResponse, error) {
	if req.GetJobId() == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id é obrigatório") //nolint:wrapcheck // gRPC status
	}
	if req.GetCronExpression() == "" {
		return nil, status.Error(codes.InvalidArgument, "cron_expression é obrigatório") //nolint:wrapcheck // gRPC status
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job if exists
	if existing, ok := s.jobs[req.GetJobId()]; ok {
		s.cron.Remove(existing.id)
	}

	jobStatus := "active"
	if !req.GetEnabled() {
		jobStatus = "paused"
	}

	job := &registeredJob{
		name:     req.GetJobId(),
		cronExpr: req.GetCronExpression(),
		status:   jobStatus,
		params:   req.GetParams(),
	}

	if req.GetEnabled() {
		entryID, err := s.cron.AddFunc(req.GetCronExpression(), func() {
			s.dispatchJob(job, job.params)
		})
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "cron expression inválida: %v", err)
		}
		job.id = entryID
	}

	s.jobs[req.GetJobId()] = job

	entry := s.cron.Entry(job.id)
	var nextRun string
	if !entry.Next.IsZero() {
		nextRun = entry.Next.Format(time.RFC3339)
	}

	return &schedulerpb.SetScheduleResponse{
		Success: true,
		Job: &schedulerpb.Job{
			Id:             req.GetJobId(),
			Name:           req.GetJobId(),
			CronExpression: req.GetCronExpression(),
			Status:         jobStatus,
			NextRunAt:      nextRun,
		},
	}, nil
}
