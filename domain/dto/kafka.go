package dto

import (
	"github.com/google/uuid"
	"time"
)

type KafkaEvent struct {
	Name string `json:"name"`
}

type KafkaMetaData struct {
	Sender    string `json:"sender"`
	SendingAt string `json:"sending_at"`
}

type KafkaData struct {
	OrderID   uuid.UUID  `json:"order_id"`
	PaymentID uuid.UUID  `json:"payment_id"`
	Status    string     `json:"status"`
	ExpiredAt time.Time  `json:"expired_at"`
	PaidAt    *time.Time `json:"paid_at"`
}

type KafkaBody struct {
	Type string     `json:"type"`
	Data *KafkaData `json:"data"`
}

type KafkaMessage struct {
	Event    KafkaEvent    `json:"event"`
	MetaData KafkaMetaData `json:"meta_data"`
	Body     KafkaBody     `json:"body"`
}
