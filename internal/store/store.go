// Package store registra eventos (decisões de curadoria, e depois conversões)
// para análise posterior. A interface isola o destino: NopStore em dev/local,
// BigQueryStore em produção (atrás da build tag `gcp`, para não pesar o build
// padrão nem o CI com a dependência do BigQuery).
package store

import (
	"context"
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
	// ID é a chave primária da busca (ex.: "perfumaria-japonesa"). Gerado
	// automaticamente como slug da primeira keyword se não fornecido.
	ID          string    `json:"id"`
	Keywords    []string  `json:"keywords"`             // um ou mais termos de busca
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

	// Legado: campo keyword como string única. Lido na deserialização mas
	// convertido para Keywords imediatamente pelo normalizador.
	KeywordLegado string `json:"keyword,omitempty"`
	// NomeLegado preserva compatibilidade com perfis antigos que usavam nome livre.
	NomeLegado string `json:"nome,omitempty"`
}

// NormalizarBusca garante que a busca tenha um ID e que Keywords esteja
// preenchida. Converte campos legados (nome/keyword como string).
func NormalizarBusca(b Busca) Busca {
	// compatibilidade com o modelo antigo que usava keyword (string única)
	if len(b.Keywords) == 0 && b.KeywordLegado != "" {
		b.Keywords = []string{b.KeywordLegado}
	}
	// compatibilidade: nome livre vira ID se ainda não tiver
	if b.ID == "" && b.NomeLegado != "" {
		b.ID = slugificar(b.NomeLegado)
	}
	// se ainda não tem ID, usa o slug da primeira keyword
	if b.ID == "" && len(b.Keywords) > 0 {
		b.ID = slugificar(b.Keywords[0])
	}
	// estratégia padrão
	if b.Estrategia == "" {
		b.Estrategia = "nicho"
	}
	// limpa legados para não gerar ruído no JSON de resposta
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
	// Estatisticas devolve o resumo descritivo dos snapshots dos últimos `dias`.
	Estatisticas(ctx context.Context, dias int) (Estatisticas, error)
	// SalvarBusca persiste (append) um perfil de busca; ListarBuscas devolve o
	// estado atual (último registro por ID, só os ativos).
	SalvarBusca(ctx context.Context, b Busca) error
	ListarBuscas(ctx context.Context) ([]Busca, error)
	// HistoricoColetas retorna os snapshots agrupados por keyword/data nos últimos `dias`.
	HistoricoColetas(ctx context.Context, dias int) ([]ColetaResumo, error)
	// EnsureSchema cria as tabelas do BigQuery se ainda não existirem.
	// Idempotente — seguro chamar no startup toda vez.
	EnsureSchema(ctx context.Context) error
	Nome() string
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
func (NopStore) EnsureSchema(context.Context) error { return nil }
func (NopStore) Nome() string                       { return "nop" }
