package routes

import (
	"payment-service/clients"
	"payment-service/constants"
	controllers "payment-service/controllers/http"
	"payment-service/middlewares"

	"github.com/gin-gonic/gin"
)

type PaymentRoutes struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

type IPaymentRoutes interface {
	Run()
}

func NewPaymentRoutes(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) IPaymentRoutes {
	return &PaymentRoutes{
		group:      group,
		controller: controller,
		client:     client,
	}
}

func (p *PaymentRoutes) Run() {
	p.group.POST("/webhook", p.controller.GetPayment().Webhook)
	group := p.group.Group("/payments")

	group.Use(middlewares.Authenticate())
	group.GET("", middlewares.CheckRole([]string{constants.Admin, constants.Customer}, p.client), p.controller.GetPayment().GetAllWithPagination)
	group.GET("/:uuid", middlewares.CheckRole([]string{constants.Admin, constants.Customer}, p.client), p.controller.GetPayment().GetByUUID)
	group.POST("", middlewares.CheckRole([]string{constants.Customer}, p.client), p.controller.GetPayment().Create)
}
