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
//   - WhatsApp: se WHATSAPP_ACCESS_TOKEN + WHATSAPP_PHONE_NUMBER_ID estiverem definidos (Meta Cloud API)
func NovoComDestinos(destinos DestinoStore) Publicador {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatPadrao := os.Getenv("TELEGRAM_CHAT_ID")
	waPhoneID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	waToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")

	// Sem nenhum provedor configurado → mock
	if token == "" && (waPhoneID == "" || waToken == "") {
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
	if waSender := NovoWhatsAppSenderFromEnv(); waSender != nil {
		senders = append(senders, waSender)
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
