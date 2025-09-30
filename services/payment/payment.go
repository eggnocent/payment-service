package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	client "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	"payment-service/common/util"
	"payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSlient
	kafka      kafka.IKafkaRegistry
	midtrans   client.IMidtransClient
}

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*util.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	WebHook(context.Context, *dto.WebHook) error
}

func NewPaymentService(repository repositories.IRepositoryRegistry, gcs gcs.IGCSlient, kafka kafka.IKafkaRegistry, midtrans client.IMidtransClient) IPaymentService {
	return &PaymentService{
		repository: repository,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}

func (p *PaymentService) GetAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) (*util.PaginationResult, error) {
	payment, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	paymentResult := make([]dto.PaymentResponse, 0, len(payment))
	for _, payment := range payment {
		paymentResult = append(paymentResult, dto.PaymentResponse{
			UUID:          payment.UUID,
			TransactionID: payment.TransactionID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			InvoiceLink:   payment.InvoiceLink,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			Description:   payment.Description,
			ExpiredAt:     payment.ExpiredAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	paginationParam := util.PaginationParams{
		Page:  param.Page,
		Limit: param.Limit,
		Count: total,
		Data:  paymentResult,
	}

	response := util.GeneratePagination(paginationParam)

	return &response, nil
}

func (p *PaymentService) GetByUUID(ctx context.Context, uuid string) (*dto.PaymentResponse, error) {
	log.Println("Fetching payment by UUID:", uuid)
	payment, err := p.repository.GetPayment().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return &dto.PaymentResponse{
		UUID:          payment.UUID,
		TransactionID: payment.TransactionID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		InvoiceLink:   payment.InvoiceLink,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		Description:   payment.Description,
		ExpiredAt:     payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

func (p *PaymentService) Create(ctx context.Context, request *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment // Deklarasi di luar transaction
		response   *dto.PaymentResponse
		midtrans   *client.MidtransData
	)

	// ...existing debug code...

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		if !request.ExpiredAt.After(time.Now()) {
			fmt.Printf("ERROR: ExpiredAt validation failed!\n")
			fmt.Printf("Validation ExpiredAt: %v\n", request.ExpiredAt)
			fmt.Printf("Validation time.Now(): %v\n", time.Now())
			fmt.Printf("Validation result: %v\n", request.ExpiredAt.After(time.Now()))
			return errPayment.ErrExpiredAtInvalid
		}

		fmt.Printf("ExpiredAt validation passed, proceeding to create payment link...\n")

		// Pre-Midtrans validation
		fmt.Printf("=== PRE-MIDTRANS VALIDATION ===\n")
		if p.midtrans == nil {
			fmt.Printf("FATAL ERROR: Midtrans client is NIL!\n")
			return fmt.Errorf("midtrans client not initialized")
		}
		fmt.Printf("Midtrans client OK: %p\n", p.midtrans)

		// Validate request components before sending to Midtrans
		if request.CustomerDetail == nil {
			fmt.Printf("ERROR: CustomerDetail is NIL!\n")
			return fmt.Errorf("customer detail is required")
		}

		if len(request.ItemDetail) == 0 {
			fmt.Printf("ERROR: ItemDetail is EMPTY!\n")
			return fmt.Errorf("item detail is required")
		}

		fmt.Printf("Pre-validation passed, calling Midtrans...\n")

		// Call Midtrans with detailed error catching
		fmt.Printf("About to call p.midtrans.CreatePaymentLink(request)...\n")
		midtrans, txErr = p.midtrans.CreatePaymentLink(request)

		if txErr != nil {
			fmt.Printf("ERROR: Failed to create payment link: %v\n", txErr)
			fmt.Printf("ERROR Type: %T\n", txErr)
			fmt.Printf("ERROR String: %s\n", txErr.Error())
			return txErr
		}

		if midtrans == nil {
			fmt.Printf("ERROR: Midtrans response is NIL!\n")
			return fmt.Errorf("midtrans response is nil")
		}

		fmt.Printf("Payment link created successfully: %s\n", midtrans.RedirectURL)
		fmt.Printf("Midtrans Token: %s\n", midtrans.Token)

		paymentRequest := &dto.PaymentRequest{
			OrderID:     request.OrderID,
			Amount:      request.Amount,
			Description: request.Description,
			ExpiredAt:   request.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}

		fmt.Printf("Creating payment in database...\n")
		fmt.Printf("PaymentRequest for DB: %+v\n", paymentRequest)

		// UBAH: Hapus deklarasi variabel baru, gunakan yang sudah ada
		payment, txErr = p.repository.GetPayment().Create(ctx, tx, paymentRequest)
		if txErr != nil {
			fmt.Printf("ERROR: Failed to create payment in database: %v\n", txErr)
			return txErr
		}

		if payment == nil {
			fmt.Printf("ERROR: Payment created but response is NIL!\n")
			return fmt.Errorf("payment creation returned nil")
		}

		fmt.Printf("Payment created successfully with UUID: %s\n", payment.UUID)

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: payment.ID,
			Status:    payment.Status.GetStatusString(),
		})
		if txErr != nil {
			fmt.Printf("ERROR: Failed to create payment history: %v\n", txErr)
			return txErr
		}

		fmt.Printf("Payment history created successfully\n")
		return nil
	})

	if err != nil {
		fmt.Printf("ERROR: Transaction failed: %v\n", err)
		return nil, err
	}

	if payment == nil {
		fmt.Printf("ERROR: Payment is NIL after transaction!\n")
		return nil, fmt.Errorf("payment is nil after transaction")
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}

	fmt.Printf("Payment creation completed successfully. UUID: %s\n", payment.UUID)
	return response, nil
}

