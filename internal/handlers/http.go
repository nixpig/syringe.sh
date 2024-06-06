package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nixpig/syringe.sh/server/internal/services"
)

type HttpHandlers struct {
	appService services.AppService
}

func NewHttpHandlers(appService services.AppService) HttpHandlers {
	return HttpHandlers{appService}
}

func (h *HttpHandlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case http.MethodPost:
		var req services.RegisterUserRequestDto

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("decode create user request failed", "err", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		createdUser, err := h.appService.RegisterUser(req)
		if err != nil {
			slog.Error("create user failed", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(createdUser); err != nil {
			slog.Error("decode created user failed", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	default:
		slog.Error("method not allowed", "method", method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func (h *HttpHandlers) AddPublicKey(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case http.MethodPost:
		var req services.AddPublicKeyRequestDto

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("decode add public key request failed", "err", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		addedPublicKey, err := h.appService.AddPublicKey(req)
		if err != nil {
			slog.Error("add public key failed", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(addedPublicKey); err != nil {
			slog.Error("failed to encode and return added public key details", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	default:
		slog.Error("method not allowed", "method", method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
