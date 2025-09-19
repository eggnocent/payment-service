package models

import (
	"payment-service/constants"
	"time"
)

type PaymentHistory struct {
	ID        uint                    `gorm:"primary_key;autoIncrement"`
	PaymentID uint                    `gorm:"type:bigint;not null"`
	Status    constants.PaymentStatus `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time               `gorm:"type:timestamp;not null"`
	UpdatedAt time.Time               `gorm:"type:timestamp;not null"`
}
