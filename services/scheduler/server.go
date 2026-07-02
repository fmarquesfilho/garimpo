package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	schedulerpb "github.com/fmarquesfilho/garimpo/gen/go/scheduler/v1"
)

// SchedulerServer implementa scheduler.v1.SchedulerService.
type SchedulerServer struct {
	schedulerpb.UnimplementedSchedulerServiceServer

	cron            *cron.Cron
	logger          *slog.Logger
	collector       collectorpb.CollectorServiceClient
	couponCollector couponpb.CouponCollectorServiceClient
	analyzerURL     string

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

func NewSchedulerServer(collectorAddr, publisherAddr, alerterAddr string, logger *slog.Logger) (*SchedulerServer, error) {
	// Connect to product collector
	collConn, err := grpc.NewClient(collectorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("conectar ao collector %s: %w", collectorAddr, err)
	}

	// Connect to coupon collector (optional — uses env vars)
	var couponClient couponpb.CouponCollectorServiceClient
	couponAddr := envOrDefault("COUPON_COLLECTOR_SHOPEE_ADDR", "localhost:50061")
	couponConn, err := grpc.NewClient(couponAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		couponClient = couponpb.NewCouponCollectorServiceClient(couponConn)
	}

	analyzerURL := envOrDefault("ANALYZER_URL", "http://localhost:8060")

	// Publisher and alerter connections will be used when those features are wired
	_, _ = publisherAddr, alerterAddr

	return &SchedulerServer{
		cron:            cron.New(cron.WithLocation(time.FixedZone("BRT", -3*60*60))),
		logger:          logger,
		collector:       collectorpb.NewCollectorServiceClient(collConn),
		couponCollector: couponClient,
		analyzerURL:     analyzerURL,
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

// dispatchJob routes to the correct executor based on job params.
func (s *SchedulerServer) dispatchJob(job *registeredJob, params map[string]string) {
	jobType := params["type"]
	if jobType == "coupon_collection" {
		s.executeCouponCollectionJob(job, params)
	} else {
		s.executeJob(job, params)
	}
}

func (s *SchedulerServer) executeJob(job *registeredJob, params map[string]string) {
	s.mu.Lock()
	job.status = "running"
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		job.status = "active"
		job.lastRun = time.Now().UTC()
		s.mu.Unlock()
	}()

	keyword := params["keyword"]
	if keyword == "" {
		keyword = job.name
	}

	s.logger.Info("executing job", slog.String("job", job.name), slog.String("keyword", keyword))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resp, err := s.collector.Fetch(ctx, &collectorpb.FetchRequest{
		Keyword:     keyword,
		Limit:       50,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
	})
	if err != nil {
		s.logger.Error("job falhou", slog.String("job", job.name), slog.String("erro", err.Error()))
		return
	}

	s.logger.Info("job concluído", slog.String("job", job.name), slog.Int("produtos", int(resp.GetTotalFound())))
}

// executeCouponCollectionJob collects coupons from all configured marketplaces sequentially.
func (s *SchedulerServer) executeCouponCollectionJob(job *registeredJob, params map[string]string) {
	s.mu.Lock()
	job.status = "running"
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		job.status = "active"
		job.lastRun = time.Now().UTC()
		s.mu.Unlock()
	}()

	ownerUID := params["owner_uid"]
	if ownerUID == "" {
		s.logger.Error("coupon job sem owner_uid", slog.String("job", job.name))
		return
	}

	if s.couponCollector == nil {
		s.logger.Error("coupon collector não configurado")
		return
	}

	s.logger.Info("executing coupon collection job",
		slog.String("job", job.name), slog.String("owner_uid", ownerUID))

	// Sequential: Shopee → Amazon → Mercado Livre
	marketplaces := []collectorpb.Marketplace{
		collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		collectorpb.Marketplace_MARKETPLACE_AMAZON,
		collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE,
	}

	for _, mkt := range marketplaces {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		resp, err := s.couponCollector.FetchCoupons(ctx, &couponpb.FetchCouponsRequest{
			OwnerUid:    ownerUID,
			Marketplace: mkt,
			PageSize:    500,
		})
		cancel()

		if err != nil {
			// Skip this marketplace, continue with next (R4-AC3 graceful degradation)
			s.logger.Warn("coupon collection skipped",
				slog.String("marketplace", mkt.String()),
				slog.String("error", err.Error()))
			continue
		}

		s.logger.Info("coupons collected",
			slog.String("marketplace", mkt.String()),
			slog.Int("count", int(resp.GetTotalFound())))

		// Trigger detection in analyzer
		s.triggerCouponDetection(ownerUID, mkt.String(), resp.GetFetchedAt())
	}

	s.logger.Info("coupon collection job complete", slog.String("job", job.name))
}

// triggerCouponDetection calls the Python analyzer to detect new/modified/expired coupons.
func (s *SchedulerServer) triggerCouponDetection(ownerUID, marketplace, fetchedAt string) {
	url := s.analyzerURL + "/detect-coupons"

	// Map proto marketplace name to domain string
	mktName := strings.ToLower(marketplace)
	// MARKETPLACE_SHOPEE -> shopee, etc.
	mktName = strings.TrimPrefix(mktName, "marketplace_")

	body := fmt.Sprintf(`{"owner_uid":%q,"marketplace":%q,"snapshot_timestamp":%q}`,
		ownerUID, mktName, fetchedAt)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		s.logger.Error("falha ao criar request detect-coupons", slog.String("error", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Error("falha ao chamar detect-coupons", slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("detect-coupons retornou erro",
			slog.Int("status", resp.StatusCode),
			slog.String("marketplace", mktName))
	}
}
