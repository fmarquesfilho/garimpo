package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// whatsappGrupos lista os grupos do WhatsApp disponíveis na sessão da WaSenderAPI.
// Usado pelo frontend para popular o select no cadastro de destinos.
func (srv *Server) whatsappGrupos(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para listar grupos")
		return
	}

	apiKey := os.Getenv("WHATSAPP_API_KEY")
	if apiKey == "" {
		writeErr(w, http.StatusServiceUnavailable, "WHATSAPP_API_KEY não configurada")
		return
	}

	grupos, err := buscarGruposWaSender(r, apiKey)
	if err != nil {
		srv.Logger.Error("whatsapp grupos falhou", "erro", err.Error())
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"grupos": grupos})
}

// grupoWA representa um grupo do WhatsApp para o frontend.
type grupoWA struct {
	ID   string `json:"id"`   // group JID (ex.: 123456789-987654321@g.us)
	Nome string `json:"nome"` // nome do grupo
}

// buscarGruposWaSender chama GET /api/groups na WaSenderAPI.
func buscarGruposWaSender(r *http.Request, apiKey string) ([]grupoWA, error) {
	base := os.Getenv("WHATSAPP_API_BASE")
	if base == "" {
		base = "https://www.wasenderapi.com"
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, base+"/api/groups", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WaSenderAPI retornou HTTP %d", resp.StatusCode)
	}

	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Subject string `json:"subject"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}
	if !body.Success {
		return nil, fmt.Errorf("WaSenderAPI: resposta não-success")
	}

	grupos := make([]grupoWA, 0, len(body.Data))
	for _, g := range body.Data {
		nome := g.Name
		if nome == "" {
			nome = g.Subject
		}
		if nome == "" {
			nome = g.ID
		}
		grupos = append(grupos, grupoWA{ID: g.ID, Nome: nome})
	}
	return grupos, nil
}
