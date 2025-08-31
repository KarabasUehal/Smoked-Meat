package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Assortment struct {
	ID           int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Meat         string     `json:"meat" gorm:"type:text;not null"`
	Availability bool       `json:"avail" gorm:"type:boolean;not null"`
	Price        float64    `json:"price" gorm:"type:double precision;not null"`
	Spice        Spice      `json:"spice" gorm:"type:jsonb;not null"`
	CreatedAt    time.Time  `json:"-" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"-" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"-" gorm:"index"`
}

type Spice struct {
	Recipe1 string `json:"recipe1" gorm:"type:text;not null"`
	Recipe2 string `json:"recipe2" gorm:"type:text;not null"`
}

// Реализация sql.Scanner для Spice
func (s *Spice) Scan(value interface{}) error {
	bytes, ok := value.([]uint8)
	if !ok {
		return errors.New("Scan: unable to scan value into Spice")
	}
	return json.Unmarshal(bytes, s)
}

// Реализация driver.Valuer для Spice
func (s Spice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type ProductRequest struct {
	Meat  string  `json:"meat" gorm:"type:text;not null" binding:"required"`
	Avail bool    `json:"avail" gorm:"type:boolean;not null" binding:"required"`
	Price float64 `json:"price" gorm:"type:double precision;not null" binding:"required,gt=0"`
	Spice Spice   `json:"spice" gorm:"type:jsonb;not null" binding:"required"`
}
