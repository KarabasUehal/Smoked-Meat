package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"smoked-meat/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET"))

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	log.Printf("Connecting to Redis at %s", addr)
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен авторизации отсутствует"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CacheMiddleware(c *gin.Context) {

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	if c.Request.Method != http.MethodGet {
		c.Next()
		return
	}

	cacheKey := c.Request.URL.String()
	ctx := context.Background()

	// Проверяем, инициализирован ли redisClient
	if redisClient == nil {
		log.Printf("Redis client is nil, skipping cache for key: %s", cacheKey)
		c.Next()
		return
	}

	// Проверяем кеш в Redis
	cached, err := redisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		log.Printf("Cache miss for key: %s", cacheKey)
		c.Next()
		return
	}
	if err != nil {
		log.Printf("Redis error for key %s: %v", cacheKey, err)
		c.Next()
		return
	}

	// Десериализация в зависимости от маршрута
	if cacheKey == "/api/assortment" {
		var assortment []models.Assortment
		if err := json.Unmarshal([]byte(cached), &assortment); err != nil {
			log.Printf("Failed to unmarshal cached assortment list for key %s: %v", cacheKey, err)
			c.Next()
			return
		}
		log.Printf("Cache hit for assortment list")
		c.JSON(http.StatusOK, assortment)
	} else {
		var product models.Assortment
		if err := json.Unmarshal([]byte(cached), &product); err != nil {
			log.Printf("Failed to unmarshal cached product for key %s: %v", cacheKey, err)
			c.Next()
			return
		}
		log.Printf("Cache hit for key: %s", cacheKey)
		c.JSON(http.StatusOK, product)
	}
	c.Abort()
}
