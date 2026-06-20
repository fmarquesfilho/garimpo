package source

import (
	"encoding/json"
	"testing"
)

type amostraFlex struct {
	F flexFloat  `json:"f"`
	I flexInt    `json:"i"`
	S flexString `json:"s"`
}

func TestFlexAceitaStringENumero(t *testing.T) {
	casos := []struct {
		nome string
		in   string
		f    float64
		i    int
		s    string
	}{
		{"string", `{"f":"0.085","i":"12","s":"abc"}`, 0.085, 12, "abc"},
		{"numero", `{"f":0.085,"i":12,"s":"abc"}`, 0.085, 12, "abc"},
		{"int_via_string_decimal", `{"f":"1.5","i":"3.0","s":"x"}`, 1.5, 3, "x"},
		{"vazio", `{"f":"","i":"","s":""}`, 0, 0, ""},
		{"nulo", `{"f":null,"i":null,"s":"y"}`, 0, 0, "y"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var a amostraFlex
			if err := json.Unmarshal([]byte(c.in), &a); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if float64(a.F) != c.f {
				t.Errorf("F=%v, quer %v", float64(a.F), c.f)
			}
			if int(a.I) != c.i {
				t.Errorf("I=%v, quer %v", int(a.I), c.i)
			}
			if string(a.S) != c.s {
				t.Errorf("S=%q, quer %q", string(a.S), c.s)
			}
		})
	}
}

func TestFlexFloatInvalido(t *testing.T) {
	var a amostraFlex
	if err := json.Unmarshal([]byte(`{"f":"abc"}`), &a); err == nil {
		t.Error("string não-numérica em flexFloat deveria falhar")
	}
}
