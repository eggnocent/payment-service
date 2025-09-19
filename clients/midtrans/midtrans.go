package client

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	error2 "payment-service/constants/error/payment"
	"payment-service/domain/dto"
	"time"
)

type MidtransClient struct {
	ServerKey    string
	IsProduction bool
}

type IMidtransClient interface {
	CreatePaymentLink(request *dto.PaymentRequest) (*MidtransData, error)
}

func NewMidtransClient(serverKey string, isProduction bool) *MidtransClient {
	return &MidtransClient{
		ServerKey:    serverKey,
		IsProduction: isProduction,
	}
}

func (m *MidtransClient) CreatePaymentLink(request *dto.PaymentRequest) (*MidtransData, error) {
	var (
		snapClient   snap.Client
		ISProduction = midtrans.Sandbox
	)

	expiryDateTime := request.ExpiredAt
	currentTime := time.Now()
	duration := expiryDateTime.Sub(currentTime)
	if duration <= 0 {
		logrus.Errorf("expiryDateTime is invalid")
		return nil, error2.ErrExpiredAtInvalid
	}

	expiryUnit := "minute"
	expiryDuration := int64(duration.Minutes())
	if duration.Hours() >= 1 {
		expiryUnit = "hour"
		expiryDuration = int64(duration.Hours())
	} else if duration.Hours() >= 24 {
		expiryUnit = "day"
		expiryDuration = int64(duration.Hours() / 24)
	}

	if m.IsProduction {
		ISProduction = midtrans.Production
	}

	snapClient.New(m.ServerKey, ISProduction)
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  request.OrderID,
			GrossAmt: int64(request.Amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: request.CustomerDetail.Name,
			Email: request.CustomerDetail.Email,
			Phone: request.CustomerDetail.Phone,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    request.ItemDetail[0].ID,
				Price: int64(request.ItemDetail[0].Amount),
				Qty:   int32(request.ItemDetail[0].Quantity),
				Name:  request.ItemDetail[0].Name,
			},
		},
		Expiry: &snap.ExpiryDetails{
			Unit:     expiryUnit,
			Duration: expiryDuration,
		},
	}

	response, err := snapClient.CreateTransaction(req)
	if err != nil {
		logrus.Errorf("snapClient.CreateTransaction err: %v", err)
		return nil, err
	}

	return &MidtransData{
		RedirectURL: response.RedirectURL,
		Token:       response.Token,
	}, nil
}
