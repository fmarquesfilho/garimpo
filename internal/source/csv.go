package source

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// CSVSource é o ADAPTADOR que funciona HOJE: lê um CSV exportado manualmente
// da aba de afiliados da Shopee (ou montado à mão). É a fonte do Incremento 0,
// que destrava o teste de viabilidade sem depender da API.
//
// Cabeçalho esperado:
//
//	id,name,category,price,commission,sales_30d,rating
type CSVSource struct {
	Path string
}

func NewCSVSource(path string) *CSVSource { return &CSVSource{Path: path} }

func (c *CSVSource) Name() string { return "csv:" + c.Path }

func (c *CSVSource) Fetch() ([]domain.Product, error) {
	f, err := os.Open(c.Path)
	if err != nil {
		return nil, fmt.Errorf("abrindo csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("lendo csv: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("csv sem linhas de dados: %w", apperr.ErrCSV)
	}

	produtos := make([]domain.Product, 0, len(rows)-1)
	for i, row := range rows[1:] { // pula o cabeçalho
		linha := i + 2
		if len(row) < 7 {
			return nil, fmt.Errorf("linha %d: esperava 7 colunas, veio %d: %w", linha, len(row), apperr.ErrCSV)
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("linha %d: preço inválido %q: %w", linha, row[3], err)
		}
		comm, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return nil, fmt.Errorf("linha %d: comissão inválida %q: %w", linha, row[4], err)
		}
		sales, err := strconv.Atoi(strings.TrimSpace(row[5]))
		if err != nil {
			return nil, fmt.Errorf("linha %d: vendas inválidas %q: %w", linha, row[5], err)
		}
		rating, err := strconv.ParseFloat(strings.TrimSpace(row[6]), 64)
		if err != nil {
			return nil, fmt.Errorf("linha %d: rating inválido %q: %w", linha, row[6], err)
		}

		link := ""
		if len(row) >= 8 {
			link = strings.TrimSpace(row[7])
		}

		produtos = append(produtos, domain.Product{
			ID:         strings.TrimSpace(row[0]),
			Name:       strings.TrimSpace(row[1]),
			Category:   strings.TrimSpace(row[2]),
			Price:      price,
			Commission: comm,
			Sales30d:   sales,
			Rating:     rating,
			Link:       link,
		})
	}
	return produtos, nil
}
