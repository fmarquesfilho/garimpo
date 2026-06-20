package publish

import "os"

// Novo escolhe o publicador: Telegram de verdade se TELEGRAM_BOT_TOKEN e
// TELEGRAM_CHAT_ID estiverem definidos; caso contrário, o Mock (padrão).
// Assim você roda mockado agora e, ao preencher as variáveis, passa a publicar
// de verdade sem mudar mais nada.
func Novo() Publicador {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chat := os.Getenv("TELEGRAM_CHAT_ID")
	if token != "" && chat != "" {
		return NovoTelegram(token, chat)
	}
	return NovoMock("telegram")
}
