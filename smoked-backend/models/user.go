package models

import (
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username    string `json:"username" gorm:"type:text;unique;not null"`
	Password    string `json:"password" gorm:"type:text;min=6;not null"`
	PhoneNumber string `json:"phone_number" gorm:"type:varchar(15);unique;not null"`
	Role        string `json:"role" gorm:"not null;default:'client'"`
	Name        string `json:"name" gorm:"type:text"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
	Role        string `json:"role"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
}
