package entities

import (
	"time"

	"gorm.io/gorm"
)

type OperationType string

const (
	OperationTypeAccrual  OperationType = "accrual"
	OperationTypeWithdraw OperationType = "withdraw"
)

type Operation struct {
	gorm.Model
	ProcessedAt        time.Time     `json:"processedAt"`
	Type               OperationType `json:"type"`
	OrderNumber        string        `json:"orderNumber"`
	Sum                float32       `json:"sum"`
	SenderAccountID    uint          `json:"senderAccountId"`
	SenderAccount      Account       `json:"senderAccount"`
	RecipientAccountID uint          `json:"recipientAccountId"`
	RecipientAccount   Account       `json:"recipientAccount"`
}
