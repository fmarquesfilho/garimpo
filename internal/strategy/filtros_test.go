package strategy

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

var produtosTeste = []domain.Product{
	{ID: "A", Commission: 0.15, Sales30d: 100, Rating: 4.5, Price: 50},
	{ID: "B", Commission: 0.05, Sales30d: 200, Rating: 4.8, Price: 30}, // comissão baixa
	{ID: "C", Commission: 0.10, Sales30d: 2, Rating: 4.0, Price: 100},  // poucas vendas
	{ID: "D", Commission: 0.12, Sales30d: 50, Rating: 3.0, Price: 200}, // nota baixa
	{ID: "E", Commission: 0.03, Sales30d: 0, Rating: 0, Price: 10},     // tudo baixo
}

func TestPipelineVazioPassaTudo(t *testing.T) {
	pl := PipelineMonitoramento()
	out := pl.Aplicar(produtosTeste)
	if len(out) != len(produtosTeste) {
		t.Errorf("pipeline vazio deveria passar tudo: esperava %d, veio %d", len(produtosTeste), len(out))
	}
}

func TestPipelineCuradoriaPadraoFiltraComissao(t *testing.T) {
	pl := PipelineCuradoria(Elegibilidade{ComissaoMin: 0.07})
	out := pl.Aplicar(produtosTeste)
	// B (5%) e E (3%) devem ser filtrados
	for _, p := range out {
		if p.Commission < 0.07 {
			t.Errorf("produto %s com comissão %.2f não deveria passar (piso 7%%)", p.ID, p.Commission)
		}
	}
	if len(out) != 3 { // A, C, D passam
		t.Errorf("esperava 3 produtos, veio %d", len(out))
	}
}

func TestPipelineComVendasMin(t *testing.T) {
	pl := PipelineCuradoria(Elegibilidade{ComissaoMin: 0.05, VendasMin: 10})
	out := pl.Aplicar(produtosTeste)
	// E (0 vendas) e C (2 vendas) devem ser filtrados por vendas
	for _, p := range out {
		if p.Sales30d < 10 {
			t.Errorf("produto %s com %d vendas não deveria passar (piso 10)", p.ID, p.Sales30d)
		}
	}
}

func TestPipelineComNotaMin(t *testing.T) {
	pl := PipelineCuradoria(Elegibilidade{ComissaoMin: 0, NotaMin: 4.0})
	out := pl.Aplicar(produtosTeste)
	// D (3.0) e E (0) devem ser filtrados
	for _, p := range out {
		if p.Rating < 4.0 {
			t.Errorf("produto %s com nota %.1f não deveria passar (piso 4.0)", p.ID, p.Rating)
		}
	}
}

func TestPipelineComFiltroPreco(t *testing.T) {
	pl := Pipeline{FiltroPrecoMax{Max: 80}}
	out := pl.Aplicar(produtosTeste)
	for _, p := range out {
		if p.Price > 80 {
			t.Errorf("produto %s com preço %.2f não deveria passar (teto 80)", p.ID, p.Price)
		}
	}
	// A (50), B (30), E (10) passam; C (100), D (200) filtrados
	if len(out) != 3 {
		t.Errorf("esperava 3, veio %d", len(out))
	}
}

func TestFiltrosCompostos(t *testing.T) {
	// Pipeline com múltiplos filtros: comissão + vendas + preço
	pl := Pipeline{
		FiltroComissao{Min: 0.05},
		FiltroVendas{Min: 10},
		FiltroPrecoMax{Max: 150},
	}
	out := pl.Aplicar(produtosTeste)
	// A: comissão ok, vendas ok, preço ok → passa
	// B: comissão baixa (5%=ok pois piso é 5%), vendas ok, preço ok → passa
	// C: comissão ok, vendas baixa → filtrado
	// D: comissão ok, vendas ok, preço 200 > 150 → filtrado
	// E: comissão baixa (3%) → filtrado
	if len(out) != 2 {
		ids := ""
		for _, p := range out {
			ids += p.ID + " "
		}
		t.Errorf("esperava 2 (A,B), veio %d: %s", len(out), ids)
	}
}

func TestPipelineCuradoriaCompatibilidadeComElegibilidade(t *testing.T) {
	// Verifica que PipelineCuradoria produz o mesmo resultado que Elegibilidade.Atende
	elig := Elegibilidade{ComissaoMin: 0.07, VendasMin: 5, NotaMin: 4.0}
	pl := PipelineCuradoria(elig)

	outPipeline := pl.Aplicar(produtosTeste)
	var outElig []domain.Product
	for _, p := range produtosTeste {
		if elig.Atende(p) {
			outElig = append(outElig, p)
		}
	}

	if len(outPipeline) != len(outElig) {
		t.Errorf("pipeline e elegibilidade devem produzir o mesmo resultado: pipeline=%d elig=%d",
			len(outPipeline), len(outElig))
	}
}
