// Comando garimpo-api: servidor HTTP que serve a curadoria em JSON para o
// frontend SvelteKit consumir.
//
//	go run ./cmd/garimpo-api                       # CSV de exemplo, porta 8080
//	go run ./cmd/garimpo-api -addr :9000
//
// Fonte ao vivo (front passa a receber dados reais sem nenhuma mudança):
//
//	export SHOPEE_APP_ID=... SHOPEE_SECRET=...
//	go run ./cmd/garimpo-api -fonte shopee -keyword "skincare" -categoria cosméticos
//
// Endpoints:
//
//	GET /api/health
//	GET /api/candidatos?estrategia=nicho|diversificada&top=10&comissao_min=0.07
//	GET /api/comparar?top=8
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/auth"
	"github.com/fmarquesfilho/garimpo/internal/httpapi"
	"github.com/fmarquesfilho/garimpo/internal/logs"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/store"
	"github.com/fmarquesfilho/garimpo/internal/tenant"
)

func main() {
	// Cloud Run injeta PORT; honramos como padrão do -addr.
	addrPadrao := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addrPadrao = ":" + p
	}

	addr := flag.String("addr", addrPadrao, "endereço de escuta")
	csv := flag.String("csv", "data/candidatos_exemplo.csv", "CSV padrão (fonte csv)")
	fonte := flag.String("fonte", "csv", "fonte padrão: csv | shopee")
	cat := flag.Int("cat", 0, "productCatId da Shopee (fonte shopee)")
	categoria := flag.String("categoria", "", "rótulo de categoria carimbado (fonte shopee)")
	keyword := flag.String("keyword", "", "palavra-chave de busca (fonte shopee)")
	vendasMin := flag.Int("vendas-min", 0, "piso de vendas padrão (0 = sem filtro)")
	notaMin := flag.Float64("nota-min", 0, "piso de avaliação padrão 0..5 (0 = sem filtro)")
	exploracao := flag.Float64("exploracao", 0, "fração de vagas para exploração (hold-out), 0..0.9")
	cacheSeg := flag.Int("cache", 60, "TTL do cache de fetch, em segundos")
	flag.Parse()

	// Logging estruturado por criticidade (LOG_LEVEL / LOG_FORMAT no ambiente).
	logger := logs.Init()

	// Store de eventos: NopStore por padrão; BigQueryStore com -tags gcp + env.
	eventos, err := store.Novo(context.Background())
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	// Migration automática: garante que as tabelas existam no BigQuery. Idempotente.
	if err := eventos.EnsureSchema(context.Background()); err != nil {
		logger.Warn("EnsureSchema falhou (talvez tabelas já existam ou permissão insuficiente)",
			"erro", err)
	} else {
		logger.Info("schema do banco verificado", "store", eventos.Nome())
	}

	// Publicador: Dispatcher com TelegramSender se TELEGRAM_BOT_TOKEN/CHAT_ID
	// estiverem no ambiente (com suporte a múltiplos destinos); senão, Mock.
	destinos, templates := criarStoresAuxiliares(eventos)
	pub := publish.NovoComDestinos(destinos)

	// Scheduler: Cloud Scheduler com -tags gcp + env; NopScheduler caso contrário.
	sched, err := scheduler.Novo(context.Background())
	if err != nil {
		logger.Warn("scheduler não disponível", "erro", err)
		sched = scheduler.NopScheduler{}
	} else {
		logger.Info("scheduler configurado", "tipo", sched.Nome())
	}

	// Auth: Firebase Auth com -tags gcp + env; NopVerifier caso contrário.
	verifier, err := auth.Novo(context.Background())
	if err != nil {
		logger.Warn("auth não disponível", "erro", err)
		verifier = auth.NopVerifier{}
	}

	srv := &httpapi.Server{
		DefaultCSV: *csv,
		Fonte:      *fonte,
		CatID:      *cat,
		Categoria:  *categoria,
		Keyword:    *keyword,
		VendasMin:  *vendasMin,
		NotaMin:    *notaMin,
		Exploracao: *exploracao,
		CacheTTL:   time.Duration(*cacheSeg) * time.Second,
		Eventos:    eventos,
		Logger:     logger,
		Publicador: pub,
		Scheduler:  sched,
		Auth:       verifier,
		Destinos:   destinos,
		Templates:  templates,
		Tenants:    tenant.NewMemoryStore(),
		LogBuffer:  logs.NovoBuffer(500),
	}
	logger.Info("garimpo-api iniciando",
		"addr", *addr, "fonte", *fonte, "categoria", *categoria, "keyword", *keyword,
		"vendas_min", *vendasMin, "nota_min", *notaMin, "exploracao", *exploracao,
		"cache_s", *cacheSeg, "store", eventos.Nome(), "publicador", pub.Nome())

	httpSrv := &http.Server{
		Addr:         *addr,
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