func (p *PaymentService) ConvertToIndonesianMonth(englishMonth string) string {
	monthMap := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}

	indonesianMounth, ok := monthMap[englishMonth]
	if !ok {
		return errors.New("month not found").Error()
	}

	return indonesianMounth
}

func (p *PaymentService) GeneratePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "template/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	pdf, err := util.GeneratePDFfromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func (p *PaymentService) UploadToGCS(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	invoiceNumberReplace := strings.ToLower(strings.ReplaceAll(invoiceNumber, "/", "-"))
	fileName := fmt.Sprintf("%s.pdf", invoiceNumberReplace)
	url, err := p.gcs.UploadFile(ctx, fileName, pdf)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) + 100000 // Generate a number between 100000 and 999999
	return number
}

func (p *PaymentService) mapTransactionStatusToEvent(status constants.PaymentStatusString) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(string(constants.PendingString.String()))
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(string(constants.SettlementString.String()))
	case constants.ExpireString:
		paymentStatus = strings.ToUpper(string(constants.ExpireString.String()))
	default:
		paymentStatus = "UNKNOWN STATUS"
	}
	return paymentStatus
}

func (p *PaymentService) produceToKafka(req *dto.WebHook, payment *models.Payment, paidAt *time.Time) error {
	event := dto.KafkaEvent{
		Name: p.mapTransactionStatusToEvent(req.TransactionStatus),
	}

	metadata := dto.KafkaMetaData{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}

	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    req.TransactionStatus.String(),
			PaidAt:    paidAt,
			ExpiredAt: *payment.ExpiredAt,
		},
	}

	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		MetaData: metadata,
		Body:     body,
	}

	topic := config.Config.Kafka.Topic
	kafkaMessageJSON, _ := json.Marshal(kafkaMessage)
	err := p.kafka.GetKafkaProducer().ProduceMessage(topic, kafkaMessageJSON)
	if err != nil {
		return err
	}
	return nil
}

