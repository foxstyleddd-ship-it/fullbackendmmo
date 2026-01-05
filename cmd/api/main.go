package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/hp-mmo/backend/cmd/api/handlers"
	"github.com/hp-mmo/backend/internal/audit"
	"github.com/hp-mmo/backend/internal/auth"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/inventory"
	"github.com/hp-mmo/backend/internal/pkg/config"
	"github.com/hp-mmo/backend/internal/pkg/db"
	"github.com/hp-mmo/backend/internal/pkg/logger"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
	pkgredis "github.com/hp-mmo/backend/internal/pkg/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Str("env", cfg.Env).Msg("Starting HP MMO API Server")

	// Connect to PostgreSQL
	database, err := db.Connect(
		cfg.GetDBConnectionString(),
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close(database)

	// Connect to Redis
	redisClient, err := pkgredis.Connect(cfg.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	authRepo := auth.NewRepository(database)
	charRepo := character.NewRepository(database)
	auditRepo := audit.NewRepository(database)
	inventoryRepo := inventory.NewRepository(database)

	// Initialize services
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	authService := auth.NewService(authRepo, jwtService, redisClient, cfg.Session.Duration)
	charService := character.NewService(charRepo)
	auditService := audit.NewService(auditRepo)
	inventoryService := inventory.NewService(inventoryRepo, charRepo, database)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, charRepo)
	charHandler := handlers.NewCharacterHandler(charService, jwtService)
	adminHandler := handlers.NewAdminHandler(charService, charRepo, auditService, inventoryService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)

	// Setup Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.CORSMiddleware([]string{"*"})) // Configure properly in production

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Auth routes (public)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", middleware.AuthMiddleware(jwtService), authHandler.Logout)
		}

		// Character routes (authenticated)
		charGroup := v1.Group("/characters")
		charGroup.Use(middleware.AuthMiddleware(jwtService))
		{
			charGroup.GET("", charHandler.ListCharacters)
			charGroup.POST("", charHandler.CreateCharacter)
			charGroup.GET("/:id", charHandler.GetCharacter)
			charGroup.POST("/:id/select", charHandler.SelectCharacter)
			charGroup.DELETE("/:id", charHandler.DeleteCharacter)

			// Inventory routes (nested under character)
			charGroup.GET("/:id/inventory", inventoryHandler.GetInventory)
			charGroup.GET("/:id/equipment", inventoryHandler.GetEquipment)
			charGroup.POST("/:id/equip", inventoryHandler.EquipItem)
			charGroup.POST("/:id/unequip", inventoryHandler.UnequipItem)
			charGroup.POST("/:id/purchase", inventoryHandler.PurchaseItem)
			charGroup.POST("/:id/sell", inventoryHandler.SellItem)
		}

		// Admin routes (authenticated + role check)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(jwtService))
		adminGroup.Use(middleware.RequireRole("gm", "admin", "superadmin"))
		{
			adminGroup.GET("/health", func(c *gin.Context) {
				middleware.SuccessResponse(c, http.StatusOK, gin.H{
					"message": "Admin API ready",
				})
			})

			// GM Commands
			adminGroup.POST("/commands/execute", adminHandler.ExecuteCommand)

			// Audit Logs
			adminGroup.GET("/audit/character/:id", adminHandler.GetCharacterAuditLogs)
			adminGroup.GET("/audit/my-actions", adminHandler.GetAdminAuditLogs)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.APIPort),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Server.APIPort).Msg("API server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited successfully")
}
