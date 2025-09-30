package services

import (
	clients "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	"payment-service/controllers/kafka"
	"payment-service/repositories"
	service "payment-service/services/payment"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSlient
	kafka      kafka.IKafkaRegistry
	midtrans   clients.IMidtransClient
}

type IServiceRegistry interface {
	GetPayment() service.IPaymentService
}

func NewServiceRegistry(
	repositories repositories.IRepositoryRegistry,
	gcs gcs.IGCSlient,
	kafka kafka.IKafkaRegistry,
	midtrans clients.IMidtransClient,
) IServiceRegistry {
	return &Registry{
		repository: repositories,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}

func (r *Registry) GetPayment() service.IPaymentService {
	return service.NewPaymentService(r.repository, r.gcs, r.kafka, r.midtrans)
}
