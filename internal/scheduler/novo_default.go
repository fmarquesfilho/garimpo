//go:build !gcp

package scheduler

import "context"

// Novo (padrão) não gerencia scheduler. Compile com -tags gcp para usar o Cloud Scheduler.
func Novo(ctx context.Context) (Scheduler, error) {
	_ = ctx
	return NopScheduler{}, nil
}
