package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/taskqueue"
)

// dispatchJob routes to the correct executor based on job params.
func (s *SchedulerServer) dispatchJob(job *registeredJob, params map[string]string) {
	jobType := params["type"]
	switch jobType {
	case "coupon_collection":
		s.executeCouponCollectionJob(job, params)
	case "scheduled_publish":
		s.executeScheduledPublish(job, params)
	default:
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	jobType := params["type"]
	var totalFound int32
	var keyword string

	switch jobType {
	case "shop_collection":
		// Coleta por loja: usa FetchShop(shop_id) ou Fetch(keyword) se keywords presentes
		shopID := params["shop_id"]
		keywords := params["keywords"]
		keyword = shopID // usado para alertas

		if keywords != "" {
			// Coleta filtrada: busca cada keyword dentro da loja
			for _, kw := range strings.Split(keywords, ",") {
				kw = strings.TrimSpace(kw)
				if kw == "" {
					continue
				}
				s.logger.Info("fetching filtered", slog.String("job", job.name), slog.String("keyword", kw), slog.String("shop_id", shopID))
				resp, err := s.collector.Fetch(ctx, &collectorpb.FetchRequest{
					Keyword:     kw,
					Limit:       50,
					Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
					OwnerUid:    params["owner_uid"],
				})
				if err != nil {
					s.logger.Error("fetch filtered falhou", slog.String("job", job.name), slog.String("keyword", kw), slog.String("erro", err.Error()))
					continue
				}
				totalFound += resp.GetTotalFound()
			}
			keyword = keywords
		} else if shopID != "" {
			// Coleta completa da loja: usa FetchShop
			shopIDInt, err := strconv.ParseInt(shopID, 10, 64)
			if err != nil {
				s.logger.Error("shop_id inválido", slog.String("job", job.name), slog.String("shop_id", shopID))
				return
			}
			s.logger.Info("fetching shop", slog.String("job", job.name), slog.Int64("shop_id", shopIDInt))
			resp, err := s.collector.FetchShop(ctx, &collectorpb.FetchShopRequest{
				ShopId:      shopIDInt,
				Limit:       50,
				Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
				OwnerUid:    params["owner_uid"],
			})
			if err != nil {
				s.logger.Error("fetch shop falhou", slog.String("job", job.name), slog.String("erro", err.Error()))
				return
			}
			totalFound = resp.GetTotalFound()
		}

	default:
		// Coleta por keyword(s) — buscas por palavra-chave (sem loja).
		// Suporta params["keywords"] (plural, comma-separated) ou params["keyword"] (singular, legado).
		keywords := params["keywords"]
		if keywords != "" {
			// Múltiplas keywords: itera como shop_collection faz
			for _, kw := range strings.Split(keywords, ",") {
				kw = strings.TrimSpace(kw)
				if kw == "" {
					continue
				}
				s.logger.Info("executing keyword job", slog.String("job", job.name), slog.String("keyword", kw))
				resp, err := s.collector.Fetch(ctx, &collectorpb.FetchRequest{
					Keyword:     kw,
					Limit:       50,
					Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
					OwnerUid:    params["owner_uid"],
				})
				if err != nil {
					s.logger.Error("keyword job falhou", slog.String("job", job.name), slog.String("keyword", kw), slog.String("erro", err.Error()))
					continue
				}
				totalFound += resp.GetTotalFound()
			}
			keyword = keywords
		} else {
			// Keyword única (legado)
			keyword = params["keyword"]
			if keyword == "" {
				keyword = job.name
			}
			s.logger.Info("executing keyword job", slog.String("job", job.name), slog.String("keyword", keyword))
			resp, err := s.collector.Fetch(ctx, &collectorpb.FetchRequest{
				Keyword:     keyword,
				Limit:       50,
				Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
				OwnerUid:    params["owner_uid"],
			})
			if err != nil {
				s.logger.Error("job falhou", slog.String("job", job.name), slog.String("erro", err.Error()))
				return
			}
			totalFound = resp.GetTotalFound()
		}
	}

	s.logger.Info("job concluído", slog.String("job", job.name), slog.Int("produtos", int(totalFound)))

	// Enqueue price alert via Cloud Tasks (rate-limited, durable, deduped).
	if s.alertQueue != nil {
		ownerUID := params["owner_uid"]
		chatID := params["chat_id"]
		if chatID == "" {
			chatID = s.alertChatID
		}
		if chatID != "" {
			if err := s.alertQueue.EnqueueAlert(ctx, taskqueue.AlertPayload{
				OwnerUID:  ownerUID,
				Keyword:   keyword,
				Threshold: 0.15,
				ChatID:    chatID,
			}, 0); err != nil {
				s.logger.Warn("failed to enqueue alert task", slog.String("keyword", keyword), slog.String("error", err.Error()))
			}
		}
	}
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
			s.logger.Warn("coupon collection skipped",
				slog.String("marketplace", mkt.String()),
				slog.String("error", err.Error()))
			continue
		}

		s.logger.Info("coupons collected",
			slog.String("marketplace", mkt.String()),
			slog.Int("count", int(resp.GetTotalFound())))

		s.triggerCouponDetection(ownerUID, mkt.String(), resp.GetFetchedAt())
	}

	s.logger.Info("coupon collection job complete", slog.String("job", job.name))
}

// triggerCouponDetection calls the Python analyzer to detect new/modified/expired coupons.
func (s *SchedulerServer) triggerCouponDetection(ownerUID, marketplace, fetchedAt string) {
	url := s.analyzerURL + "/detect-coupons"

	mktName := strings.ToLower(marketplace)
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

// executeScheduledPublish calls the C# API internal endpoint to send a scheduled publication.
// The Scheduler is responsible for WHEN, the C# API handles the actual sending.
func (s *SchedulerServer) executeScheduledPublish(job *registeredJob, params map[string]string) {
	s.mu.Lock()
	job.status = "running"
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		job.status = "active"
		job.lastRun = time.Now().UTC()
		s.mu.Unlock()
	}()

	publicacaoID := params["publicacao_id"]
	if publicacaoID == "" {
		s.logger.Error("scheduled_publish sem publicacao_id", slog.String("job", job.name))
		return
	}

	s.logger.Info("dispatching scheduled publish",
		slog.String("job", job.name),
		slog.String("publicacao_id", publicacaoID))

	// Call C# API internal endpoint to execute the publish
	apiURL := envOrDefault("API_URL", "http://localhost:8080")
	url := apiURL + "/internal/publish-scheduled"

	body := fmt.Sprintf(`{"publicacao_id":%q}`, publicacaoID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		s.logger.Error("falha ao criar request publish-scheduled", slog.String("error", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Error("falha ao chamar publish-scheduled", slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		s.logger.Error("publish-scheduled retornou erro",
			slog.Int("status", resp.StatusCode),
			slog.String("body", string(respBody)))
		return
	}

	s.logger.Info("scheduled publish dispatched successfully",
		slog.String("publicacao_id", publicacaoID))

	// Remove one-shot job after execution
	s.mu.Lock()
	delete(s.jobs, job.name)
	s.cron.Remove(job.id)
	s.mu.Unlock()
}
