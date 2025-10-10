package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/opencodehq/opencode/pkg/agents"
	"github.com/opencodehq/opencode/pkg/models"
	"github.com/opencodehq/opencode/pkg/session"
	"github.com/opencodehq/opencode/pkg/tools"
)

type Server struct {
	modelMgr   *models.Manager
	sessionMgr *session.Manager
	toolReg    *tools.Registry
	mux        *http.ServeMux
	httpServer *http.Server
}

func New() *Server {
	toolReg := tools.NewRegistry()
	tools.RegisterDefaults(toolReg)
	modelMgr := models.NewManager()
	agents.RegisterDefaults(modelMgr, toolReg)
	return &Server{
		modelMgr:   modelMgr,
		sessionMgr: session.NewManager(),
		toolReg:    toolReg,
	}
}

func (s *Server) Start(ctx context.Context, addr string) error {
	if addr == "" {
		addr = ":8080"
	}
	s.registerRoutes()
	srv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	s.httpServer = srv
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
