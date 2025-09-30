package dto

import (
	"payment-service/constants"
	"time"

	"github.com/google/uuid"
)

type PaymentRequest struct {
	PaymentLink    string          `json:"payment_link"`
	OrderID        string          `json:"orderId"`
	ExpiredAt      time.Time       `json:"expiredAt"`
	Amount         float64         `json:"amount"`
	Description    *string         `json:"description"`
	CustomerDetail *CustomerDetail `json:"customerDetail"`
	ItemDetail     []ItemDetail    `json:"itemDetails"`
}

type CustomerDetail struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type ItemDetail struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
}

type PaymentRequestParam struct {
	Page       int     `form:"page" validate:"required"`
	Limit      int     `form:"limit" validate:"required"`
	SortColumn *string `form:"sort_column" validate:"required"`
	SortOrder  *string `form:"sort_order" validate:"required"`
}

type UpdatePaymentRequest struct {
	TransactionID *string                  `form:"transaction_id"`
	Status        *constants.PaymentStatus `form:"status"`
	PaidAt        *time.Time               `form:"paid_at"`
	VANumber      *string                  `form:"va_number"`
	Bank          *string                  `form:"bank"`
	InvoiceLink   *string                  `form:"invoice_link,omitempty"`
	Acquirer      *string                  `form:"acquirer"`
}

type PaymentResponse struct {
	UUID          uuid.UUID                     `json:"uuid"`
	OrderID       uuid.UUID                     `json:"order_id"`
	Amount        float64                       `json:"amount"`
	Status        constants.PaymentStatusString `json:"status"`
	PaymentLink   string                        `json:"payment_link"`
	InvoiceLink   *string                       `json:"invoice_link"`
	TransactionID *string                       `form:"transaction_id,omitempty"`
	PaidAt        *time.Time                    `form:"paid_at,omitempty"`
	VANumber      *string                       `form:"va_number,omitempty"`
	Bank          *string                       `form:"bank,omitempty"`
	Acquirer      *string                       `form:"acquirer,omitempty"`
	Description   *string                       `form:"description"`
	ExpiredAt     *time.Time                    `json:"expired_at"`
	CreatedAt     *time.Time                    `json:"created_at"`
	UpdatedAt     *time.Time                    `json:"updated_at"`
}

type WebHook struct {
	VANumbers         []VANumber                    `json:"va_numbers"`
	TransactionTime   string                        `json:"transaction_time"`
	TransactionStatus constants.PaymentStatusString `json:"transaction_status"`
	TransactionID     string                        `json:"transaction_id"`
	StatusMessage     string                        `json:"status_message"`
	StatusCode        string                        `json:"status_code"`
	SignatureKey      string                        `json:"signature_key"`
	SettlementTime    string                        `json:"settlement_time"`
	PaymentType       string                        `json:"payment_type"`
	PaymentAmount     []PaymentAmount               `json:"payment_amount"`
	OrderID           uuid.UUID                     `json:"order_id"`
	MerchantID        string                        `json:"merchant_id"`
	GrossAmount       string                        `json:"gross_amount"`
	FraudStatus       string                        `json:"fraud_status"`
	Currency          string                        `json:"currency"`
	Acquirer          *string                       `json:"acquirer"`
}

type VANumber struct {
	VaNumber string `json:"va_number"`
	Bank     string `json:"bank"`
}

type PaymentAmount struct {
	PaidAt *string `json:"paid_at"`
	Amount *string `json:"amount"`
}
