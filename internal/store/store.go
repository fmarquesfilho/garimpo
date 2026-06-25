// Package store registra eventos (decisões de curadoria, e depois conversões)
// para análise posterior. A interface isola o destino: NopStore em dev/local,
// BigQueryStore em produção (atrás da build tag `gcp`, para não pesar o build
// padrão nem o CI com a dependência do BigQuery).
package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Evento é um fato registrado. As tags JSON casam com o objeto `candidato` que
// o front já manipula, então dá para postar o candidato direto + um `tipo`.
type Evento struct {
	Tipo       string    `json:"tipo"` // ex.: "selecao", "publicacao"
	ProdutoID  string    `json:"id"`
	Nome       string    `json:"nome"`
	Categoria  string    `json:"categoria"`
	Estrategia string    `json:"estrategia"`
	Canal      string    `json:"canal,omitempty"`  // preenchido em publicações
	SubID      string    `json:"sub_id,omitempty"` // atribuição (canal_estrategia_data)
	Comissao   float64   `json:"comissao"`
	Preco      float64   `json:"preco"`
	Vendas     int       `json:"vendas"`
	Score      float64   `json:"score"`
	Em         time.Time `json:"-"` // carimbado no servidor
}

// EstatCategoria resume os snapshots de mercado de uma categoria na janela.
// É a primeira camada descritiva do pipeline de análise (Fase 1 da estratégia):
// "como está cada categoria agora", sobre o dado que já se acumula no tempo.
type EstatCategoria struct {
	Categoria       string  `json:"categoria"`
	Amostras        int     `json:"amostras"`
	ComissaoMedia   float64 `json:"comissao_media"`
	ComissaoMediana float64 `json:"comissao_mediana"`
	PrecoMedio      float64 `json:"preco_medio"`
	VendasMedia     float64 `json:"vendas_media"`
	TeorMedio       float64 `json:"teor_medio"`
}

// Estatisticas é o resumo descritivo dos snapshots coletados numa janela.
type Estatisticas struct {
	Fonte         string           `json:"fonte"`          // "bigquery" | "nop"
	DiasJanela    int              `json:"dias_janela"`    // janela considerada
	TotalAmostras int              `json:"total_amostras"` // itens de snapshot
	PorCategoria  []EstatCategoria `json:"por_categoria"`  // agregado por categoria
	GeradoEm      time.Time        `json:"gerado_em"`
}

// Busca é um "perfil de coleta": um agrupamento de keywords com filtros comuns,
// reusável manualmente e candidato à coleta periódica. O campo `Keywords` é a
// lista de termos que serão buscados na Shopee (ex.: ["kenzo", "shiseido"]).
// Compatibilidade: se o cliente enviar apenas `keyword` (string), o servidor
// normaliza para `Keywords` com um único elemento.
type Busca struct {
	ID          string    `json:"id"`
	Keywords    []string  `json:"keywords"`             // um ou mais termos de busca
	ShopIDs     []int64   `json:"shop_ids,omitempty"`   // IDs de lojas a monitorar (shopOfferV2)
	Categoria   string    `json:"categoria"`
	Estrategia  string    `json:"estrategia"`           // "nicho" | "diversificada" | "ambas"
	ComissaoMin float64   `json:"comissao_min"`
	VendasMin   int       `json:"vendas_min"`
	NotaMin     float64   `json:"nota_min"`
	Top         int       `json:"top"`
	Cron        string    `json:"cron"`  // ex.: "0 8 * * *" (vazio = só manual)
	Ativo       bool      `json:"ativo"` // false = removida (tombstone)
	OwnerUID    string    `json:"owner_uid,omitempty"` // uid do Firebase Auth
	SalvoEm     time.Time `json:"salvo_em"`

	// Rotação de catálogo: próxima página a coletar por loja (JSON map shopID→page)
	RotationCursor map[int64]int `json:"rotation_cursor,omitempty"`
	// Timestamp da última varredura completa do catálogo (por loja)
	FullScanAt map[int64]string `json:"full_scan_at,omitempty"`

	// Legado
	KeywordLegado string `json:"keyword,omitempty"`
	NomeLegado    string `json:"nome,omitempty"`
}

// NormalizarBusca garante que a busca tenha um ID e que Keywords esteja
// preenchida. Converte campos legados (nome/keyword como string).
func NormalizarBusca(b Busca) Busca {
	if len(b.Keywords) == 0 && b.KeywordLegado != "" {
		b.Keywords = []string{b.KeywordLegado}
	}
	if b.ID == "" && b.NomeLegado != "" {
		b.ID = slugificar(b.NomeLegado)
	}
	if b.ID == "" && len(b.Keywords) > 0 {
		b.ID = slugificar(b.Keywords[0])
	}
	// Se só tem shop_ids sem keywords, gera ID a partir do primeiro shopId
	if b.ID == "" && len(b.ShopIDs) > 0 {
		b.ID = fmt.Sprintf("loja-%d", b.ShopIDs[0])
	}
	if b.Estrategia == "" {
		b.Estrategia = "nicho"
	}
	b.KeywordLegado = ""
	b.NomeLegado = ""
	return b
}

