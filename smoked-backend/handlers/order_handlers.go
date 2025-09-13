package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smoked-meat/database"
	"smoked-meat/middleware"
	"smoked-meat/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type OrderRequest struct {
	Items []struct {
		ID            int     `json:"id" binding:"required"`
		Quantity      float64 `json:"quantity" binding:"required,gte=0"`
		SelectedSpice string  `json:"selected_spice" binding:"required"`
		Meat          string  `json:"meat" binding:"required"`
	} `json:"items" binding:"required"`
}

func CreateClientOrder(c *gin.Context) {
	var req OrderRequest
	redisClient := middleware.GetRedisClient()
	ctx := context.Background()
	phone_number, ok := c.Get("phone_number")
	if !ok {
		log.Error().Msg("Value of phone number does not exist")
	}

	log.Info().Any("phone_number", phone_number).Msg("Got phone number")

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Error to bind json")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data: " + err.Error()})
		return
	}

	// Логируем заказ, что в нём получилось
	log.Info().Any("request", req).Msg("Got request for order")

	// Рассчитываем общую сумму
	var totalPrice float64
	orderItems := make([]models.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		var product models.Assortment
		tx := database.DB.Begin()
		if err := tx.First(&product, item.ID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				log.Error().Err(err).Msg("Product not found")
				c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Product with ID %d not found", item.ID)})
				return
			}
			log.Error().Err(err).Msg("Error of database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}
		tx.Commit()

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

	var user models.User
	if err := database.DB.Where("phone_number = ?", phone_number).Find(&user).Error; err != nil {
		log.Error().Err(err).Msg("Error to find user by phone number")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	phoneNumber, ok := phone_number.(string)
	if !ok {
		log.Error().Msg("Failed to bring phone_number to string")
	}
	// Создаем заказ
	order := models.Order{
		CreatedAt:   time.Now(),
		Items:       orderItems,
		TotalPrice:  totalPrice,
		PhoneNumber: phoneNumber,
		Name:        user.Name,
	}

	// Логируем заказ перед сохранением
	log.Info().Any("order", order).Msg("Creating order")

	tx := database.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Database error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error to create order: " + err.Error()})
		return
	}
	tx.Commit()

	// Очищаем кэш
	pattern := []string{"/orders*", "/client/orders*" + phoneNumber, "assortment*"}
	for _, pattern := range pattern {
		cursor := uint64(0)
		for {
			keys, nextCursor, err := redisClient.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan orders cache keys")
				break
			}
			if len(keys) > 0 {
				if err := redisClient.Del(ctx, keys...).Err(); err != nil {
					log.Error().Err(err).Msg("Failed to delete orders cache keys")
				} else {
					log.Info().Str("pattern", pattern).Int("keys_deleted", len(keys)).Msg("Cache keys deleted")
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"order_id":     order.ID,
		"created_at":   order.CreatedAt,
		"items":        order.Items,
		"total_price":  order.TotalPrice,
		"phone_number": order.PhoneNumber,
		"name":         order.Name,
	})
}

func GetMyOrders(c *gin.Context) {

	phone_number, ok := c.Get("phone_number")
	if !ok {
		log.Error().Msg("Value of phone number does not exist")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	log.Info().Any("phone_number", phone_number).Msg("Got phone number")

	// Парсинг параметров пагинации
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Warn().Str("page", pageStr).Msg("Invalid page parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	sizeStr := c.DefaultQuery("size", "10")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 { // Добавьте лимит на размер, если нужно
		log.Warn().Str("size", sizeStr).Msg("Invalid size parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size"})
		return
	}

	redisClient := middleware.GetRedisClient()
	ctx := context.Background()

	// Ключ кэша включает URL с параметрами и phone_number
	cacheKey := c.Request.URL.String() + phone_number.(string)
	cached, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr != nil {
			log.Error().Err(jsonErr).Msg("Failed to unmarshal cached response")
		} else {
			log.Info().Msg("Cache hit for my orders")
			c.JSON(http.StatusOK, resp)
			return
		}
	} else if err != redis.Nil {
		log.Error().Err(err).Msg("Failed to hit cache")
	}

	// Подсчёт общего количества заказов для пользователя
	var totalCount int64
	if err := database.DB.Model(&models.Order{}).Where("phone_number = ?", phone_number).Count(&totalCount).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count orders for user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count orders"})
		return
	}

	// Выборка заказов с пагинацией
	var orders []models.Order
	if err := database.DB.Limit(size).Offset((page-1)*size).Where("phone_number = ?", phone_number).Find(&orders).Error; err != nil {
		log.Error().Err(err).Msg("Failed to find order by phone number")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	// Преобразование формата времени
	response := make([]models.OrderResponse, len(orders))
	for i, or := range orders {
		response[i] = models.ToOrdersResponse(or)
		if response[i].Name == "" {
			response[i].Name = "Name not specified"
		}
	}

	// Вычисление общего количества страниц
	totalPages := (int(totalCount) + size - 1) / size

	// Формирование ответа
	resp := gin.H{
		"orders":       response,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": page,
		"page_size":    size,
	}

	// Кэшируем весь ответ
	jsonResp, err := json.Marshal(resp)
	if err == nil {
		redisClient.Set(ctx, cacheKey, jsonResp, time.Hour)
	} else {
		log.Error().Err(err).Msg("Failed to serialize response for cache")
	}

	log.Info().
		Int("page", page).
		Int("size", size).
		Int64("total_count", totalCount).
		Msg("Fetched my orders list")

	c.JSON(http.StatusOK, resp)
}

func GetAllOrders(c *gin.Context) {

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Warn().Str("page", pageStr).Msg("Invalid page parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	sizeStr := c.DefaultQuery("size", "10")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 { // Добавьте лимит на размер, если нужно
		log.Warn().Str("size", sizeStr).Msg("Invalid size parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size"})
		return
	}

	redisClient := middleware.GetRedisClient()
	ctx := context.Background()

	// Ключ кэша включает URL с параметрами
	cacheKey := c.Request.URL.String()
	cached, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr != nil {
			log.Error().Err(jsonErr).Msg("Failed to unmarshal cached response")
		} else {
			log.Info().Msg("Cache hit for all orders")
			c.JSON(http.StatusOK, resp)
			return
		}
	} else if err != redis.Nil {
		log.Error().Err(err).Msg("Failed to hit cache")
	}

	// Подсчёт общего количества заказов
	var totalCount int64
	if err := database.DB.Model(&models.Order{}).Count(&totalCount).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count orders"})
		return
	}

	// Выборка заказов с пагинацией
	var orders []models.Order
	if err := database.DB.Limit(size).Offset((page - 1) * size).Find(&orders).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get orders list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find orders"})
		return
	}

	// Преобразование формата времени
	response := make([]models.OrderResponse, len(orders))
	for i, or := range orders {
		response[i] = models.ToOrdersResponse(or)
		if response[i].Name == "" {
			response[i].Name = "Name not specified"
		}
	}

	// Вычисление общего количества страниц
	totalPages := (int(totalCount) + size - 1) / size

	// Формирование ответа
	resp := gin.H{
		"orders":       response,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": page,
		"page_size":    size,
	}

	// Кэшируем весь ответ
	jsonResp, err := json.Marshal(resp)
	if err == nil {
		redisClient.Set(ctx, cacheKey, jsonResp, time.Hour)
	} else {
		log.Error().Err(err).Msg("Failed to serialize response for cache")
	}

	// Логирование успешного запроса
	log.Info().
		Int("page", page).
		Int("size", size).
		Int64("total_count", totalCount).
		Msg("Fetched orders list")

	c.JSON(http.StatusOK, resp)
}

func GetOrderByID(ctx *gin.Context) {
	var order models.Order

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize id")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invalid id of product"})
		return
	}

	if err := database.DB.First(&order, id).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get order by ID")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error to find orders",
		})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

func DeleteOrderByID(ctx *gin.Context) {
	var order models.Order

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize ID")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invalid id of product"})
		return
	}

	if err := database.DB.First(&order, id).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get order by ID")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error to find orders",
		})
		return
	}

	tx := database.DB.Begin()
	if err := tx.Delete(&order, id).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to delete order by ID")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ID not found"})
		return
	}
	tx.Commit()

	ctx.Status(http.StatusNoContent)
}
