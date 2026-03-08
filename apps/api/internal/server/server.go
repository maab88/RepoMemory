package server

import (
	"fmt"
	"net/http"

	"github.com/maab88/repomemory/apps/api/internal/config"
)

func New(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: handler,
	}
}
