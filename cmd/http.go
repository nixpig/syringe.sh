package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/rs/zerolog"
)

type SyringeHttpServer struct {
	handlers handlers.HttpHandlers
	log      *zerolog.Logger
}

func NewSyringeHttpServer(
	handlers handlers.HttpHandlers,
	log *zerolog.Logger,
) SyringeHttpServer {
	return SyringeHttpServer{
		handlers: handlers,
		log:      log,
	}

}

func (h SyringeHttpServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/users", h.handlers.RegisterUser)
	mux.HandleFunc("/keys", h.handlers.AddPublicKey)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", "3000"),
		Handler:      (mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	h.log.Info().Msg("starting http server")
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
