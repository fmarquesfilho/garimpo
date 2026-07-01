package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
)

func TestSentinelsAreDistinct(t *testing.T) {
	sentinels := []error{
		apperr.ErrShopeeAPI,
		apperr.ErrTelegram,
		apperr.ErrWhatsApp,
		apperr.ErrInvalidInput,
		apperr.ErrNotFound,
		apperr.ErrInactive,
		apperr.ErrUnauthorized,
		apperr.ErrForbidden,
		apperr.ErrCrypto,
		apperr.ErrIO,
		apperr.ErrNoConfig,
		apperr.ErrTooManyRedirects,
		apperr.ErrNoProvider,
		apperr.ErrCSV,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinels %d e %d não devem ser iguais: %v == %v", i, j, a, b)
			}
		}
	}
}

func TestWrappedErrorsUnwrapWithErrorsIs(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		sentinel error
	}{
		{
			name:     "shopee wrapped",
			err:      fmt.Errorf("shopee api erro 10020 invalid signature: %w", apperr.ErrShopeeAPI),
			sentinel: apperr.ErrShopeeAPI,
		},
		{
			name:     "telegram wrapped",
			err:      fmt.Errorf("telegram bot bloqueado: %w", apperr.ErrTelegram),
			sentinel: apperr.ErrTelegram,
		},
		{
			name:     "crypto double-wrapped",
			err:      fmt.Errorf("decrypt base64: %w", fmt.Errorf("inner: %w", apperr.ErrCrypto)),
			sentinel: apperr.ErrCrypto,
		},
		{
			name:     "not found wrapped",
			err:      fmt.Errorf("destino %q: %w", "ofertas", apperr.ErrNotFound),
			sentinel: apperr.ErrNotFound,
		},
		{
			name:     "too many redirects",
			err:      apperr.ErrTooManyRedirects,
			sentinel: apperr.ErrTooManyRedirects,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.sentinel) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tc.err, tc.sentinel)
			}
		})
	}
}

func TestWrappedErrorPreservesContext(t *testing.T) {
	err := fmt.Errorf("shopee api erro 10020 invalid signature: %w", apperr.ErrShopeeAPI)

	// O erro contém a mensagem de contexto
	msg := err.Error()
	if msg != "shopee api erro 10020 invalid signature: shopee api" {
		t.Errorf("mensagem inesperada: %s", msg)
	}

	// Mas o sentinel é acessível via Is
	if !errors.Is(err, apperr.ErrShopeeAPI) {
		t.Error("errors.Is falhou para ErrShopeeAPI")
	}

	// E não confunde com outro sentinel
	if errors.Is(err, apperr.ErrTelegram) {
		t.Error("errors.Is não deveria casar com ErrTelegram")
	}
}
