package source

import "testing"

func TestInferirOrigemDeShopType(t *testing.T) {
	cases := []struct {
		shopTypes []string
		want      string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"mall"}, ""},
		{[]string{"preferred"}, ""},
		{[]string{"overseas"}, "Importado"},
		{[]string{"mall", "overseas"}, "Importado"},
		{[]string{"cb"}, "Importado"},
		{[]string{"cross_border"}, "Importado"},
		{[]string{"OVERSEAS"}, "Importado"}, // case insensitive
	}
	for _, tc := range cases {
		got := inferirOrigemDeShopType(tc.shopTypes)
		if got != tc.want {
			t.Errorf("inferirOrigemDeShopType(%v) = %q, want %q", tc.shopTypes, got, tc.want)
		}
	}
}

func TestNormalizarOrigem(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"Coreia", "Coreia"},
		{"korea", "Coreia"},
		{"Japão", "Japão"},
		{"japan", "Japão"},
		{"China", "China"},
		{"Importado", "Importado"},
		{"brasil", "Brasil"},
	}
	for _, tc := range cases {
		got := NormalizarOrigem(tc.input)
		if got != tc.want {
			t.Errorf("NormalizarOrigem(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
