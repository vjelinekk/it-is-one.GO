package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	"github.com/vjelinekk/it-is-one.GO/pkg/api"
	_ "github.com/vjelinekk/it-is-one.GO/pkg/server/docs" // Import generated docs
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

	// API v1 Routes
	s.router.Route("/api/v1", func(r chi.Router) {

		// Public Routes (No Auth needed)
		r.Post("/users", api.NewUserHandler(s.db).Create)

		// Hardware Endpoints

		r.Group(func(hw chi.Router) {
			hw.Use(api.HardwareAuthMiddleware)
			hwHandler := api.NewHardwareHandler(s.db)
			hw.Post("/device/heartbeat", hwHandler.Heartbeat)
			hw.Post("/device/intake", hwHandler.LogIntake)
		})

		// Mobile Endpoints
		r.Group(func(mob chi.Router) {
			mob.Use(api.MobileAuthMiddleware)
			mobHandler := api.NewMobileHandler(s.db)
			userHandler := api.NewUserHandler(s.db)

			// User & Device Linking
			// Specific routes like /me MUST come before wildcards like /{id}
			mob.Get("/users/me", mobHandler.GetMe)
			mob.Put("/users/me", mobHandler.UpdateMe)
			mob.Put("/users/me/device", mobHandler.LinkDevice)

			// Standard CRUD
			mob.Get("/users", userHandler.List)
			mob.Get("/users/{id}", userHandler.Get)
			mob.Put("/users/{id}", userHandler.Update)
			mob.Delete("/users/{id}", userHandler.Delete)

			// Schedules
			mob.Post("/schedules", mobHandler.CreateSchedule)
			mob.Get("/schedules", mobHandler.ListSchedules)
			mob.Delete("/schedules/{id}", mobHandler.DeleteSchedule)

			// Caregivers
			mob.Post("/caregivers", mobHandler.AddCaregiver)
			mob.Get("/caregivers", mobHandler.ListCaregivers)
			mob.Delete("/caregivers/{id}", mobHandler.DeleteCaregiver)

			// Notifications & Logs
			mob.Post("/push-tokens", mobHandler.RegisterPushToken)
			mob.Get("/intake-logs", mobHandler.ListIntakeLogs)
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
