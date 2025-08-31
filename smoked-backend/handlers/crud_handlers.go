package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"smoked-meat/database"
	"smoked-meat/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client
var assortment []models.Assortment

func GetAssortment(ctx *gin.Context) {
	var ass *[]models.Assortment
	cacheKey := ctx.Request.URL.String()
	c := context.Background()

	if redisClient != nil {
		cached, err := redisClient.Get(c, cacheKey).Result()
		if err == nil {
			var assortment []models.Assortment
			if json.Unmarshal([]byte(cached), &assortment) == nil {
				log.Printf("Cache hit for assortment list")
				ctx.JSON(http.StatusOK, assortment)
				return
			}
			log.Printf("Failed to unmarshal cached assortment list: %v", err)
		} else if err != redis.Nil {
			log.Printf("Redis error for assortment list: %v", err)
		}
	} else {
		log.Printf("Redis client is nil, skipping cache for assortment list")
	}

	if res := database.DB.Find(&ass); res.Error != nil {
		log.Printf("Failed to find assortment")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get assortment"})
		return
	}

	assortmentJSON, err := json.Marshal(assortment)
	if err != nil {
		log.Printf("Failed to serialize assortment: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize assortment"})
		return
	}

	if redisClient != nil {
		err = redisClient.Set(ctx, cacheKey, assortmentJSON, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache assortment: %v", err)
		}
	} else {
		log.Printf("Redis client is nil, skipping cache for assortment list")
	}

	ctx.JSON(http.StatusOK, ass)
}

func GetProduct(ctx *gin.Context) {
	var product models.Assortment

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Failed to initialize id:%e", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invalid id of product"})
		return
	}

	if res := database.DB.First(&product, id); res == nil {
		log.Printf("Failed to find product:%v", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	ctx.JSON(http.StatusOK, &product)
}

func AddProduct(ctx *gin.Context) {
	var req models.Assortment

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error to bind JSON:%v", err)
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

	if res := database.DB.Create(&product); res.Error != nil {
		log.Printf("Error to create product:%v", res.Error)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Failed to create product": res.Error})
		return
	}

	ctx.JSON(http.StatusCreated, product)
}

func UpdateProduct(ctx *gin.Context) {
	var product models.Assortment
	var updatedProduct models.Assortment

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Invalid input id:%e", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid id": err,
		})
	}

	if err := ctx.ShouldBindJSON(&updatedProduct); err != nil {
		log.Printf("Failed to bind JSON for item ID %d: %v", id, err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid of input data": err.Error()})
		return
	}

	res := database.DB.First(&product, id)
	if res == nil {
		database.DB.Create(&updatedProduct)
		ctx.JSON(http.StatusCreated, updatedProduct)
		return
	}

	product.Meat = updatedProduct.Meat
	product.Availability = updatedProduct.Availability
	product.Price = updatedProduct.Price
	product.Spice = updatedProduct.Spice

	if err := database.DB.Save(product); err != nil {
		log.Printf("Failed to save product: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error to save product": err,
		})
	}

	ctx.JSON(http.StatusOK, product)
}

func DeleteProduct(ctx *gin.Context) {
	var product models.Assortment

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Invalid input id:%e", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Invalid id": err,
		})
	}

	if err := database.DB.Delete(&product, id); err != nil {
		log.Printf("Invalid id:%v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Failed to delete id": err,
		})
	}

	ctx.Status(http.StatusNoContent)
}