// slugificar transforma uma string em identificador sem espaços/acentos.
func slugificar(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var out []rune
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			out = append(out, r)
		case r == ' ' || r == '_':
			out = append(out, '-')
		default:
			// mapeia acentos comuns do português
			switch r {
			case 'á', 'à', 'ã', 'â', 'ä':
				out = append(out, 'a')
			case 'é', 'è', 'ê', 'ë':
				out = append(out, 'e')
			case 'í', 'ì', 'î', 'ï':
				out = append(out, 'i')
			case 'ó', 'ò', 'õ', 'ô', 'ö':
				out = append(out, 'o')
			case 'ú', 'ù', 'û', 'ü':
				out = append(out, 'u')
			case 'ç':
				out = append(out, 'c')
			case 'ñ':
				out = append(out, 'n')
			}
		}
	}
	// remove hífens duplicados/no início/fim
	result := strings.Trim(strings.ReplaceAll(string(out), "--", "-"), "-")
	if result == "" {
		return "busca"
	}
	return result
}

// EventoStore registra eventos. Implementações: NopStore e BigQueryStore.
type EventoStore interface {
	Registrar(ctx context.Context, e Evento) error
	RegistrarSnapshot(ctx context.Context, s Snapshot) error
	Estatisticas(ctx context.Context, dias int) (Estatisticas, error)
	SalvarBusca(ctx context.Context, b Busca) error
	ListarBuscas(ctx context.Context) ([]Busca, error)
	HistoricoColetas(ctx context.Context, dias int) ([]ColetaResumo, error)
	Conversoes(ctx context.Context, dias int) ([]ConversaoResumo, error)
	// Publicações agendadas
	SalvarPublicacao(ctx context.Context, p Publicacao) error
	ListarPublicacoes(ctx context.Context, status string) ([]Publicacao, error)
	AtualizarPublicacao(ctx context.Context, id, status, detalhe string) error
	// Novidades de lojas monitoradas (diff de snapshots)
	Novidades(ctx context.Context, buscaID string, dias int) (NovidadesLojas, error)
	// Evolução de preços das lojas monitoradas ao longo do tempo
	EvolucaoLojas(ctx context.Context, dias int) (EvolucaoLojasResult, error)
	EnsureSchema(ctx context.Context) error
	Nome() string
}

// NovidadesLojas é o resultado da comparação de snapshots para lojas monitoradas.
type NovidadesLojas struct {
	BuscaID       string           `json:"busca_id"`
	DiasJanela    int              `json:"dias_janela"`
	ProdutosNovos []ProdutoNovo    `json:"produtos_novos"`
	Variacoes     []VariacaoPreco  `json:"variacoes"`
	TotalAtual    int              `json:"total_atual"`
}

// ProdutoNovo é um produto que apareceu pela primeira vez no último snapshot.
type ProdutoNovo struct {
	ProdutoID string  `json:"produto_id"`
	Nome      string  `json:"nome"`
	Preco     float64 `json:"preco"`
	Comissao  float64 `json:"comissao"`
	Vendas    int     `json:"vendas"`
	Nota      float64 `json:"nota"`
	DetectadoEm string `json:"detectado_em"`
}

// VariacaoPreco sinaliza um produto cujo preço mudou entre snapshots.
type VariacaoPreco struct {
	ProdutoID    string  `json:"produto_id"`
	Nome         string  `json:"nome"`
	PrecoAnterior float64 `json:"preco_anterior"`
	PrecoAtual   float64 `json:"preco_atual"`
	Variacao     float64 `json:"variacao_pct"` // ex.: -0.20 = baixou 20%
	DetectadoEm  string  `json:"detectado_em"`
}

// EvolucaoLojasResult contém os dados de evolução de preço das lojas monitoradas.
type EvolucaoLojasResult struct {
	DiasJanela int                   `json:"dias_janela"`
	Lojas      []EvolucaoLoja        `json:"lojas"`
	Resumo     EvolucaoResumo        `json:"resumo"`
}

// EvolucaoLoja agrupa a evolução de uma loja específica.
type EvolucaoLoja struct {
	BuscaID        string                `json:"busca_id"`
	TotalProdutos  int                   `json:"total_produtos"`
	PrecoMedioAtual float64              `json:"preco_medio_atual"`
	PrecoMedioInicio float64             `json:"preco_medio_inicio"`
	VariacaoMedia  float64              `json:"variacao_media_pct"` // ex.: -0.05 = queda de 5%
	Coletas        int                   `json:"coletas"`           // número de snapshots na janela
	Pontos         []PontoEvolucao       `json:"pontos"`            // série temporal
	TopQuedas      []VariacaoPreco       `json:"top_quedas"`        // maiores quedas
	TopAltas       []VariacaoPreco       `json:"top_altas"`         // maiores altas
}

// PontoEvolucao é um ponto na série temporal (preço médio por dia/coleta).
type PontoEvolucao struct {
	Data       string  `json:"data"`        // YYYY-MM-DD
	PrecoMedio float64 `json:"preco_medio"`
	Produtos   int     `json:"produtos"`    // quantos produtos naquele dia
}

