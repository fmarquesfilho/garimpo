package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
)

// whatsappGrupos lista os grupos do WhatsApp disponíveis na sessão do Maytapi.
// Usado pelo frontend para popular o select no cadastro de destinos.
func (srv *Server) whatsappGrupos(w http.ResponseWriter, r *http.Request) {

	productID := os.Getenv("WHATSAPP_PRODUCT_ID")
	phoneID := os.Getenv("WHATSAPP_PHONE_ID")
	token := os.Getenv("WHATSAPP_API_KEY")
	if productID == "" || phoneID == "" || token == "" {
		writeErr(w, http.StatusServiceUnavailable, "WhatsApp não configurado (WHATSAPP_PRODUCT_ID, WHATSAPP_PHONE_ID, WHATSAPP_API_KEY)")
		return
	}

	grupos, err := buscarGruposMaytapi(r, productID, phoneID, token)
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

// buscarGruposMaytapi chama GET /{phone_id}/getGroups na API do Maytapi.
func buscarGruposMaytapi(r *http.Request, productID, phoneID, token string) ([]grupoWA, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://api.maytapi.com/api/%s/%s/getGroups", productID, phoneID)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("maytapi criar request: %w", apperr.ErrMaytapi)
	}
	req.Header.Set("x-maytapi-key", token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("maytapi getGroups: %w", apperr.ErrMaytapi)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("maytapi HTTP %d: %w", resp.StatusCode, apperr.ErrMaytapi)
	}

	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("maytapi decodificar resposta: %w", err)
	}
	if !body.Success {
		return nil, fmt.Errorf("maytapi resposta não-success: %w", apperr.ErrMaytapi)
	}

	grupos := make([]grupoWA, 0, len(body.Data))
	for _, g := range body.Data {
		nome := g.Name
		if nome == "" {
			nome = g.ID
		}
		grupos = append(grupos, grupoWA{ID: g.ID, Nome: nome})
	}
	return grupos, nil
}
