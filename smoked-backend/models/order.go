package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type OrderItem struct {
	ProductID     int     `json:"product_id" gorm:"not null"`
	Quantity      float64 `json:"quantity" gorm:"type:decimal(10,2);not null"`
	SelectedSpice string  `json:"selected_spice" gorm:"type:text;not null"`
	Meat          string  `json:"meat" gorm:"type:text;not null"`
}

type Order struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time  `json:"created_at"`
	Items       OrderItems `json:"items" gorm:"type:jsonb;not null"`
	TotalPrice  float64    `json:"total_price" gorm:"type:decimal(10,2);not null"`
	PhoneNumber string     `json:"phone_number" gorm:"type:varchar(15);not null"`
}

type OrderItems []OrderItem

func (oi *OrderItems) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Scan: unable to scan value into OrderItem")
	}
	return json.Unmarshal(bytes, oi)
}

// Реализация driver.Valuer для OrderItem
func (oi OrderItems) Value() (driver.Value, error) {
	return json.Marshal(oi)
}
