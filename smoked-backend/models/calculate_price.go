package models

type CalculatePriceRequest struct {
	ID       int     `json:"id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gte=0"`
}
