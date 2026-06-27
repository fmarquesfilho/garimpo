package source

import "testing"

func TestCategoriaPorNome(t *testing.T) {
	tests := []struct {
		nome   string
		wantID int
	}{
		{"Perfumaria", 100640},
		{"perfumaria", 100640},
		{"Cuidados com a Pele", 100664},
		{"cuidados com a pele", 100664},
		{"Maquiagem", 100663},
		{"Beleza", 100630},
		{"inexistente", 0},
		{"", 0},
		{"  Perfumaria  ", 100640},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			got := CategoriaPorNome(tt.nome)
			if got != tt.wantID {
				t.Errorf("CategoriaPorNome(%q) = %d, want %d", tt.nome, got, tt.wantID)
			}
		})
	}
}

func TestNomeCategoriaPrincipal(t *testing.T) {
	tests := []struct {
		catIDs []int
		want   string
	}{
		{[]int{100640}, "Perfumaria"},
		{[]int{999, 100663}, "Maquiagem"},
		{[]int{999}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		got := NomeCategoriaPrincipal(tt.catIDs)
		if got != tt.want {
			t.Errorf("NomeCategoriaPrincipal(%v) = %q, want %q", tt.catIDs, got, tt.want)
		}
	}
}
