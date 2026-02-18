package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/arihershowitz/translate-xhtml-local/internal/translator"
)

// TranslationRequest represents the request body for translation.
type TranslationRequest struct {
	XHTML      string `json:"xhtml" binding:"required"`
	SourceLang string `json:"source_lang" binding:"required"`
	TargetLang string `json:"target_lang" binding:"required"`
}

// TranslationResponse represents the response body for translation.
type TranslationResponse struct {
	TranslatedXHTML string              `json:"translated_xhtml"`
	Metadata        translator.Metadata `json:"metadata"`
}

// Handler handles API requests.
type Handler struct {
	service translator.TranslationService
}

// NewHandler creates a new API handler.
func NewHandler(service translator.TranslationService) *Handler {
	return &Handler{service: service}
}

// Translate godoc
// @Summary Translate XHTML content
// @Description Translates XHTML content from source language to target language using a local LLM.
// @Tags translation
// @Accept json
// @Produce json
// @Param request body TranslationRequest true "Translation Request"
// @Success 200 {object} TranslationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /translate [post]
func (h *Handler) Translate(w http.ResponseWriter, r *http.Request) {
	var req TranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.XHTML == "" || req.SourceLang == "" || req.TargetLang == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	translated, metadata, err := h.service.Translate(ctx, strings.NewReader(req.XHTML), req.SourceLang, req.TargetLang)
	if err != nil {
		http.Error(w, "Translation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := TranslationResponse{
		TranslatedXHTML: translated,
		Metadata:        metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
