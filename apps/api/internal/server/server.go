package server

import (
	"fmt"
	"net/http"

	"github.com/maab88/repomemory/apps/api/internal/config"
	"github.com/maab88/repomemory/apps/api/internal/http/router"
)

func New(cfg config.Config) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router.New(),
	}
}
