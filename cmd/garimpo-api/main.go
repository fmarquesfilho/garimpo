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

	"github.com/fmarquesfilho/garimpo/internal/httpapi"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/store"
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

	// Store de eventos: NopStore por padrão; BigQueryStore com -tags gcp + env.
	eventos, err := store.Novo(context.Background())
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	// Publicador: Telegram se TELEGRAM_BOT_TOKEN/CHAT_ID estiverem no ambiente;
	// senão, Mock (não envia nada).
	pub := publish.Novo()

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
		Publicador: pub,
	}
	log.Printf("Garimpo API em %s | fonte=%s categoria=%q keyword=%q vendas-min=%d nota-min=%.1f cache=%ds store=%s publicador=%s",
		*addr, *fonte, *categoria, *keyword, *vendasMin, *notaMin, *cacheSeg, eventos.Nome(), pub.Nome())
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
