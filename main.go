package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strpe-connect/handlers"
	"strpe-connect/repository"
	"strpe-connect/services"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "rival")
	stripeSecretKey := getEnv("STRIPE_SECRET_KEY", "")
	stripeWebhookSecret := getEnv("STRIPE_WEBHOOK_SECRET", "")
	port := getEnv("PORT", "8080")

	if stripeSecretKey == "" {
		log.Fatal("STRIPE_SECRET_KEY is required")
	}

	// Initialize database connection
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Initialize repository
	repo := repository.NewStripeConnectRepository(dbPool)

	// Initialize service
	stripeService := services.NewStripeConnectService(repo, stripeSecretKey)

	// Initialize handler
	handler := handlers.NewStripeConnectHandler(stripeService, stripeWebhookSecret)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Organization-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "stripe-connect-marketplace",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Stripe Connect routes
		connect := api.Group("/connect")
		{
			// Onboarding
			connect.POST("/onboard", handler.CreateConnectAccount)
			connect.GET("/status", handler.GetConnectAccountStatus)
			connect.POST("/refresh-onboarding", handler.RefreshOnboardingLink)
			connect.GET("/connected-developers", handler.GetConnectedDevelopersForOrg)

			// Wallet
			wallet := connect.Group("/wallet")
			{
				wallet.GET("/balance", handler.GetWalletBalance)
				wallet.GET("/transactions", handler.GetTransactionHistory)
			}

			// Withdrawals
			withdrawals := connect.Group("/withdrawals")
			{
				withdrawals.POST("/request", handler.RequestWithdrawal)
				withdrawals.GET("/history", handler.GetWithdrawalHistory)
			}

			// Payments
			payments := connect.Group("/payments")
			{
				payments.POST("/execute", handler.ProcessFunctionPayment)
			}
		}

		// Admin endpoints
		admin := api.Group("/admin")
		{
			admin.GET("/connected-developers", handler.GetConnectedDevelopers)
		}

		// Webhooks
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/stripe-connect", handler.HandleWebhook)
		}
	}

	// Serve static files (React build)
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		log.Printf("📖 API Documentation: http://localhost:%s/swagger/index.html", port)
		log.Printf("🏥 Health Check: http://localhost:%s/health", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
