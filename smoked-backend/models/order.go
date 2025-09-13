package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	ProductID     int     `json:"product_id" gorm:"not null"`
	Quantity      float64 `json:"quantity" gorm:"type:decimal(10,2);not null"`
	SelectedSpice string  `json:"selected_spice" gorm:"type:text;not null"`
	Meat          string  `json:"meat" gorm:"type:text;not null"`
}

type Order struct {
	gorm.Model
	ID          int        `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time  `json:"created_at"`
	Items       OrderItems `json:"items" gorm:"type:jsonb;not null"`
	TotalPrice  float64    `json:"total_price" gorm:"type:decimal(10,2);not null"`
	PhoneNumber string     `json:"phone_number" gorm:"type:varchar(15);not null"`
	Name        string     `json:"name" gorm:"type:text"`
}

type OrderItems []OrderItem

func (oi *OrderItems) Scan(value interface{}) error {
	bytes, ok := value.([]byte) // Проверяем, что данные формата jsonb из поля Items имеют тип []byte
	if !ok {
		return errors.New("Scan: unable to scan value into OrderItem")
	}
	return json.Unmarshal(bytes, oi) // Десериализуем данные из JSON в []OrderItems, записываем в oi
}

func (oi OrderItems) Value() (driver.Value, error) {
	return json.Marshal(oi) // Сериализуем данные OrderItems для сохранения в базу данных
}

type OrderResponse struct {
	ID          int        `json:"id"`
	CreatedAt   string     `json:"created_at"`
	Items       OrderItems `json:"items"`
	TotalPrice  float64    `json:"total_price"`
	PhoneNumber string     `json:"phone_number"`
	Name        string     `json:"name"`
}

func ToOrdersResponse(order Order) OrderResponse {
	return OrderResponse{
		ID:          order.ID,
		CreatedAt:   order.CreatedAt.Format("2006-01-02-03:04"),
		Items:       order.Items,
		TotalPrice:  order.TotalPrice,
		PhoneNumber: order.PhoneNumber,
		Name:        order.Name,
	}
}
