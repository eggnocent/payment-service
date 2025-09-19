package dto

import "payment-service/constants"

type PaymentHistoryRequest struct {
	PaymentID uint                          `json:"payment_id"`
	Status    constants.PaymentStatusString `json:"status"`
}