// EvolucaoResumo agrega métricas de todas as lojas monitoradas.
type EvolucaoResumo struct {
	TotalLojas        int     `json:"total_lojas"`
	TotalProdutos     int     `json:"total_produtos"`
	PrecoMedioGlobal  float64 `json:"preco_medio_global"`
	VariacaoMediaGlobal float64 `json:"variacao_media_global_pct"`
	TotalQuedas       int     `json:"total_quedas"`
	TotalAltas        int     `json:"total_altas"`
}

// Publicacao representa uma publicação agendada ou executada (espelhado do publish).
type Publicacao struct {
	ID         string  `json:"id"`
	ProdutoID  string  `json:"produto_id"`
	Nome       string  `json:"nome"`
	Categoria  string  `json:"categoria"`
	Preco      float64 `json:"preco"`
	Comissao   float64 `json:"comissao"`
	Link       string  `json:"link"`
	Imagem     string  `json:"imagem"`
	Estrategia string  `json:"estrategia"`
	DestinoID  string  `json:"destino_id"`
	TemplateID string  `json:"template_id"`
	AgendadaEm string  `json:"agendada_em"`
	Status     string  `json:"status"`
	Detalhe    string  `json:"detalhe"`
	CriadaEm   string  `json:"criada_em"`
	EnviadaEm  string  `json:"enviada_em,omitempty"`
	OwnerUID   string  `json:"owner_uid,omitempty"`
}

// ConversaoResumo agrupa publicações por canal+sub_id, mostrando volume e
// potencial de conversão. Quando o webhook de conversão estiver ativo,
// o campo Conversoes será preenchido com dados reais.
type ConversaoResumo struct {
	Canal        string  `json:"canal"`
	SubID        string  `json:"sub_id"`
	Publicacoes  int     `json:"publicacoes"`
	ProdutoID    string  `json:"produto_id"`
	Nome         string  `json:"nome"`
	Estrategia   string  `json:"estrategia"`
	Preco        float64 `json:"preco"`
	ComissaoEst  float64 `json:"comissao_estimada"` // comissao * preco * publicacoes
	PublicadoEm  string  `json:"publicado_em"`      // data mais recente
}

// ColetaResumo é um registro resumido de uma coleta executada.
type ColetaResumo struct {
	ColetadoEm time.Time `json:"coletado_em"`
	Keyword    string    `json:"keyword"`
	Categoria  string    `json:"categoria"`
	Estrategia string    `json:"estrategia"`
	Produtos   int       `json:"produtos"`
}

// ItemSnapshot é um produto na foto de mercado de uma categoria, num instante.
type ItemSnapshot struct {
	Posicao   int
	ProdutoID string
	Nome      string
	Preco     float64
	Comissao  float64
	Vendas    int
	Nota      float64
	Score     float64
}

// Snapshot é a foto periódica de uma categoria: os top N do momento. É o que
// permite, depois, analisar o impacto das campanhas contra o pano de fundo do
// mercado (preço/comissão/demanda mudaram em volta da publicação?).
type Snapshot struct {
	Categoria  string
	Keyword    string
	Estrategia string
	Em         time.Time
	Itens      []ItemSnapshot
}

// NopStore descarta eventos — usado localmente e quando o BigQuery não está ligado.
type NopStore struct{}

func (NopStore) Registrar(context.Context, Evento) error           { return nil }
func (NopStore) RegistrarSnapshot(context.Context, Snapshot) error { return nil }
func (NopStore) Estatisticas(_ context.Context, dias int) (Estatisticas, error) {
	// Sem persistência local: devolve um resumo vazio, deixando claro a fonte.
	return Estatisticas{Fonte: "nop", DiasJanela: dias, GeradoEm: time.Now().UTC()}, nil
}

// SalvarBusca/ListarBuscas no Nop são no-op: localmente, as buscas vivem no
// navegador (localStorage). O sync server-side só acontece com o BigQuery ligado.
func (NopStore) SalvarBusca(context.Context, Busca) error            { return nil }
func (NopStore) ListarBuscas(context.Context) ([]Busca, error)       { return nil, nil }
func (NopStore) HistoricoColetas(context.Context, int) ([]ColetaResumo, error) {
	return nil, nil
}
func (NopStore) Conversoes(context.Context, int) ([]ConversaoResumo, error) {
	return nil, nil
}
func (NopStore) SalvarPublicacao(context.Context, Publicacao) error              { return nil }
func (NopStore) ListarPublicacoes(context.Context, string) ([]Publicacao, error) { return nil, nil }
func (NopStore) AtualizarPublicacao(context.Context, string, string, string) error { return nil }
func (NopStore) Novidades(_ context.Context, buscaID string, dias int) (NovidadesLojas, error) {
	return NovidadesLojas{BuscaID: buscaID, DiasJanela: dias}, nil
}
func (NopStore) EvolucaoLojas(_ context.Context, dias int) (EvolucaoLojasResult, error) {
	return EvolucaoLojasResult{DiasJanela: dias}, nil
}
func (NopStore) EnsureSchema(context.Context) error { return nil }
func (NopStore) Nome() string                       { return "nop" }
