package main

import (
	"log"
	"os"
	"smoked-meat/database"
	"smoked-meat/handlers"
	"smoked-meat/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	if err := database.NewPostgresDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	redisClient := middleware.NewRedisClient()
	if redisClient == nil {
		log.Fatalf("Error to start redis")
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
	router.POST("/order", handlers.CreateOrder(database.DB, redisClient))

	// Защищенные роуты с JWT
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware())

	api.POST("/product", handlers.AddProduct)
	api.PUT("/product/:id", handlers.UpdateProduct)
	api.DELETE("/product/:id", handlers.DeleteProduct)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
