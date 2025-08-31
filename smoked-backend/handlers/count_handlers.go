package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smoked-meat/database"
	"smoked-meat/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type BulkCalculateRequest struct {
	Items []struct {
		ID            int     `json:"id" binding:"required"`
		Quantity      float64 `json:"quantity" binding:"required,gte=0"`
		SelectedSpice string  `json:"selectedSpice" binding:"required"`
	} `json:"items" binding:"required"`
}

func CalculateBulk(db *gorm.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkCalculateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные входные данные: " + err.Error()})
			return
		}

		// Формируем ключ кэша на основе ID, количества и специи
		cacheKey := "bulk_price:"
		for _, item := range req.Items {
			cacheKey += fmt.Sprintf("%d:%.2f:%s;", item.ID, item.Quantity, item.SelectedSpice)
		}
		const cacheTTL = 5 * time.Minute

		// Проверяем кэш
		cached, err := redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}

		// Запрашиваем продукты из базы
		var totalPrice float64
		var totalQuantity float64
		resultItems := make([]map[string]interface{}, 0, len(req.Items))
		for _, item := range req.Items {
			var product models.Assortment
			if err := db.First(&product, item.ID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Товар с ID %d не найден", item.ID)})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных: " + err.Error()})
				return
			}
			itemTotal := product.Price * item.Quantity

			totalQuantity += item.Quantity
			totalPrice += itemTotal

			resultItems = append(resultItems, map[string]interface{}{
				"id":            product.ID,
				"meat":          product.Meat,
				"quantity":      item.Quantity,
				"total_price":   itemTotal,
				"selectedSpice": item.SelectedSpice,
				"spice":         product.Spice,
			})
		}

		if totalQuantity > 19 {
			totalPrice *= 0.88
		}

		if totalQuantity > 9 && totalQuantity < 20 {
			totalPrice *= 0.92
		}

		result := gin.H{
			"items":       resultItems,
			"total_price": totalPrice,
		}

		// Сохраняем в кэш
		data, err := json.Marshal(result)
		if err == nil {
			redisClient.Set(context.Background(), cacheKey, data, cacheTTL)
		}

		c.JSON(http.StatusOK, result)
	}
}

func CalculatePrice(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.CalculatePriceRequest
		var product models.Assortment

		if err := ctx.ShouldBindJSON(&req); err != nil {
			log.Printf("Invalid input request")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data: " + err.Error()})
			return
		}

		if err := database.DB.First(&product, req.ID).Error; err != nil {
			log.Printf("Product not found:%v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Failed to found product": err,
			})
			return
		}
		totalPrice := product.Price * req.Quantity

		ctx.JSON(http.StatusOK, gin.H{
			"id":          product.ID,
			"meat":        product.Meat,
			"quantity":    req.Quantity,
			"total_price": totalPrice,
		})
	}
}
