package publish

import "time"

// Publicacao representa uma publicação agendada ou já executada.
// Segue o mesmo padrão append-only do modelo de Busca.
type Publicacao struct {
	ID         string    `json:"id"`
	ProdutoID  string    `json:"produto_id"`
	Nome       string    `json:"nome"`
	Categoria  string    `json:"categoria"`
	Preco      float64   `json:"preco"`
	Comissao   float64   `json:"comissao"`
	Link       string    `json:"link"`
	Imagem     string    `json:"imagem"`
	Estrategia string    `json:"estrategia"`
	DestinoID  string    `json:"destino_id"`
	TemplateID string    `json:"template_id"`
	AgendadaEm string    `json:"agendada_em"` // ISO 8601 — quando deve ser enviada (vazio = imediata)
	Status     string    `json:"status"`      // "agendada" | "enviada" | "erro"
	Detalhe    string    `json:"detalhe"`     // mensagem de erro ou sub_id
	CriadaEm   time.Time `json:"criada_em"`
	EnviadaEm  string    `json:"enviada_em,omitempty"` // ISO 8601
	OwnerUID   string    `json:"owner_uid,omitempty"`
}
