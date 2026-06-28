package store

// Destino é um canal de publicação: um grupo de Telegram, um número de
// WhatsApp, etc. É o que a usuária cadastra no dashboard.
type Destino struct {
	ID     string `json:"id"`     // slug único (ex.: "ofertas-beleza")
	Nome   string `json:"nome"`   // nome amigável
	Tipo   string `json:"tipo"`   // "telegram" | "whatsapp" | ...
	Config string `json:"config"` // dado específico do provedor (chat_id, telefone, ...)
	Ativo  bool   `json:"ativo"`
}
