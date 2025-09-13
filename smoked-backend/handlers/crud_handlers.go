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
)

func GetAssortment(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Warn().Str("page", pageStr).Msg("Invalid page parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	sizeStr := ctx.DefaultQuery("size", "10")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		log.Warn().Str("size", sizeStr).Msg("Invalid size parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid size"})
		return
	}

	var totalCount int64
	if err := database.DB.Model(&models.Assortment{}).Count(&totalCount).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count assortment")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count assortment"})
		return
	}

	var ass *[]models.Assortment
	redisClient := middleware.GetRedisClient()
	cacheKey := ctx.Request.URL.String()
	c := context.Background()

	if redisClient != nil {
		cached, err := redisClient.Get(c, cacheKey).Result()
		if err == nil {
			if json.Unmarshal([]byte(cached), &ass) == nil {
				log.Printf("Cache hit for assortment list")
				ctx.JSON(http.StatusOK, ass)
				return
			}
			log.Error().Err(err).Msg("Failed to unmarshal cached assortment list")
		} else if err != redis.Nil {
			log.Error().Err(err).Msg("Redis error for assortment list")
		}
	} else {
		log.Info().Str("cacheKey", cacheKey).Msg("Redis client is nil, skipping cache for assortment list")
	}

	if res := database.DB.Limit(size).Offset((page - 1) * size).Find(&ass); res.Error != nil {
		log.Error().Err(res.Error).Msg("Failed to find assortment")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get assortment"})
		return
	}

	totalPages := (int(totalCount) + size - 1) / size

	resp := gin.H{
		"assortment":   ass,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": page,
		"page_size":    size,
	}

	assortmentJSON, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize assortment")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize assortment"})
		return
	}

	if redisClient != nil {
		err = redisClient.Set(ctx, cacheKey, assortmentJSON, 5*time.Minute).Err()
		if err != nil {
			log.Error().Err(err).Msg("Failed to cache assortment")
		}
	} else {
		log.Info().Str("cacheKey", cacheKey).Msg("Redis client is nil, skipping cache for assortment list")
	}

	ctx.JSON(http.StatusOK, resp)
}

func GetProduct(ctx *gin.Context) {
	var product models.Assortment
	redisClient := middleware.GetRedisClient()
	cacheKey := ctx.Request.URL.String()
	c := context.Background()

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get id")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invalid id of product"})
		return
	}

	if redisClient != nil {
		cached, err := redisClient.Get(c, cacheKey).Result()
		if err == nil {
			if json.Unmarshal([]byte(cached), &product) == nil {
				log.Info().Str("cacheKey", cacheKey).Msg("Cache hit for product id")
				ctx.JSON(http.StatusOK, product)
				return
			}
			log.Error().Err(err).Msg("Failed to unmarshal cached product")
		} else if err != redis.Nil {
			log.Error().Err(err).Msg("Redis error for product id")
		}
	} else {
		log.Info().Str("cacheKey", cacheKey).Msg("Redis client is nil, skipping cache for product id")
	}

	if res := database.DB.First(&product, id); res == nil {
		log.Error().Err(err).Msg("Failed to find product id")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	productJSON, err := json.Marshal(product)
	if err != nil {
		log.Error().Err(err).Msg("Error to serialize product")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize product"})
		return
	}

	if redisClient != nil {
		err = redisClient.Set(ctx, cacheKey, productJSON, 5*time.Minute).Err()
		if err != nil {
			log.Error().Err(err).Msg("Failed to cache product")
		}
	} else {
		log.Info().Str("cacheKey", cacheKey).Msg("Redis client is nil, skipping cache for product")
	}

	ctx.JSON(http.StatusOK, &product)
}

