// Package tenant gerencia configurações por tenant (multi-tenancy leve).
// Cada usuário tem suas próprias credenciais de API, destinos de alerta, etc.
package tenant

import "github.com/fmarquesfilho/garimpo/internal/crypto"

// Encrypt delega para o pacote crypto compartilhado.
func Encrypt(plaintext string) (string, error) { return crypto.Encrypt(plaintext) }

// Decrypt delega para o pacote crypto compartilhado.
func Decrypt(ciphertext string) (string, error) { return crypto.Decrypt(ciphertext) }
