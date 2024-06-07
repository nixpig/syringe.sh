package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/rs/zerolog"
)

type HttpHandlers struct {
	appService services.AppService
	log        zerolog.Logger
}

func NewHttpHandlers(appService services.AppService, log zerolog.Logger) HttpHandlers {
	return HttpHandlers{
		appService: appService,
		log:        log,
	}
}

func (h *HttpHandlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case http.MethodPost:
		var req services.RegisterUserRequestDto

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.log.Error().Err(err).Msg("decode create user request failed")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		createdUser, err := h.appService.RegisterUser(req)
		if err != nil {
			h.log.Error().Err(err).Msg("create user failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(createdUser); err != nil {
			h.log.Error().Err(err).Msg("decode created user failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		h.log.Info().Any("createdUser", createdUser).Msg("created user")

	default:
		h.log.Error().Str("method", method).Msg("method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func (h *HttpHandlers) AddPublicKey(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case http.MethodPost:
		var req services.AddPublicKeyRequestDto

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.log.Error().Err(err).Msg("decode add public key request failed")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		addedPublicKey, err := h.appService.AddPublicKey(req)
		if err != nil {
			h.log.Error().Err(err).Msg("add public key failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(addedPublicKey); err != nil {
			h.log.Error().Err(err).Msg("failed to encode and return added public key details")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		h.log.Info().Any("addedPublicKey", addedPublicKey).Msg("added public key")

	default:
		h.log.Error().Str("method", method).Msg("method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
