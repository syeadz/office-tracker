// Package app handles application initialization and dependency wiring.
package app

import (
	"database/sql"

	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
)

// Services holds all application services.
type Services struct {
	Attendance  *service.AttendanceService
	User        *service.UserService
	Session     *service.SessionService
	Stats       *service.OfficeStatsService
	Scheduler   *service.SchedulerService
	Reports     *service.ReportsService
	Environment *service.EnvironmentService
}

// Repositories holds all data access repositories.
type Repositories struct {
	User    *repository.UserRepo
	Session *repository.SessionRepo
}

// App encapsulates all application components.
type App struct {
	Repos    *Repositories
	Services *Services
	HTTP     *httptransport.Server
}

// New initializes all application components with dependency injection.
func New(db *sql.DB, httpPort string, mwConfig httptransport.MiddlewareConfig) *App {
	// Initialize repositories
	repos := &Repositories{
		User:    &repository.UserRepo{DB: db},
		Session: &repository.SessionRepo{DB: db},
	}

	// Initialize services
	statsService := &service.OfficeStatsService{
		Sessions: repos.Session,
	}

	services := &Services{
		Attendance: service.NewAttendanceService(repos.User, repos.Session),
		User: &service.UserService{
			Users: repos.User,
		},
		Session: &service.SessionService{
			Sessions: repos.Session,
		},
		Scheduler:   service.NewSchedulerService(repos.Session),
		Stats:       statsService,
		Environment: service.NewEnvironmentService(service.DefaultEnvironmentMaxAge),
	}

	// Initialize HTTP server
	httpServer := httptransport.New(httpPort, services.Attendance, services.User, services.Session, services.Stats, services.Environment, nil, mwConfig)

	return &App{
		Repos:    repos,
		Services: services,
		HTTP:     httpServer,
	}
}
