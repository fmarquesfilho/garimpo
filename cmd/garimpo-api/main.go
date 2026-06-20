// Comando garimpo-api: servidor HTTP que serve a curadoria em JSON para o
// frontend SvelteKit consumir.
//
//	go run ./cmd/garimpo-api                       # CSV de exemplo, porta 8080
//	go run ./cmd/garimpo-api -addr :9000
//
// Fonte ao vivo (front passa a receber dados reais sem nenhuma mudança):
//	export SHOPEE_APP_ID=... SHOPEE_SECRET=...
//	go run ./cmd/garimpo-api -fonte shopee -keyword "skincare" -categoria cosméticos
//
// Endpoints:
//	GET /api/health
//	GET /api/candidatos?estrategia=nicho|diversificada&top=10&comissao_min=0.07
//	GET /api/comparar?top=8
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/httpapi"
)

func main() {
	addr := flag.String("addr", ":8080", "endereço de escuta")
	csv := flag.String("csv", "data/candidatos_exemplo.csv", "CSV padrão (fonte csv)")
	fonte := flag.String("fonte", "csv", "fonte padrão: csv | shopee")
	cat := flag.Int("cat", 0, "productCatId da Shopee (fonte shopee)")
	categoria := flag.String("categoria", "", "rótulo de categoria carimbado (fonte shopee)")
	keyword := flag.String("keyword", "", "palavra-chave de busca (fonte shopee)")
	vendasMin := flag.Int("vendas-min", 0, "piso de vendas padrão (0 = sem filtro)")
	notaMin := flag.Float64("nota-min", 0, "piso de avaliação padrão 0..5 (0 = sem filtro)")
	cacheSeg := flag.Int("cache", 60, "TTL do cache de fetch, em segundos")
	flag.Parse()

	srv := &httpapi.Server{
		DefaultCSV: *csv,
		Fonte:      *fonte,
		CatID:      *cat,
		Categoria:  *categoria,
		Keyword:    *keyword,
		VendasMin:  *vendasMin,
		NotaMin:    *notaMin,
		CacheTTL:   time.Duration(*cacheSeg) * time.Second,
	}
	log.Printf("Garimpo API em %s | fonte=%s categoria=%q keyword=%q vendas-min=%d nota-min=%.1f cache=%ds",
		*addr, *fonte, *categoria, *keyword, *vendasMin, *notaMin, *cacheSeg)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
