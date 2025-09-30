package repositories

import (
	"context"
	"gorm.io/gorm"
	error2 "payment-service/common/error"
	errConstant "payment-service/constants/error"
	"payment-service/domain/dto"
	"payment-service/domain/models"
)

type PaymentHistoryRepository struct {
	db *gorm.DB
}

type IPaymentHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.PaymentHistoryRequest) error
}

func NewPaymentHistoryRepository(db *gorm.DB) IPaymentHistoryRepository {
	return &PaymentHistoryRepository{db: db}
}

func (r *PaymentHistoryRepository) Create(ctx context.Context, tx *gorm.DB, req *dto.PaymentHistoryRequest) error {
	paymentHistory := &models.PaymentHistory{
		PaymentID: req.PaymentID,
		Status:    req.Status,
	}

	err := tx.Create(paymentHistory).Error
	if err != nil {
		return error2.WrapError(errConstant.ErrSQLError)
	}

	return nil
}
