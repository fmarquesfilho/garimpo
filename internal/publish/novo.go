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
func NovoComDestinos(destinos DestinoStore) Publicador {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatPadrao := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" {
		return NovoMock("telegram")
	}
	if destinos == nil {
		destinos = NovoMemDestinoStore()
	}
	return NovoDispatcher(
		DispatcherConfig{
			Destinos:     destinos,
			TipoPadrao:   "telegram",
			ConfigPadrao: chatPadrao, // pode ser vazio — nesse caso, exige destino explícito
		},
		NovoTelegramSender(token),
	)
}
