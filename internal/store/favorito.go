package store

import "time"

// Favorito é um produto salvo pelo usuário para análise posterior.
type Favorito struct {
	ProdutoID string    `json:"produto_id"`
	Nome      string    `json:"nome"`
	Preco     float64   `json:"preco"`
	Comissao  float64   `json:"comissao"`
	Link      string    `json:"link"`
	Imagem    string    `json:"imagem"`
	Loja      string    `json:"loja"`
	Categoria string    `json:"categoria"`
	Origem    string    `json:"origem,omitempty"`
	SalvoEm   time.Time `json:"salvo_em"`
	OwnerUID  string    `json:"owner_uid,omitempty"`
	Ativo     bool      `json:"ativo"`
}
