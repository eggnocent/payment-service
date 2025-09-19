package dto

type InvoiceRequest struct {
	InvoiceNumber string      `json:"invoice_number"`
	Data          InvoiceData `json:"data"`
}

type InvoiceData struct {
	PaymentDetail InvoicePaymentDetail `json:"payment_detail"`
	Items         []InvoiceItem        `json:"items"`
	Total         string               `json:"total"`
}

type InvoicePaymentDetail struct {
	BankName      string `json:"bank_name"`
	PaymentMethod string `json:"payment_method"`
	VANumber      string `json:"va_number"`
	Date          string `json:"date"`
	IsPaid        string `json:"is_paid"`
}

type InvoiceItem struct {
	Description string `json:"description"`
	Price       string `json:"price"`
}
