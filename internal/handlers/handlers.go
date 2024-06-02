package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nixpig/syringe.sh/server/internal/user"
)

type HttpHandlers struct {
	userService user.UserService
}

func NewHttpHandlers(userService user.UserService) HttpHandlers {
	return HttpHandlers{userService}
}

func (h *HttpHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case http.MethodPost:
		var req user.RegisterUserRequestJsonDto

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("decode create user request", "err", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		createdUser, err := h.userService.Create(req)
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
	}
}