func AddProduct(ctx *gin.Context) {
	var req models.Assortment
	c := context.Background()
	redisClient := middleware.GetRedisClient()

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind json")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid product input": err.Error()})
		return
	}

	product := models.Assortment{
		Meat:         req.Meat,
		Availability: req.Availability,
		Price:        req.Price,
		Spice:        req.Spice,
	}

	tx := database.DB.Begin()
	if res := tx.Create(&product); res.Error != nil {
		tx.Rollback()
		log.Error().Err(res.Error).Msg("Error to create product")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Failed to create product": res.Error})
		return
	}
	tx.Commit()

	if redisClient != nil {
		pattern := []string{"/assortment*"}
		for _, pattern := range pattern {
			cursor := uint64(0)
			for {
				keys, nextCursor, err := redisClient.Scan(c, cursor, pattern, 100).Result()
				if err != nil {
					log.Error().Err(err).Msg("Failed to scan assortment cache keys")
					break
				}
				if len(keys) > 0 {
					if err := redisClient.Del(ctx, keys...).Err(); err != nil {
						log.Error().Err(err).Msg("Failed to delete assortment cache keys")
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
	}

	ctx.JSON(http.StatusCreated, product)
}

func UpdateProduct(ctx *gin.Context) {
	var product models.Assortment
	var updatedProduct models.Assortment
	redisClient := middleware.GetRedisClient()
	c := context.Background()

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid input id")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid id": err,
		})
	}

	if err := ctx.ShouldBindJSON(&updatedProduct); err != nil {
		log.Error().Err(err).Msg("Failed to bind json")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid of input data": err.Error()})
		return
	}

	res := database.DB.First(&product, id)
	if res == nil {
		tx := database.DB.Begin()
		err := tx.Create(&updatedProduct).Error
		if err != nil {
			tx.Rollback()
			log.Error().Err(err).Msg("Failed to create product")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error to create updated product",
			})
			return
		}
		tx.Commit()
		ctx.JSON(http.StatusCreated, updatedProduct)
		return
	}

	product.Meat = updatedProduct.Meat
	product.Availability = updatedProduct.Availability
	product.Price = updatedProduct.Price
	product.Spice = updatedProduct.Spice

	tx := database.DB.Begin()
	if err := tx.Save(product).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to save product")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error to save product": err,
		})
	}
	tx.Commit()

	if redisClient != nil {
		pattern := []string{"/assortment*", fmt.Sprintf("/product/%v", id)}
		for _, pattern := range pattern {
			cursor := uint64(0)
			for {
				keys, nextCursor, err := redisClient.Scan(c, cursor, pattern, 100).Result()
				if err != nil {
					log.Error().Err(err).Msg("Failed to scan assortment cache keys")
					break
				}
				if len(keys) > 0 {
					if err := redisClient.Del(ctx, keys...).Err(); err != nil {
						log.Error().Err(err).Msg("Failed to delete assortment cache keys")
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
	}

	ctx.JSON(http.StatusOK, product)
}

func DeleteProduct(ctx *gin.Context) {
	var product models.Assortment
	redisClient := middleware.GetRedisClient()
	c := context.Background()

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Invalid input id")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid id": err,
		})
	}

	if err := database.DB.First(&product, id).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get product by ID")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error to find products",
		})
		return
	}

	tx := database.DB.Begin()
	if err := tx.Delete(&product, id).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Invalid id")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Failed to delete id": err,
		})
	}
	tx.Commit()

	if redisClient != nil {
		pattern := []string{"/assortment*", fmt.Sprintf("/product/%v", id)}
		for _, pattern := range pattern {
			cursor := uint64(0)
			for {
				keys, nextCursor, err := redisClient.Scan(c, cursor, pattern, 100).Result()
				if err != nil {
					log.Error().Err(err).Msg("Failed to scan assortment cache keys")
					break
				}
				if len(keys) > 0 {
					if err := redisClient.Del(ctx, keys...).Err(); err != nil {
						log.Error().Err(err).Msg("Failed to delete assortment cache keys")
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
	}

	ctx.Status(http.StatusNoContent)
}
