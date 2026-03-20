package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vjelinekk/it-is-one.GO/pkg/api"
)

type Server struct {
	addr   string
	router *chi.Mux
}

func New(addr string) *Server {
	s := &Server{
		addr:   addr,
		router: chi.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// A good base middleware stack
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Set a timeout value on the request context
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Healthcheck endpoint
	s.router.Get("/health", api.HealthCheckHandler)

	// Example of another endpoint
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Welcome to the API!"))
	})
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.router)
}
