package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"smoked-meat/database"
	"smoked-meat/handlers"
	"smoked-meat/middleware"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {

	if err := database.NewPostgresDB(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")
	}

	redisClient, err := middleware.NewRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to init Redis")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	router.GET("/assortment", handlers.GetAssortment)
	router.GET("/product/:id", handlers.GetProduct)
	router.POST("/calculate-price", handlers.CalculatePrice(database.DB))
	router.POST("/calculate-bulk", handlers.CalculateBulk(database.DB, redisClient))

	// Защищённые роуты с JWT
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware())

	api.POST("/product", handlers.AddProduct)
	api.PUT("/product/:id", handlers.UpdateProduct)
	api.DELETE("/product/:id", handlers.DeleteProduct)
	api.GET("/orders", handlers.GetAllOrders)
	api.GET("/orders/:id", handlers.GetOrderByID)
	api.DELETE("/orders/:id", handlers.DeleteOrderByID)
	api.POST("/admin/register", middleware.OwnerOnly(), handlers.RegisterByOwner)
	api.GET("/client/orders", handlers.GetMyOrders)
	api.POST("/order", handlers.CreateClientOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}
}
