package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// PaymentGateway defines the loosely coupled interface for any payment provider
type PaymentGateway interface {
	InitializeTransaction(email string, amount float64, reference string) (authURL string, gatewayRef string, err error)
	VerifyTransaction(reference string) (isValid bool, amount float64, err error)
	CreateTransferRecipient(name, accountNumber, bankCode string) (string, error)
	InitiateTransfer(amount float64, recipientCode, reference string) (string, error)
}

// PaystackGateway implements PaymentGateway
type PaystackGateway struct {
	SecretKey string
	BaseURL   string
}

func NewPaystackGateway(secretKey string) *PaystackGateway {
	return &PaystackGateway{
		SecretKey: secretKey,
		BaseURL:   "https://api.paystack.co",
	}
}

func (p *PaystackGateway) InitializeTransaction(email string, amount float64, reference string) (string, string, error) {
	if p.SecretKey == "" {
		return "", "", fmt.Errorf("paystack secret key is empty")
	}

	payload := map[string]interface{}{
		"email":        email,
		"amount":       int(amount * 100), // Paystack requires kobo/cents
		"reference":    reference,
		"callback_url": os.Getenv("FRONTEND_URL") + "/wallet/callback",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/transaction/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if !result.Status {
		return "", "", fmt.Errorf("paystack error: %s", result.Message)
	}

	return result.Data.AuthorizationURL, result.Data.Reference, nil
}

func (p *PaystackGateway) VerifyTransaction(reference string) (bool, float64, error) {
	if p.SecretKey == "" {
		return false, 0, fmt.Errorf("paystack secret key is empty")
	}

	req, err := http.NewRequest("GET", p.BaseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+p.SecretKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status string  `json:"status"`
			Amount float64 `json:"amount"` // in kobo
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, 0, err
	}

	if !result.Status {
		return false, 0, fmt.Errorf("verify API error: %s", result.Message)
	}

	isSuccess := result.Data.Status == "success"
	realAmount := result.Data.Amount / 100.0 // Convert kobo to Naira

	return isSuccess, realAmount, nil
}

func (p *PaystackGateway) CreateTransferRecipient(name, accountNumber, bankCode string) (string, error) {
	if p.SecretKey == "" {
		return "", fmt.Errorf("paystack secret key is empty")
	}

	payload := map[string]interface{}{
		"type":           "nuban",
		"name":           name,
		"account_number": accountNumber,
		"bank_code":      bankCode,
		"currency":       "NGN",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/transferrecipient", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			RecipientCode string `json:"recipient_code"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Status {
		return "", fmt.Errorf("paystack recipient error: %s", result.Message)
	}

	return result.Data.RecipientCode, nil
}

func (p *PaystackGateway) InitiateTransfer(amount float64, recipientCode, reference string) (string, error) {
	if p.SecretKey == "" {
		return "", fmt.Errorf("paystack secret key is empty")
	}

	payload := map[string]interface{}{
		"source":    "balance",
		"amount":    int(amount * 100), // convert to kobo
		"recipient": recipientCode,
		"reference": reference,
		"reason":    "Car Wash Payout",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/transfer", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			TransferCode string `json:"transfer_code"`
			Status       string `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Status {
		return "", fmt.Errorf("paystack transfer error: %s", result.Message)
	}

	return result.Data.TransferCode, nil
}
