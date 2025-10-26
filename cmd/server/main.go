package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fanzru/social-media-service-go/infrastructure/config"
	accountApp "github.com/fanzru/social-media-service-go/internal/app/account/app"
	accountHTTP "github.com/fanzru/social-media-service-go/internal/app/account/port"
	"github.com/fanzru/social-media-service-go/internal/app/account/port/genhttp"
	"github.com/fanzru/social-media-service-go/internal/app/account/repo"
	healthApp "github.com/fanzru/social-media-service-go/internal/app/health/app"
	healthHTTP "github.com/fanzru/social-media-service-go/internal/app/health/port"
	healthGenHTTP "github.com/fanzru/social-media-service-go/internal/app/health/port/genhttp"
	healthRepo "github.com/fanzru/social-media-service-go/internal/app/health/repo"
	"github.com/fanzru/social-media-service-go/pkg/jwt"
	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/middleware"
	"github.com/fanzru/social-media-service-go/pkg/reqctx"
	"github.com/fanzru/social-media-service-go/pkg/sqlwrap"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	logger.InitFromEnv()
	log := logger.GetGlobal()

	// Load configuration
	cfg := config.Load()
	log.Info("Configuration loaded", "serverPort", cfg.Server.Port, "dbHost", cfg.Database.Host)

	// Build database connection string
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		// Build connection string from config
		dbConnStr = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
			cfg.Database.SSLMode,
		)
	}

	// Initialize database
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Error("Failed to open database", "error", err.Error())
		os.Exit(1)
	}

	// Test the connection (skip for now)
	// if err := db.Ping(); err != nil {
	// 	log.Error("Failed to ping database", "error", err.Error())
	// 	os.Exit(1)
	// }

	log.Info("PostgreSQL database connected successfully (ping skipped)", "host", cfg.Database.Host, "port", cfg.Database.Port, "database", cfg.Database.DBName)
	defer db.Close()

	// Wrap database with logging if enabled
	var dbInterface interface{} = db
	if cfg.Database.LogQueries {
		loggedDB := sqlwrap.NewDB(db)
		dbInterface = loggedDB
		log.Info("Database query logging enabled", "slowQueryThreshold", cfg.Database.SlowQueryThreshold)
	}

	// Initialize JWT service
	jwtService := jwt.NewService(cfg.JWT.Secret, time.Duration(cfg.JWT.Expiration)*time.Hour)
	log.Info("JWT service initialized")

	// Initialize account repository and service
	accountRepository := repo.NewRepository(dbInterface)
	log.Info("Account repository initialized")

	accountService := accountApp.NewService(accountRepository, jwtService)
	log.Info("Account service initialized")

	accountHandler := accountHTTP.NewHandler(accountService)
	log.Info("Account HTTP handler initialized")

	// Initialize health repository and service
	healthRepository := healthRepo.NewRepository(dbInterface)
	log.Info("Health repository initialized")

	healthService := healthApp.NewService(healthRepository)
	log.Info("Health service initialized")

	healthHandler := healthHTTP.NewHandler(healthService)
	log.Info("Health HTTP handler initialized")

	// Initialize middleware
	loggingMiddleware := middleware.LoggingMiddleware()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Add security requirements manually for now
	authMiddleware.AddSecurityRequirement("GET", "/api/account/profile", true)
	log.Info("Security requirements loaded manually")

	// Create OpenAPI server with middleware
	apiHandler := genhttp.Handler(accountHandler)

	// Setup routes using generated OpenAPI server with comprehensive middleware
	http.Handle("/api/",
		reqctx.Middleware(
			loggingMiddleware(
				authMiddleware.Middleware()(apiHandler),
			),
		),
	)

	// Create health OpenAPI server with middleware
	healthApiHandler := healthGenHTTP.Handler(healthHandler)

	// Add health check endpoints with logging middleware only (no auth required)
	http.Handle("/health",
		reqctx.Middleware(
			loggingMiddleware(healthApiHandler),
		),
	)
	http.Handle("/health/",
		reqctx.Middleware(
			loggingMiddleware(healthApiHandler),
		),
	)

	// Add Swagger UI endpoint
	http.HandleFunc("/swagger/", serveSwaggerUI)

	// Add Swagger JSON endpoint
	http.HandleFunc("/swagger/swagger.json", serveSwaggerJSON)

	log.Info("Routes configured",
		"apiPrefix", "/api/",
		"healthPrefix", "/health",
		"swaggerEndpoint", "/swagger/")

	// Start server
	port := fmt.Sprintf("%d", cfg.Server.Port)
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	// Show cool banner
	showBanner(cfg.Server.Host, port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error("❌ Server failed to start", "error", err.Error())
		os.Exit(1)
	}
}

// showBanner displays a cool ASCII banner when server starts
func showBanner(host, port string) {
	banner := `
-------------------------------------------------------------------
  ██╗   ██╗ ██████╗ ██╗   ██╗██████╗  █████╗ ██████╗ ██████╗ 
  ╚██╗ ██╔╝██╔═══██╗██║   ██║██╔══██╗██╔══██╗██╔══██╗██╔══██╗
   ╚████╔╝ ██║   ██║██║   ██║██████╔╝███████║██████╔╝██████╔╝
    ╚██╔╝  ██║   ██║██║   ██║██╔══██╗██╔══██║██╔═══╝ ██╔═══╝ 
     ██║   ╚██████╔╝╚██████╔╝██████╔╝██║  ██║██║     ██║     
     ╚═╝    ╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝     
-------------------------------------------------------------------
  🚀 Social Media Service

  📡 Server: http://%s:%s
  📚 Docs:   http://%s:%s/swagger/
  ❤️  Health: http://%s:%s/health
  🔍 Live:   http://%s:%s/health/live
  ✅ Ready:  http://%s:%s/health/ready
-------------------------------------------------------------------
`

	fmt.Printf(banner, host, port, host, port, host, port, host, port, host, port)
	fmt.Println()
}

// serveSwaggerUI serves the Swagger UI HTML page
func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	// Redirect to index.html if accessing /swagger/ without trailing slash
	if r.URL.Path == "/swagger" {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		return
	}

	// Serve the index.html file
	http.ServeFile(w, r, "docs/index.html")
}

// serveSwaggerJSON serves the Swagger JSON specification
func serveSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Serve the swagger JSON file
	http.ServeFile(w, r, "docs/swagger/docs.json")
}
