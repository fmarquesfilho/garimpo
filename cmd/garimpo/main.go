// Comando garimpo: gera a lista priorizada de candidatos do dia.
//
// Fonte CSV (padrão):
//
//	go run ./cmd/garimpo
//	go run ./cmd/garimpo -estrategia diversificada -top 8
//
// Fonte Shopee (API de afiliados): exige credenciais no ambiente.
//
//	export SHOPEE_APP_ID=...
//	export SHOPEE_SECRET=...
//	go run ./cmd/garimpo -fonte shopee -cat 100017 -categoria "cosméticos"
package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fmarquesfilho/garimpo/internal/engine"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

func main() {
	fonte := flag.String("fonte", "csv", "fonte de candidatos: csv | shopee")
	csvPath := flag.String("csv", "data/candidatos_exemplo.csv", "caminho do CSV (fonte csv)")
	catID := flag.Int("cat", 0, "productCatId da Shopee (fonte shopee)")
	categoria := flag.String("categoria", "", "rótulo de categoria carimbado nos produtos (fonte shopee)")
	keyword := flag.String("keyword", "", "busca por palavra-chave (fonte shopee)")

	estrategia := flag.String("estrategia", "nicho", "estratégia: nicho | diversificada")
	topN := flag.Int("top", 5, "quantos produtos retornar")
	minComm := flag.Float64("comissao-min", strategy.MinCommission, "piso de comissão (0..1)")
	minVendas := flag.Int("vendas-min", 0, "piso de vendas (0 = sem filtro)")
	minNota := flag.Float64("nota-min", 0, "piso de avaliação 0..5 (0 = sem filtro)")
	flag.Parse()

	var st strategy.Strategy
	switch *estrategia {
	case "nicho":
		st = strategy.NewNiche()
	case "diversificada":
		st = strategy.Diversified{}
	default:
		fmt.Fprintf(os.Stderr, "estratégia desconhecida: %q (use 'nicho' ou 'diversificada')\n", *estrategia)
		os.Exit(1)
	}

	var src source.ProductSource
	switch *fonte {
	case "csv":
		src = source.NewCSVSource(*csvPath)
	case "shopee":
		appID := os.Getenv("SHOPEE_APP_ID")
		secret := os.Getenv("SHOPEE_SECRET")
		if appID == "" || secret == "" {
			fmt.Fprintln(os.Stderr, "defina SHOPEE_APP_ID e SHOPEE_SECRET no ambiente")
			os.Exit(1)
		}
		sh := source.NewShopeeAPISource(appID, secret)
		sh.ProductCatID = *catID
		sh.CategoryLabel = *categoria
		sh.Keyword = *keyword
		src = sh
	default:
		fmt.Fprintf(os.Stderr, "fonte desconhecida: %q (use 'csv' ou 'shopee')\n", *fonte)
		os.Exit(1)
	}

	elig := strategy.Elegibilidade{ComissaoMin: *minComm, VendasMin: *minVendas, NotaMin: *minNota}
	eng := engine.New(src, st, elig)

	top, err := eng.Top(*topN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fonte: %s | Estratégia: %s | Piso de comissão: %.0f%%\n\n",
		src.Name(), st.Name(), *minComm*100)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "#\tPRODUTO\tCATEGORIA\tPREÇO\tCOMISSÃO\tVENDAS\tSCORE")
	for i, s := range top {
		p := s.Product
		fmt.Fprintf(w, "%d\t%s\t%s\tR$ %.2f\t%.0f%%\t%d\t%.3f\n",
			i+1, p.Name, p.Category, p.Price, p.Commission*100, p.Sales30d, s.Score)
	}
	w.Flush()
}
