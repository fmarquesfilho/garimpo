// Package problem implementa RFC 9457 — Problem Details for HTTP APIs.
// Fornece um formato padronizado e machine-readable para erros HTTP,
// com campos consistentes que facilitam tratamento no frontend.
//
// Formato de resposta:
//
//	{
//	  "type":     "https://garimpei.app.br/problemas/shopee-indisponivel",
//	  "title":    "API da Shopee indisponível",
//	  "status":   502,
//	  "detail":   "A API de afiliados retornou timeout após 20s.",
//	  "instance": "/api/candidatos?keyword=skincare"
//	}
//
// Content-Type: application/problem+json (RFC 9457 §3)
package problem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// baseURI é o prefixo para types de problemas do Garimpei.
const baseURI = "https://garimpei.app.br/problemas"

// Details é a estrutura conforme RFC 9457, §3.1.
type Details struct {
	// Type identifica o tipo de problema (URI). Default: "about:blank".
	Type string `json:"type"`
	// Title é um resumo curto e estável do tipo de problema (não muda por ocorrência).
	Title string `json:"title"`
	// Status é o HTTP status code (advisory, deve bater com o status real).
	Status int `json:"status"`
	// Detail é uma explicação específica desta ocorrência (para o usuário).
	Detail string `json:"detail,omitempty"`
	// Instance identifica a ocorrência específica (URI do request).
	Instance string `json:"instance,omitempty"`

	// ── Extensões (campos adicionais específicos do Garimpei) ──
	// Código interno para o frontend mapear ações.
	Code string `json:"code,omitempty"`
	// Retry indica se o client pode tentar novamente.
	Retry bool `json:"retry,omitempty"`
}

// Write serializa o Problem Details na resposta HTTP.
func Write(w http.ResponseWriter, p Details) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// ── Construtores para problemas comuns ───────────────────────────────────────

// New cria um Problem Details genérico.
func New(status int, title, detail string) Details {
	return Details{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// BadRequest (400) — input inválido do usuário.
func BadRequest(detail string) Details {
	return Details{
		Type:   baseURI + "/entrada-invalida",
		Title:  "Dados inválidos",
		Status: http.StatusBadRequest,
		Detail: detail,
		Code:   "entrada_invalida",
	}
}

// Unauthorized (401) — não autenticado.
func Unauthorized(detail string) Details {
	if detail == "" {
		detail = "Faça login para acessar este recurso."
	}
	return Details{
		Type:   baseURI + "/nao-autenticado",
		Title:  "Não autenticado",
		Status: http.StatusUnauthorized,
		Detail: detail,
		Code:   "nao_autenticado",
	}
}

// Forbidden (403) — autenticado mas sem permissão.
func Forbidden(detail string) Details {
	if detail == "" {
		detail = "Você não tem permissão para acessar este recurso."
	}
	return Details{
		Type:   baseURI + "/sem-permissao",
		Title:  "Acesso negado",
		Status: http.StatusForbidden,
		Detail: detail,
		Code:   "sem_permissao",
	}
}

// NotFound (404) — recurso não existe.
func NotFound(detail string) Details {
	return Details{
		Type:   baseURI + "/nao-encontrado",
		Title:  "Não encontrado",
		Status: http.StatusNotFound,
		Detail: detail,
		Code:   "nao_encontrado",
	}
}

// Conflict (409) — conflito de estado (ex: duplicata).
func Conflict(detail string) Details {
	return Details{
		Type:   baseURI + "/conflito",
		Title:  "Conflito",
		Status: http.StatusConflict,
		Detail: detail,
		Code:   "conflito",
	}
}

// ExternalService (502) — serviço externo falhou.
func ExternalService(service, detail string) Details {
	return Details{
		Type:   baseURI + "/servico-externo",
		Title:  fmt.Sprintf("Falha no serviço: %s", service),
		Status: http.StatusBadGateway,
		Detail: detail,
		Code:   "servico_externo",
		Retry:  true,
	}
}

// ServiceUnavailable (503) — serviço temporariamente indisponível.
func ServiceUnavailable(detail string) Details {
	return Details{
		Type:   baseURI + "/indisponivel",
		Title:  "Serviço temporariamente indisponível",
		Status: http.StatusServiceUnavailable,
		Detail: detail,
		Code:   "indisponivel",
		Retry:  true,
	}
}

// Internal (500) — erro interno não esperado.
func Internal(detail string) Details {
	return Details{
		Type:   baseURI + "/erro-interno",
		Title:  "Erro interno",
		Status: http.StatusInternalServerError,
		Detail: detail,
		Code:   "erro_interno",
	}
}

// WithInstance adiciona o campo instance (URI do request).
func (p Details) WithInstance(instance string) Details {
	p.Instance = instance
	return p
}
