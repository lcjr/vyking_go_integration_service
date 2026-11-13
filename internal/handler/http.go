package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"vyking-service/internal/domain"
	"vyking-service/internal/service"

	_ "vyking-service/internal/domain"
	_ "vyking-service/internal/service"

	"github.com/google/uuid"
)

type HTTPHandler struct {
	txService *service.TransactionService
	repo      domain.Repository // for auth lookup
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func NewHTTPHandler(txService *service.TransactionService, repo domain.Repository) *HTTPHandler {
	return &HTTPHandler{txService: txService, repo: repo}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.New().String()
	w.Header().Set("X-Request-ID", reqID)
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing bearer token", reqID)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	brandID, err := h.repo.GetBrandIDByPartnerToken(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusForbidden, "FORBIDDEN", "Invalid partner token", reqID)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "Email is required", reqID)
		return
	}

	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	txs, err := h.txService.GetRecentTransactions(r.Context(), brandID, email, status, limit)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), reqID)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"request_id":   reqID,
		"transactions": txs,
	})
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, code, msg, reqID string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:      code,
		Message:   msg,
		RequestID: reqID,
	})
}
