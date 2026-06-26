package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

// syncConversoes consulta o conversionReport da Shopee e grava conversões reais.
// Chamado periodicamente pelo Cloud Scheduler (1x/dia).
// POST /api/conversoes/sync
func (srv *Server) syncConversoes(w http.ResponseWriter, r *http.Request) {
	if !srv.autorizadoColeta(r) {
		writeErr(w, http.StatusUnauthorized, "token inválido")
		return
	}

	appID := os.Getenv("SHOPEE_APP_ID")
	secret := os.Getenv("SHOPEE_SECRET")
	if appID == "" || secret == "" {
		writeErr(w, http.StatusBadGateway, "SHOPEE_APP_ID/SECRET não configurados")
		return
	}

	dias := 7
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}

	// Consulta conversionReport da Shopee
	conversoes, err := buscarConversoeShopee(appID, secret, dias)
	if err != nil {
		srv.Logger.Error("sync conversoes falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, "falha ao consultar Shopee: "+err.Error())
		return
	}

	srv.Logger.Info("sync conversoes",
		slog.Int("encontradas", len(conversoes)),
		slog.Int("dias", dias),
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"conversoes": len(conversoes),
		"dias":       dias,
		"dados":      conversoes,
	})
}

// shopeeConversao representa uma conversão retornada pelo conversionReport.
type shopeeConversao struct {
	ConversionID    string  `json:"conversion_id"`
	UtmContent      string  `json:"utm_content"` // é o sub_id
	TotalCommission float64 `json:"total_commission"`
	Status          string  `json:"status"` // UNPAID, PENDING, COMPLETED, CANCELLED
	ClickTime       string  `json:"click_time"`
	PurchaseTime    string  `json:"purchase_time"`
	Orders          []struct {
		Items []struct {
			ItemID              string  `json:"item_id"`
			ItemName            string  `json:"item_name"`
			ItemTotalCommission float64 `json:"item_total_commission"`
		} `json:"items"`
	} `json:"orders"`
}

// buscarConversoeShopee consulta o conversionReport da API de afiliados.
func buscarConversoeShopee(appID, secret string, dias int) ([]shopeeConversao, error) {
	// A query do conversionReport usa purchaseTime como filtro
	agora := time.Now().UTC()
	inicio := agora.AddDate(0, 0, -dias)

	query := fmt.Sprintf(`{
		conversionReport(
			purchaseTimeFrom: "%s",
			purchaseTimeTo: "%s"
		) {
			nodes {
				conversionId
				utmContent
				totalCommission
				status
				clickTime
				purchaseTime
				orders {
					items {
						itemId
						itemName
						itemTotalCommission
					}
				}
			}
			pageInfo {
				scrollId
				hasNextPage
			}
		}
	}`, inicio.Format("2006-01-02"), agora.Format("2006-01-02"))

	body, _ := json.Marshal(map[string]string{"query": query})
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sum := sha256.Sum256([]byte(appID + ts + string(body) + secret))
	sig := hex.EncodeToString(sum[:])

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, "https://open-api.affiliate.shopee.com.br/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", appID, ts, sig))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requisição falhou: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gql struct {
		Data struct {
			ConversionReport struct {
				Nodes []struct {
					ConversionID    string  `json:"conversionId"`
					UtmContent      string  `json:"utmContent"`
					TotalCommission float64 `json:"totalCommission"`
					Status          string  `json:"status"`
					ClickTime       string  `json:"clickTime"`
					PurchaseTime    string  `json:"purchaseTime"`
					Orders          []struct {
						Items []struct {
							ItemID              string  `json:"itemId"`
							ItemName            string  `json:"itemName"`
							ItemTotalCommission float64 `json:"itemTotalCommission"`
						} `json:"items"`
					} `json:"orders"`
				} `json:"nodes"`
			} `json:"conversionReport"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(raw, &gql); err != nil {
		return nil, fmt.Errorf("resposta inválida: %w", err)
	}
	if len(gql.Errors) > 0 {
		return nil, fmt.Errorf("API: %s", gql.Errors[0].Message)
	}

	var result []shopeeConversao
	for _, n := range gql.Data.ConversionReport.Nodes {
		c := shopeeConversao{
			ConversionID:    n.ConversionID,
			UtmContent:      n.UtmContent,
			TotalCommission: n.TotalCommission,
			Status:          n.Status,
			ClickTime:       n.ClickTime,
			PurchaseTime:    n.PurchaseTime,
		}
		result = append(result, c)
	}

	return result, nil
}
