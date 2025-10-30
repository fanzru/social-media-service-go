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
	commentApp "github.com/fanzru/social-media-service-go/internal/app/comment/app"
	commentHTTP "github.com/fanzru/social-media-service-go/internal/app/comment/port"
	commentGenHTTP "github.com/fanzru/social-media-service-go/internal/app/comment/port/genhttp"
	commentRepo "github.com/fanzru/social-media-service-go/internal/app/comment/repo"
	healthApp "github.com/fanzru/social-media-service-go/internal/app/health/app"
	healthHTTP "github.com/fanzru/social-media-service-go/internal/app/health/port"
	healthGenHTTP "github.com/fanzru/social-media-service-go/internal/app/health/port/genhttp"
	healthRepo "github.com/fanzru/social-media-service-go/internal/app/health/repo"
	postApp "github.com/fanzru/social-media-service-go/internal/app/post/app"
	postHTTP "github.com/fanzru/social-media-service-go/internal/app/post/port"
	postGenHTTP "github.com/fanzru/social-media-service-go/internal/app/post/port/genhttp"
	postRepo "github.com/fanzru/social-media-service-go/internal/app/post/repo"
	"github.com/fanzru/social-media-service-go/pkg/influxdb"
	"github.com/fanzru/social-media-service-go/pkg/jwt"
	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/middleware"
	"github.com/fanzru/social-media-service-go/pkg/reqctx"
	"github.com/fanzru/social-media-service-go/pkg/sqlwrap"
	"github.com/fanzru/social-media-service-go/pkg/storage"
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

	// Initialize InfluxDB client
	influxHost := os.Getenv("INFLUXDB_HOST")
	if influxHost == "" {
		influxHost = "http://localhost:8086" // Default for local development
	}
	influxClient, err := influxdb.NewClient(influxHost, "my-super-secret-auth-token", "social-media", "metrics")
	if err != nil {
		log.Error("Failed to initialize InfluxDB client", "error", err.Error())
		os.Exit(1)
	}
	defer influxClient.Close()
	log.Info("InfluxDB client initialized")

	// Wrap database with metrics and logging
	var dbInterface interface{} = db
	if cfg.Database.LogQueries {
		dbInterface = sqlwrap.NewDBWithInfluxDB(db, influxClient)
		log.Info("Database query logging enabled", "slowQueryThreshold", cfg.Database.SlowQueryThreshold)
	}

	// Initialize JWT service
	jwtService := jwt.NewService(cfg.JWT.Secret, time.Duration(cfg.JWT.Expiration)*time.Hour)
	log.Info("JWT service initialized")

	// Initialize account repository and service
	accountRepository := repo.NewRepository(dbInterface)
	log.Info("Account repository initialized")

	// Initialize image storage service
	imageStorage := storage.NewImageStorageService(&cfg.Storage)
	log.Info("Image storage service initialized")

	accountService := accountApp.NewService(accountRepository, jwtService, imageStorage)
	log.Info("Account service initialized")

	accountHandler := accountHTTP.NewHandler(accountService)
	log.Info("Account HTTP handler initialized")

	// Initialize post repository and service
	postRepository := postRepo.NewRepository(dbInterface)
	log.Info("Post repository initialized")

	// Initialize comment repository
	commentRepository := commentRepo.NewRepository(dbInterface)
	log.Info("Comment repository initialized")

	postService := postApp.NewService(postRepository, commentRepository, imageStorage)
	log.Info("Post service initialized")

	postHandler := postHTTP.NewHandler(postService)
	log.Info("Post HTTP handler initialized")

	// Initialize comment service
	commentService := commentApp.NewService(commentRepository, postRepository)
	log.Info("Comment service initialized")

	commentHandler := commentHTTP.NewHandler(commentService)
	log.Info("Comment HTTP handler initialized")

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

	// Initialize metrics middleware
	metricsMiddleware := middleware.InfluxDBMiddleware(influxClient)
	log.Info("Metrics middleware initialized")

	// Add security requirements manually for now
	authMiddleware.AddSecurityRequirement("GET", "/api/account/profile", true)
	authMiddleware.AddSecurityRequirement("DELETE", "/api/account", true)
	authMiddleware.AddSecurityRequirement("GET", "/api/posts", false)
	authMiddleware.AddSecurityRequirement("POST", "/api/posts", true)
	authMiddleware.AddSecurityRequirement("PUT", "/api/posts", true)
	authMiddleware.AddSecurityRequirement("DELETE", "/api/posts", true)
	// New explicit paths
	authMiddleware.AddSecurityRequirement("GET", "/api/posts/by-user", false)
	authMiddleware.AddSecurityRequirement("GET", "/api/comments/by-post", false)
	authMiddleware.AddSecurityRequirement("POST", "/api/comments/by-post", true)
	authMiddleware.AddSecurityRequirement("PUT", "/api/comments", true)
	authMiddleware.AddSecurityRequirement("DELETE", "/api/comments", true)
	log.Info("Security requirements loaded manually")

	// Create combined API handler
	apiHandler := http.NewServeMux()

	// Register per-domain handlers using a single mux (generated handlers define their own patterns)
	genhttp.HandlerFromMux(accountHandler, apiHandler)
	postGenHTTP.HandlerFromMux(postHandler, apiHandler)
	commentGenHTTP.HandlerFromMux(commentHandler, apiHandler)

	// Setup routes using combined API handler with comprehensive middleware
	var apiHandlerWithMiddleware http.Handler = apiHandler

	// Apply middleware in order: metrics -> auth -> logging -> request context
	apiHandlerWithMiddleware = metricsMiddleware(apiHandlerWithMiddleware)
	apiHandlerWithMiddleware = authMiddleware.Middleware()(apiHandlerWithMiddleware)
	apiHandlerWithMiddleware = loggingMiddleware(apiHandlerWithMiddleware)
	apiHandlerWithMiddleware = reqctx.Middleware(apiHandlerWithMiddleware)

	// InfluxDB metrics are sent directly via HTTP, no endpoint needed
	log.Info("InfluxDB metrics enabled")

	// Create health OpenAPI server with middleware
	healthApiHandler := healthGenHTTP.Handler(healthHandler)

	// Create main mux for all routes
	mainMux := http.NewServeMux()

	// Add API routes with middleware
	mainMux.Handle("/api/", apiHandlerWithMiddleware)

	// Add health check endpoints with logging middleware only (no auth required)
	mainMux.Handle("/health",
		reqctx.Middleware(
			loggingMiddleware(healthApiHandler),
		),
	)
	mainMux.Handle("/health/",
		reqctx.Middleware(
			loggingMiddleware(healthApiHandler),
		),
	)

	// Add Swagger UI endpoint
	mainMux.HandleFunc("/swagger/", serveSwaggerUI)

	// Add Swagger JSON endpoint
	mainMux.HandleFunc("/swagger/swagger.json", serveSwaggerJSON)

	// Add favicon endpoint
	mainMux.HandleFunc("/favicon.ico", serveFavicon)

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

	if err := http.ListenAndServe(":"+port, mainMux); err != nil {
		log.Error("❌ Server failed to start", "error", err.Error())
		os.Exit(1)
	}
}

