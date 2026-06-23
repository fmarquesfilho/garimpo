package auth

import (
	"os"
	"strings"
)

// AdminEmails retorna a lista de emails com papel admin.
// Configurada via ADMIN_EMAILS (separados por vírgula).
// Ex.: ADMIN_EMAILS=fernando@gmail.com,amigo@gmail.com
func AdminEmails() map[string]bool {
	raw := os.Getenv("ADMIN_EMAILS")
	if raw == "" {
		return nil
	}
	m := make(map[string]bool)
	for _, e := range strings.Split(raw, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			m[e] = true
		}
	}
	return m
}

// IsAdmin verifica se o email está na lista de admins.
func IsAdmin(email string) bool {
	admins := AdminEmails()
	if admins == nil {
		return false
	}
	return admins[strings.ToLower(strings.TrimSpace(email))]
}
