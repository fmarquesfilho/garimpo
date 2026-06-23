package logs

import (
	"sync"
	"time"
)

// Entry representa um log capturado para o dashboard.
type Entry struct {
	Nivel   string    `json:"nivel"`
	Msg     string    `json:"msg"`
	Metodo  string    `json:"metodo,omitempty"`
	Rota    string    `json:"rota,omitempty"`
	Status  int       `json:"status,omitempty"`
	DurMs   float64   `json:"dur_ms,omitempty"`
	Detalhe string    `json:"detalhe,omitempty"`
	Em      time.Time `json:"em"`
}

// Buffer é um ring buffer thread-safe que guarda os últimos N logs.
type Buffer struct {
	mu      sync.RWMutex
	entries []Entry
	max     int
	pos     int
	total   int
}

// NovoBuffer cria um buffer com capacidade máxima.
func NovoBuffer(max int) *Buffer {
	if max <= 0 {
		max = 500
	}
	return &Buffer{
		entries: make([]Entry, max),
		max:     max,
	}
}

// Push adiciona uma entrada ao buffer (sobrescreve a mais antiga se cheio).
func (b *Buffer) Push(e Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e.Em.IsZero() {
		e.Em = time.Now().UTC()
	}
	b.entries[b.pos] = e
	b.pos = (b.pos + 1) % b.max
	b.total++
}

// Ultimos retorna as últimas n entradas (mais recente primeiro).
func (b *Buffer) Ultimos(n int) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := b.total
	if count > b.max {
		count = b.max
	}
	if n <= 0 || n > count {
		n = count
	}

	out := make([]Entry, 0, n)
	idx := (b.pos - 1 + b.max) % b.max
	for i := 0; i < n; i++ {
		out = append(out, b.entries[idx])
		idx = (idx - 1 + b.max) % b.max
	}
	return out
}

// Total retorna o número total de logs recebidos (incluindo os descartados).
func (b *Buffer) Total() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.total
}

// Stats retorna contagem por nível dos logs no buffer.
func (b *Buffer) Stats() map[string]int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	stats := map[string]int{"info": 0, "warn": 0, "error": 0, "debug": 0}
	count := b.total
	if count > b.max {
		count = b.max
	}
	idx := (b.pos - 1 + b.max) % b.max
	for i := 0; i < count; i++ {
		stats[b.entries[idx].Nivel]++
		idx = (idx - 1 + b.max) % b.max
	}
	return stats
}