// showBanner displays a cool ASCII banner when server starts
func showBanner(host, port string) {
	banner := `
--------------------------------------------------------------------------
'     ___  ___  ___          _________  _______   ________  _____ ______      
'    |\  \|\  \|\  \        |\___   ___\\  ___ \ |\   __  \|\   _ \  _   \    
'    \ \  \\\  \ \  \       \|___ \  \_\ \   __/|\ \  \|\  \ \  \\\__\ \  \   
'     \ \   __  \ \  \           \ \  \ \ \  \_|/_\ \   __  \ \  \\|__| \  \  
'      \ \  \ \  \ \  \           \ \  \ \ \  \_|\ \ \  \ \  \ \  \    \ \  \ 
'       \ \__\ \__\ \__\           \ \__\ \ \_______\ \__\ \__\ \__\    \ \__\
'        \|__|\|__|\|__|            \|__|  \|_______|\|__|\|__|\|__|     \|__|
'                                                                             
'  
'  🚀 Your Backend Service Already Running....🚀
'
'  📡 Server: http://%s:%s | 📚 Docs: http://%s:%s/swagger/
---------------------------------------------------------------------------
'  created with ❤️ by: fanzru.dev
---------------------------------------------------------------------------
`

	fmt.Printf(banner, host, port, host, port)
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

// serveFavicon serves the favicon.ico file
func serveFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	http.ServeFile(w, r, "docs/favicon.ico")
}
