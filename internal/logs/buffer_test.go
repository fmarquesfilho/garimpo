package logs

import "testing"

func TestBufferPushEUltimos(t *testing.T) {
	b := NovoBuffer(5)

	for i := 0; i < 3; i++ {
		b.Push(Entry{Nivel: "info", Rota: "/test"})
	}

	got := b.Ultimos(10)
	if len(got) != 3 {
		t.Errorf("esperava 3, veio %d", len(got))
	}
	if b.Total() != 3 {
		t.Errorf("total deveria ser 3, veio %d", b.Total())
	}
}

func TestBufferOverflow(t *testing.T) {
	b := NovoBuffer(3)

	for i := 0; i < 10; i++ {
		b.Push(Entry{Nivel: "info", Msg: string(rune('a' + i))})
	}

	// Só os últimos 3 devem estar no buffer
	got := b.Ultimos(10)
	if len(got) != 3 {
		t.Errorf("buffer cheio deveria ter 3 entradas, veio %d", len(got))
	}
	if b.Total() != 10 {
		t.Errorf("total deveria ser 10 (inclui descartados), veio %d", b.Total())
	}
}

func TestBufferUltimosLimita(t *testing.T) {
	b := NovoBuffer(100)
	for i := 0; i < 50; i++ {
		b.Push(Entry{Nivel: "info"})
	}

	got := b.Ultimos(5)
	if len(got) != 5 {
		t.Errorf("deveria retornar 5, veio %d", len(got))
	}
}

func TestBufferStats(t *testing.T) {
	b := NovoBuffer(100)
	b.Push(Entry{Nivel: "info"})
	b.Push(Entry{Nivel: "info"})
	b.Push(Entry{Nivel: "error"})
	b.Push(Entry{Nivel: "warn"})

	stats := b.Stats()
	if stats["info"] != 2 {
		t.Errorf("info deveria ser 2, veio %d", stats["info"])
	}
	if stats["error"] != 1 {
		t.Errorf("error deveria ser 1, veio %d", stats["error"])
	}
	if stats["warn"] != 1 {
		t.Errorf("warn deveria ser 1, veio %d", stats["warn"])
	}
}

func TestBufferVazio(t *testing.T) {
	b := NovoBuffer(10)
	got := b.Ultimos(5)
	if len(got) != 0 {
		t.Errorf("buffer vazio deveria retornar 0, veio %d", len(got))
	}
}
