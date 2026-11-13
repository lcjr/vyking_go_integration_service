package service

import (
	"context"
	"errors"
	"math"

	"vyking-service/internal/domain"
)

type TransactionService struct {
	repo domain.Repository
}

func NewTransactionService(repo domain.Repository) *TransactionService {
	return &TransactionService{repo: repo}
}

// Comment:
// MapDBStatusToHelpwise translates internal DB status codes to Helpwise vocabulary
func MapDBStatusToHelpwise(statusID uint32) string {
	switch statusID {
	case 10, 20, 30: // reserved, pending_provider, awaiting_approval
		return domain.HWStatusPending
	case 40, 50, 60, 61: // approved, booked, settled, settled_late
		return domain.HWStatusCompleted
	case 70, 71, 72, 90, 91: // declined_*, reversed, chargeback
		return domain.HWStatusFailed
	case 80, 81, 82: // cancelled_*
		return domain.HWStatusCancelled
	default:
		return domain.HWStatusFailed
	}
}

func FormatMinorUnits(amount int64, minorUnits uint8) float64 {
	if minorUnits == 0 {
		return float64(amount)
	}
	return float64(amount) / math.Pow10(int(minorUnits))
}

func (s *TransactionService) GetRecentTransactions(ctx context.Context, brandID uint32, email, statusFilter string, limit int) ([]domain.Transaction, error) {
	customer, err := s.repo.GetCustomerByEmailAndBrand(ctx, email, brandID)
	if err != nil {
		return nil, errors.New("CUSTOMER_NOT_FOUND")
	}

	txs, err := s.repo.GetTransactionsByCustomer(ctx, customer.ID, statusFilter, limit)
	if err != nil {
		return nil, err
	}

	for i := range txs {
		txs[i].AmountFormatted = FormatMinorUnits(txs[i].AmountMinor, customer.MinorUnits)
		txs[i].HelpwiseStatus = MapDBStatusToHelpwise(txs[i].StatusID)
		txs[i].Currency = customer.CurrencyAlpha
	}

	return txs, nil
}
