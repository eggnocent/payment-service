package client

type MidtransResponse struct {
	Code   string       `json:"code"`
	Status string       `json:"status"`
	Data   MidtransData `json:"data"`
}

type MidtransData struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}
