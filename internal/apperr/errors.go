// Package apperr define os erros sentinel do domínio Garimpei.
//
// Cada sentinel implementa Error() string e pode ser wrapped com
// fmt.Errorf("contexto: %w", apperr.ErrX) para adicionar detalhes sem
// perder o tipo raiz — permitindo errors.Is/As em qualquer camada.
//
// Categorias:
//   - Serviços externos (Shopee, Telegram, Maytapi)
//   - Validação e input do usuário
//   - Infraestrutura (crypto, I/O)
//   - Domínio (não encontrado, inativo, conflito)
package apperr

import "errors"

// ── Serviços externos ────────────────────────────────────────────────────────

// ErrShopeeAPI indica falha na comunicação com a API de afiliados da Shopee.
var ErrShopeeAPI = errors.New("shopee api")

// ErrTelegram indica falha na comunicação com o Telegram Bot API.
var ErrTelegram = errors.New("telegram")

// ErrWhatsApp indica falha na comunicação com a Meta WhatsApp Cloud API.
var ErrWhatsApp = errors.New("whatsapp")

// ── Validação / Input ────────────────────────────────────────────────────────

// ErrInvalidInput indica dado inválido fornecido pelo usuário.
var ErrInvalidInput = errors.New("entrada inválida")

// ErrNotFound indica recurso não encontrado.
var ErrNotFound = errors.New("não encontrado")

// ErrInactive indica recurso desabilitado/inativo.
var ErrInactive = errors.New("recurso inativo")

// ErrUnauthorized indica falta de autenticação.
var ErrUnauthorized = errors.New("não autenticado")

// ErrForbidden indica falta de permissão.
var ErrForbidden = errors.New("sem permissão")

// ── Infraestrutura ───────────────────────────────────────────────────────────

// ErrCrypto indica falha em operação criptográfica (AES, GCM, etc.).
var ErrCrypto = errors.New("crypto")

// ErrIO indica falha de I/O (leitura de arquivo, rede genérica).
var ErrIO = errors.New("i/o")

// ── Domínio ──────────────────────────────────────────────────────────────────

// ErrNoConfig indica credenciais/configuração não preenchidas.
var ErrNoConfig = errors.New("configuração ausente")

// ErrTooManyRedirects indica excesso de redirects ao resolver link.
var ErrTooManyRedirects = errors.New("muitos redirects")

// ErrNoProvider indica provedor de envio não registrado.
var ErrNoProvider = errors.New("provedor não registrado")

// ErrCSV indica erro no parsing de arquivo CSV.
var ErrCSV = errors.New("csv")

// ErrAmazonAPI indica falha na comunicação com a Amazon Creators API.
var ErrAmazonAPI = errors.New("amazon api")

// ErrRateLimited indica que a API retornou HTTP 429 (rate limiting).
var ErrRateLimited = errors.New("rate limited")
