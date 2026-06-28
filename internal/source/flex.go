package source

import (
	"fmt"
	"strconv"
	"strings"
)

// A API de afiliados da Shopee às vezes devolve números como string ("0.0850")
// e às vezes como número JSON. Estes tipos flexíveis aceitam os dois formatos,
// evitando que a desserialização quebre por causa (ou ausência) de aspas.

type flexFloat float64

func (f *flexFloat) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("flexFloat parse %q: %w", s, err)
	}
	*f = flexFloat(v)
	return nil
}

type flexInt int

func (i *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*i = 0
		return nil
	}
	if v, err := strconv.Atoi(s); err == nil {
		*i = flexInt(v)
		return nil
	}
	fv, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("flexInt parse %q: %w", s, err)
	}
	*i = flexInt(int(fv))
	return nil
}

type flexString string

func (fs *flexString) UnmarshalJSON(b []byte) error {
	*fs = flexString(strings.Trim(string(b), `"`))
	return nil
}
