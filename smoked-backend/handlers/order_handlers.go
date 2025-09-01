package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"smoked-meat/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type OrderRequest struct {
	Items []struct {
		ID            int     `json:"id" binding:"required"`
		Quantity      float64 `json:"quantity" binding:"required,gte=0"`
		SelectedSpice string  `json:"selected_spice" binding:"required"`
		Meat          string  `json:"meat" binding:"required"`
	} `json:"items" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required,len=12"`
}

func CreateOrder(db *gorm.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка привязки JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные входные данные: " + err.Error()})
			return
		}

		// Логируем полученные данные
		log.Printf("Получен запрос на заказ: %+v", req)

		// Рассчитываем общую сумму
		var totalPrice float64
		orderItems := make([]models.OrderItem, 0, len(req.Items))
		for _, item := range req.Items {
			var product models.Assortment
			if err := db.First(&product, item.ID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Товар с ID %d не найден", item.ID)})
					return
				}
				log.Printf("Ошибка базы данных для ID %d: %v", item.ID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных: " + err.Error()})
				return
			}

			totalPrice += product.Price * item.Quantity
			orderItems = append(orderItems, models.OrderItem{
				ProductID:     item.ID,
				Quantity:      item.Quantity,
				SelectedSpice: item.SelectedSpice,
				Meat:          item.Meat,
			})
		}

		// Применяем скидки
		totalQuantity := 0.0
		for _, item := range req.Items {
			totalQuantity += item.Quantity
		}
		if totalQuantity > 19 {
			totalPrice *= 0.88
		}

		if totalQuantity > 9 && totalQuantity < 20 {
			totalPrice *= 0.92
		}

		// Создаем заказ
		order := models.Order{
			CreatedAt:   time.Now(),
			Items:       orderItems,
			TotalPrice:  totalPrice,
			PhoneNumber: req.PhoneNumber,
		}

		// Логируем заказ перед сохранением
		log.Printf("Создается заказ: %+v", order)

		if err := db.Create(&order).Error; err != nil {
			log.Printf("Ошибка сохранения заказа: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании заказа: " + err.Error()})
			return
		}

		// Очищаем кэш
		redisClient.Del(context.Background(), "assortment")

		c.JSON(http.StatusOK, gin.H{
			"order_id":     order.ID,
			"created_at":   order.CreatedAt,
			"items":        order.Items,
			"total_price":  order.TotalPrice,
			"phone_number": order.PhoneNumber,
		})
	}
}
