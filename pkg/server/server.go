package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	_ "github.com/vjelinekk/it-is-one.GO/docs"
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
	s.router.Use(cors.AllowAll().Handler)

	// Set a timeout value on the request context
	s.router.Use(middleware.Timeout(60 * time.Second))

	// API v1 Routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Public Routes (No Auth needed)
		r.Post("/users", api.NewUserHandler(s.db).Create)

		// Device Endpoints (accepts X-Device-Serial or X-User-ID)
		r.Group(func(dev chi.Router) {
			dev.Use(api.HeartbeatAuthMiddleware)
			dev.Post("/device/heartbeat", api.NewHardwareHandler(s.db).Heartbeat)
		})

		// Mobile Endpoints
		r.Group(func(mob chi.Router) {
			mob.Use(api.MobileAuthMiddleware)
			mobHandler := api.NewMobileHandler(s.db)

			// User & Device Linking
			mob.Patch("/users", api.NewUserHandler(s.db).Patch)

			// Schedules
			mob.Post("/schedules", mobHandler.CreateSchedule)
			mob.Patch("/schedules", mobHandler.PatchSchedule)
			mob.Get("/schedules", mobHandler.ListSchedules)
			mob.Delete("/schedules/{id}", mobHandler.DeleteSchedule)

			// Caregivers
			mob.Post("/caregivers", mobHandler.AddCaregiver)
			mob.Post("/caregivers/verify-phone", mobHandler.VerifyPhone)
			mob.Get("/caregivers", mobHandler.ListCaregivers)
			mob.Delete("/caregivers/{email}", mobHandler.DeleteCaregiver)

			// Intake Logs
			mob.Post("/intake-logs", api.NewIntakeLogHandler(s.db).LogIntake)
		})

	})
	// Healthcheck endpoint
	s.router.Get("/health", api.HealthCheckHandler)

	// Swagger UI
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.router)
}