func (p *PaymentService) WebHook(ctx context.Context, req *dto.WebHook) error {
	var (
		// txErr, err         error
		paymentAfterUpdate *models.Payment
		paidAt             *time.Time
		invoiceLink        string
		pdf                []byte
	)

	// Debug logging untuk webhook request
	fmt.Printf("=== WEBHOOK DEBUG START ===\n")
	fmt.Printf("Request OrderID: %v\n", req.OrderID)
	fmt.Printf("Request OrderID String: %s\n", req.OrderID.String())
	fmt.Printf("Request TransactionID: %s\n", req.TransactionID)
	fmt.Printf("Request TransactionStatus: %v\n", req.TransactionStatus)
	fmt.Printf("Request TransactionStatus String: %s\n", req.TransactionStatus.String())
	fmt.Printf("Request PaymentType: %s\n", req.PaymentType)
	fmt.Printf("Request VANumbers: %+v\n", req.VANumbers)
	if len(req.VANumbers) > 0 {
		fmt.Printf("VANumbers[0] VaNumber: %s\n", req.VANumbers[0].VaNumber)
		fmt.Printf("VANumbers[0] Bank: %s\n", req.VANumbers[0].Bank)
	}
	fmt.Printf("Request Acquirer: %v\n", req.Acquirer)
	if req.Acquirer != nil {
		fmt.Printf("Request Acquirer Value: %s\n", *req.Acquirer)
	}
	fmt.Printf("===============================\n")

	err := p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		fmt.Printf("=== TRANSACTION START ===\n")

		// Find payment by OrderID
		fmt.Printf("Finding payment by OrderID: %s\n", req.OrderID.String())
		payment, txErr := p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			fmt.Printf("ERROR: Failed to find payment by OrderID: %v\n", txErr)
			fmt.Printf("ERROR Type: %T\n", txErr)
			fmt.Printf("ERROR String: %s\n", txErr.Error())
			return txErr
		}
		fmt.Printf("Payment found successfully: UUID=%s, Status=%v\n", payment.UUID, payment.Status)

		// Set paidAt if settlement
		if req.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
			fmt.Printf("Transaction is settlement, setting paidAt: %v\n", *paidAt)
		}

		// Prepare update data
		status := req.TransactionStatus.GetStatusInt()
		fmt.Printf("Status conversion: %v -> %v\n", req.TransactionStatus, status)

		var vaNumber, bank string
		if len(req.VANumbers) > 0 {
			vaNumber = req.VANumbers[0].VaNumber
			bank = req.VANumbers[0].Bank
			fmt.Printf("VA Details - Number: %s, Bank: %s\n", vaNumber, bank)
		} else {
			fmt.Printf("WARNING: No VANumbers provided in request\n")
		}

		// Update payment - UBAH: Handle update request berdasarkan payment method
		fmt.Printf("Updating payment with OrderID: %s\n", req.OrderID.String())
		updateRequest := &dto.UpdatePaymentRequest{
			TransactionID: &req.TransactionID,
			Status:        &status,
			PaidAt:        paidAt,
			Acquirer:      req.Acquirer,
		}

		// Only add VA/Bank if available (for bank transfer)
		if vaNumber != "" {
			updateRequest.VANumber = &vaNumber
		}
		if bank != "" {
			updateRequest.Bank = &bank
		}

		fmt.Printf("Update request: %+v\n", updateRequest)

		updatedPayment, txErr := p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), updateRequest)
		if txErr != nil {
			fmt.Printf("ERROR: Failed to update payment: %v\n", txErr)
			fmt.Printf("ERROR Type: %T\n", txErr)
			fmt.Printf("ERROR String: %s\n", txErr.Error())
			return txErr
		}
		fmt.Printf("Payment updated successfully: %+v\n", updatedPayment)

		// Get updated payment
		fmt.Printf("Fetching updated payment by OrderID: %s\n", req.OrderID.String())
		paymentAfterUpdate, txErr = p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			fmt.Printf("ERROR: Failed to fetch updated payment: %v\n", txErr)
			return txErr
		}
		fmt.Printf("Updated payment fetched: UUID=%s, Status=%v, Bank=%v, VANumber=%v\n",
			paymentAfterUpdate.UUID, paymentAfterUpdate.Status, paymentAfterUpdate.Bank, paymentAfterUpdate.VANumber)

		// Create payment history
		fmt.Printf("Creating payment history for PaymentID: %d\n", paymentAfterUpdate.ID)
		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: paymentAfterUpdate.ID,
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})
		if txErr != nil {
			fmt.Printf("ERROR: Failed to create payment history: %v\n", txErr)
			return txErr
		}
		fmt.Printf("Payment history created successfully\n")

		// Generate invoice if settlement
		if req.TransactionStatus == constants.SettlementString {
			fmt.Printf("=== INVOICE GENERATION START ===\n")

			if paidAt == nil {
				fmt.Printf("ERROR: paidAt is nil for settlement transaction\n")
				return fmt.Errorf("paidAt is nil for settlement transaction")
			}

			paidDay := paidAt.Format("02")
			paidMonth := p.ConvertToIndonesianMonth(paidAt.Format("January"))
			paidYear := paidAt.Format("2006")
			invoiceNumber := fmt.Sprintf("INV/%s/ORD/%d", time.Now().Format(time.DateOnly), p.randomNumber())

			fmt.Printf("Invoice details - Day: %s, Month: %s, Year: %s, Number: %s\n",
				paidDay, paidMonth, paidYear, invoiceNumber)

			// UBAH: Handle different payment methods untuk invoice
			var paymentMethodDisplay, bankDisplay, vaDisplay string

			switch req.PaymentType {
			case "qris":
				paymentMethodDisplay = "QRIS"
				bankDisplay = "Digital Payment"
				vaDisplay = "-"
				fmt.Printf("QRIS payment detected, using default values\n")
			case "bank_transfer":
				paymentMethodDisplay = "Bank Transfer"
				if paymentAfterUpdate.Bank != nil {
					bankDisplay = strings.ToUpper(*paymentAfterUpdate.Bank)
				} else {
					bankDisplay = "Unknown Bank"
				}
				if paymentAfterUpdate.VANumber != nil {
					vaDisplay = *paymentAfterUpdate.VANumber
				} else {
					vaDisplay = "-"
				}
				fmt.Printf("Bank transfer detected, Bank: %s, VA: %s\n", bankDisplay, vaDisplay)
			case "credit_card":
				paymentMethodDisplay = "Credit Card"
				bankDisplay = "Credit Card Payment"
				vaDisplay = "-"
				fmt.Printf("Credit card payment detected\n")
			default:
				paymentMethodDisplay = strings.ToUpper(req.PaymentType)
				bankDisplay = "Electronic Payment"
				vaDisplay = "-"
				fmt.Printf("Other payment method detected: %s\n", req.PaymentType)
			}

			// UBAH: Hanya cek Description (yang memang harus ada)
			if paymentAfterUpdate.Description == nil {
				fmt.Printf("ERROR: Description is nil in payment after update\n")
				return fmt.Errorf("description is nil in payment")
			}

			total := util.RupiahFormat(&paymentAfterUpdate.Amount)
			fmt.Printf("Formatted total: %s\n", total)

			invoiceRequest := &dto.InvoiceRequest{
				InvoiceNumber: invoiceNumber,
				Data: dto.InvoiceData{
					PaymentDetail: dto.InvoicePaymentDetail{
						PaymentMethod: paymentMethodDisplay, // UBAH: Gunakan variable yang sudah di-handle
						BankName:      bankDisplay,          // UBAH: Gunakan variable yang sudah di-handle
						VANumber:      vaDisplay,            // UBAH: Gunakan variable yang sudah di-handle
						Date:          fmt.Sprintf("%s %s %s", paidDay, paidMonth, paidYear),
						IsPaid:        true,
					},
					Items: []dto.InvoiceItem{
						{
							Description: *paymentAfterUpdate.Description,
							Price:       total,
						},
					},
					Total: total,
				},
			}

			fmt.Printf("Invoice request prepared: %+v\n", invoiceRequest)

			// Generate PDF
			fmt.Printf("Generating PDF for invoice...\n")
			pdf, txErr = p.GeneratePDF(invoiceRequest)
			if txErr != nil {
				fmt.Printf("ERROR: Failed to generate PDF: %v\n", txErr)
				return txErr
			}
			fmt.Printf("PDF generated successfully, size: %d bytes\n", len(pdf))

			// Upload to GCS
			fmt.Printf("Uploading PDF to GCS with invoice number: %s\n", invoiceNumber)
			invoiceLink, txErr = p.UploadToGCS(ctx, invoiceNumber, pdf)
			if txErr != nil {
				fmt.Printf("ERROR: Failed to upload to GCS: %v\n", txErr)
				return txErr
			}
			fmt.Printf("PDF uploaded successfully, link: %s\n", invoiceLink)

			// Update payment with invoice link
			fmt.Printf("Updating payment with invoice link...\n")
			_, txErr = p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				fmt.Printf("ERROR: Failed to update payment with invoice link: %v\n", txErr)
				return txErr
			}
			fmt.Printf("Payment updated with invoice link successfully\n")
			fmt.Printf("=== INVOICE GENERATION END ===\n")
		}

		fmt.Printf("=== TRANSACTION END ===\n")
		return nil
	})

	if err != nil {
		fmt.Printf("ERROR: Transaction failed: %v\n", err)
		return err
	}

	// Produce to Kafka
	fmt.Printf("=== KAFKA PRODUCTION START ===\n")
	fmt.Printf("Producing to Kafka with payment: %+v\n", paymentAfterUpdate)
	fmt.Printf("PaidAt: %v\n", paidAt)

	err = p.produceToKafka(req, paymentAfterUpdate, paidAt)
	if err != nil {
		fmt.Printf("ERROR: Failed to produce to Kafka: %v\n", err)
		return err
	}
	fmt.Printf("Kafka message produced successfully\n")
	fmt.Printf("=== KAFKA PRODUCTION END ===\n")

	fmt.Printf("=== WEBHOOK DEBUG END ===\n")
	return nil
}
