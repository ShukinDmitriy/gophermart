package models

import (
	"github.com/ShukinDmitriy/gophermart/internal/entities"
)

type GetOrdersResponse struct {
	Number     string               `json:"number"`
	Status     entities.OrderStatus `json:"status"`
	Accrual    float32              `json:"accrual,omitempty"`
	UploadedAt JSONTime             `json:"uploaded_at"`
}
