package services

import (
	"math"

	"github.com/olabanji12-ojo/CarWashApp/config"
)

type FinanceService struct {
	PlatformFeePercent     float64
	CancellationFeePercent float64
}

func NewFinanceService() *FinanceService {
	// Initialize with percentage from config
	return &FinanceService{
		PlatformFeePercent:     config.Cfg.PlatformFee,
		CancellationFeePercent: 0.05, // 5% default cancellation fee
	}
}

// CalculateSplit takes the total amount and returns (PlatformFee, MerchantNet)
func (fs *FinanceService) CalculateSplit(totalAmount float64) (float64, float64) {
	// We use the loosely coupled percentage from config
	platformFee := totalAmount * fs.PlatformFeePercent
	
	// Round to 2 decimal places for financial accuracy
	platformFee = math.Round(platformFee*100) / 100
	merchantNet := totalAmount - platformFee
	
	return platformFee, merchantNet
}

// CalculateRefundAmount returns the amount to be returned to the user after the cancellation fee
func (fs *FinanceService) CalculateRefundAmount(totalAmount float64) (float64, float64) {
	cancellationFee := totalAmount * fs.CancellationFeePercent
	cancellationFee = math.Round(cancellationFee*100) / 100
	refundAmount := totalAmount - cancellationFee
	
	return refundAmount, cancellationFee
}

// GetTotalWithServiceFee calculates what the customer pays if we add the fee on top
// (e.g. Wash is $20, we add $1 fee, Total = $21)
func (fs *FinanceService) GetTotalWithServiceFee(basePrice float64) float64 {
	fee := basePrice * fs.PlatformFeePercent
	return basePrice + (math.Round(fee*100) / 100)
}
