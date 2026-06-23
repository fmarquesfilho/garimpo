package publish

import "os"

// Novo cria o publicador padrão: Dispatcher com TelegramSender se o token
// estiver configurado; caso contrário, Mock.
func Novo() Publicador {
	return NovoComDestinos(nil)
}

// NovoComDestinos cria o Dispatcher com suporte a múltiplos destinos.
// Se destinos for nil e o token existir, usa MemDestinoStore (dev).
// Se TELEGRAM_CHAT_ID estiver vazio, o Dispatcher funciona mas só publica
// para destinos cadastrados — sem fallback padrão.
//
// Senders registrados:
//   - Telegram: se TELEGRAM_BOT_TOKEN estiver definido
//   - WhatsApp: se WHATSAPP_API_KEY estiver definido (WaSenderAPI)
func NovoComDestinos(destinos DestinoStore) Publicador {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatPadrao := os.Getenv("TELEGRAM_CHAT_ID")
	waKey := os.Getenv("WHATSAPP_API_KEY")

	// Sem nenhum provedor configurado → mock
	if token == "" && waKey == "" {
		return NovoMock("telegram")
	}
	if destinos == nil {
		destinos = NovoMemDestinoStore()
	}

	// Monta lista de senders disponíveis
	var senders []Sender
	if token != "" {
		senders = append(senders, NovoTelegramSender(token))
	}
	if waKey != "" {
		senders = append(senders, NovoWhatsAppSender(waKey))
	}

	// Tipo padrão: Telegram se configurado, senão WhatsApp
	tipoPadrao := "telegram"
	if token == "" {
		tipoPadrao = "whatsapp"
		chatPadrao = "" // não faz sentido ter chat padrão sem Telegram
	}

	return NovoDispatcher(
		DispatcherConfig{
			Destinos:     destinos,
			TipoPadrao:   tipoPadrao,
			ConfigPadrao: chatPadrao, // pode ser vazio — nesse caso, exige destino explícito
		},
		senders...,
	)
}
