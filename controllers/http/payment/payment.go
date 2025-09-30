package controllers

import (
	"fmt"
	"io"
	"net/http"
	"payment-service/common/response"
	"payment-service/domain/dto"
	"payment-service/services"
	"strings"

	errValidation "payment-service/common/error"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	service services.IServiceRegistry
}

type IPaymentController interface {
	GetAllWithPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Webhook(*gin.Context)
}

func NewPaymentController(service services.IServiceRegistry) IPaymentController {
	return &PaymentController{
		service: service,
	}
}

func (p *PaymentController) GetAllWithPagination(c *gin.Context) {
	var param dto.PaymentRequestParam
	err := c.ShouldBindQuery(&param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})

		return
	}

	validate := validator.New()
	if err = validate.Struct(param); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})

		return
	}

	result, err := p.service.GetPayment().GetAllWithPagination(c.Request.Context(), &param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})

		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})

	c.JSON(http.StatusOK, result)
}

func (p *PaymentController) GetByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	result, err := p.service.GetPayment().GetByUUID(c.Request.Context(), uuid)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})

		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (p *PaymentController) Create(c *gin.Context) {
	var req dto.PaymentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})

		return
	}

	validate := validator.New()
	if err = validate.Struct(req); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})

		return
	}

	result, err := p.service.GetPayment().Create(c.Request.Context(), &req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})

		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Gin:  c,
	})
}

func (p *PaymentController) Webhook(c *gin.Context) {
	// Debug: Log raw request body
	body, _ := c.GetRawData()
	fmt.Printf("=== RAW WEBHOOK REQUEST ===\n")
	fmt.Printf("Method: %s\n", c.Request.Method)
	fmt.Printf("Headers: %+v\n", c.Request.Header)
	fmt.Printf("Content-Type: %s\n", c.GetHeader("Content-Type"))
	fmt.Printf("Body: %s\n", string(body))
	fmt.Printf("Body Length: %d\n", len(body))
	fmt.Printf("========================\n")

	// Reset body untuk parsing
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	var request dto.WebHook
	err := c.ShouldBindJSON(&request)
	if err != nil {
		fmt.Printf("ERROR: Failed to bind JSON: %v\n", err)
		fmt.Printf("ERROR Type: %T\n", err)

		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	fmt.Printf("Successfully parsed webhook request: %+v\n", request)

	err = p.service.GetPayment().WebHook(c.Request.Context(), &request)
	if err != nil {
		fmt.Printf("ERROR: Webhook service failed: %v\n", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}
