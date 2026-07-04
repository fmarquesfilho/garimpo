package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	schedulerpb "github.com/fmarquesfilho/garimpo/gen/go/scheduler/v1"
)

func newTestSchedulerServer(t *testing.T) *SchedulerServer {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	// Use a fake collector address — we won't actually call it in these tests.
	srv, err := NewSchedulerServer("localhost:0", "localhost:0", logger)
	if err != nil {
		t.Fatalf("failed to create scheduler server: %v", err)
	}
	return srv
}

func TestSetSchedule_EmptyJobId_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestSchedulerServer(t)

	_, err := srv.SetSchedule(context.Background(), &schedulerpb.SetScheduleRequest{
		JobId:          "",
		CronExpression: "*/5 * * * *",
		Enabled:        true,
	})

	if err == nil {
		t.Fatal("expected error for empty job_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSetSchedule_EmptyCron_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestSchedulerServer(t)

	_, err := srv.SetSchedule(context.Background(), &schedulerpb.SetScheduleRequest{
		JobId:          "job-1",
		CronExpression: "",
		Enabled:        true,
	})

	if err == nil {
		t.Fatal("expected error for empty cron_expression, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSetSchedule_ValidCron_CreatesJob(t *testing.T) {
	srv := newTestSchedulerServer(t)

	resp, err := srv.SetSchedule(context.Background(), &schedulerpb.SetScheduleRequest{
		JobId:          "coleta-perfume",
		CronExpression: "*/10 * * * *",
		Enabled:        true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Error("expected success=true")
	}
	if resp.GetJob().GetId() != "coleta-perfume" {
		t.Errorf("expected job id 'coleta-perfume', got %q", resp.GetJob().GetId())
	}
	if resp.GetJob().GetStatus() != "active" {
		t.Errorf("expected status 'active', got %q", resp.GetJob().GetStatus())
	}
}

func TestListJobs_Empty(t *testing.T) {
	srv := newTestSchedulerServer(t)

	resp, err := srv.ListJobs(context.Background(), &schedulerpb.ListJobsRequest{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetJobs()) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(resp.GetJobs()))
	}
}

func TestTriggerJob_NotFound(t *testing.T) {
	srv := newTestSchedulerServer(t)

	_, err := srv.TriggerJob(context.Background(), &schedulerpb.TriggerJobRequest{
		JobId: "nonexistent-job",
	})

	if err == nil {
		t.Fatal("expected error for nonexistent job, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}
