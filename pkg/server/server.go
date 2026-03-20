package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/vjelinekk/it-is-one.GO/pkg/api"
)

type Server struct {
	addr   string
	db     *gorm.DB
	router *chi.Mux
}

func New(addr string, db *gorm.DB) *Server {
	s := &Server{
		addr:   addr,
		db:     db,
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

	// User CRUD routes
	userHandler := api.NewUserHandler(s.db)
	s.router.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)      // Create
		r.Get("/", userHandler.List)        // Read All
		r.Get("/{id}", userHandler.Get)     // Read One
		r.Put("/{id}", userHandler.Update)  // Update
		r.Delete("/{id}", userHandler.Delete) // Delete
	})

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
