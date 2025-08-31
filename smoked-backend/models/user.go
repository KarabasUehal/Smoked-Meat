package models

import "github.com/golang-jwt/jwt/v4"

type User struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" gorm:"type:text;unique;not null"`
	Password string `json:"password" gorm:"type:text;not null"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
