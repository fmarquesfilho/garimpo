package source

import (
	"os"
	"path/filepath"
	"testing"
)

func escrever(t *testing.T, conteudo string) string {
	t.Helper()
	caminho := filepath.Join(t.TempDir(), "candidatos.csv")
	if err := os.WriteFile(caminho, []byte(conteudo), 0o600); err != nil {
		t.Fatal(err)
	}
	return caminho
}

func TestCSVValido(t *testing.T) {
	caminho := escrever(t, `id,name,category,price,commission,sales_30d,rating
P01,Sérum,cosméticos,89.90,0.12,320,4.8
P02,Perfume,perfumaria,159.90,0.10,210,4.7
`)
	produtos, err := NewCSVSource(caminho).Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(produtos) != 2 {
		t.Fatalf("esperava 2 produtos, veio %d", len(produtos))
	}
	p := produtos[0]
	if p.ID != "P01" || p.Name != "Sérum" || p.Category != "cosméticos" {
		t.Errorf("campos texto errados: %+v", p)
	}
	if p.Price != 89.90 || p.Commission != 0.12 || p.Sales30d != 320 || p.Rating != 4.8 {
		t.Errorf("campos numéricos errados: %+v", p)
	}
}

func TestCSVComLinkOpcional(t *testing.T) {
	caminho := escrever(t, `id,name,category,price,commission,sales_30d,rating,link
P01,Sérum,cosméticos,89.90,0.12,320,4.8,https://s.shopee/abc
`)
	produtos, err := NewCSVSource(caminho).Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if produtos[0].Link != "https://s.shopee/abc" {
		t.Errorf("link não lido: %q", produtos[0].Link)
	}
}

func TestCSVErros(t *testing.T) {
	casos := map[string]string{
		"poucas_colunas": "id,name,category,price,commission,sales_30d,rating\nP01,Sérum,cosméticos,89.90\n",
		"preco_invalido": "id,name,category,price,commission,sales_30d,rating\nP01,Sérum,cosméticos,abc,0.12,320,4.8\n",
		"so_cabecalho":   "id,name,category,price,commission,sales_30d,rating\n",
	}
	for nome, conteudo := range casos {
		t.Run(nome, func(t *testing.T) {
			if _, err := NewCSVSource(escrever(t, conteudo)).Fetch(); err == nil {
				t.Error("esperava erro, veio nil")
			}
		})
	}
}

func TestCSVArquivoInexistente(t *testing.T) {
	if _, err := NewCSVSource("/nao/existe.csv").Fetch(); err == nil {
		t.Error("arquivo inexistente deveria falhar")
	}
}
