package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"vyking-service/internal/domain"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FetchPendingSyncQueue(ctx context.Context, batchSize int) ([]domain.SyncQueueItem, error) {
	args := m.Called(ctx, batchSize)
	return args.Get(0).([]domain.SyncQueueItem), args.Error(1)
}

func (m *MockRepository) GetCustomerWithCurrency(ctx context.Context, customerID uint32) (*domain.Customer, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockRepository) GetCustomerByEmailAndBrand(ctx context.Context, email string, brandID uint32) (*domain.Customer, error) {
	args := m.Called(ctx, email, brandID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockRepository) UpsertPartnerContact(ctx context.Context, customerID uint32, contactID string, syncedAt time.Time) error {
	args := m.Called(ctx, customerID, contactID, syncedAt)
	return args.Error(0)
}

func (m *MockRepository) MarkSyncQueueProcessed(ctx context.Context, queueIDs []int64, processedAt time.Time) error {
	args := m.Called(ctx, queueIDs, processedAt)
	return args.Error(0)
}

func (m *MockRepository) GetTransactionsByCustomer(ctx context.Context, customerID uint32, status string, limit int) ([]domain.Transaction, error) {
	args := m.Called(ctx, customerID, status, limit)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockRepository) GetBrandIDByPartnerToken(ctx context.Context, token string) (uint32, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(uint32), args.Error(1)
}

func TestMapDBStatusToHelpwise(t *testing.T) {
	tests := []struct {
		statusID uint32
		expected string
	}{
		{10, domain.HWStatusPending},
		{20, domain.HWStatusPending},
		{30, domain.HWStatusPending},
		{40, domain.HWStatusCompleted},
		{50, domain.HWStatusCompleted},
		{60, domain.HWStatusCompleted},
		{61, domain.HWStatusCompleted},
		{70, domain.HWStatusFailed},
		{71, domain.HWStatusFailed},
		{72, domain.HWStatusFailed},
		{80, domain.HWStatusCancelled},
		{81, domain.HWStatusCancelled},
		{82, domain.HWStatusCancelled},
		{90, domain.HWStatusFailed},
		{91, domain.HWStatusFailed},
		{999, domain.HWStatusFailed},
	}

	for _, tt := range tests {
		result := MapDBStatusToHelpwise(tt.statusID)
		assert.Equal(t, tt.expected, result, "Status ID %d mapping mismatch", tt.statusID)
	}
}

func TestFormatMinorUnits(t *testing.T) {
	assert.Equal(t, 12.34, FormatMinorUnits(1234, 2))
	assert.Equal(t, 100.0, FormatMinorUnits(100, 0))
	assert.Equal(t, 1.234, FormatMinorUnits(1234, 3))
}

func TestGetRecentTransactions_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewTransactionService(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	brandID := uint32(1)

	customer := &domain.Customer{
		ID:            101,
		BrandID:       brandID,
		Email:         email,
		CurrencyAlpha: "EUR",
		MinorUnits:    2,
	}

	txs := []domain.Transaction{
		{
			ID:          1,
			CustomerID:  101,
			TypeName:    "PAYIN",
			StatusID:    60,
			AmountMinor: 5000,
		},
	}

	mockRepo.On("GetCustomerByEmailAndBrand", ctx, email, brandID).Return(customer, nil)
	mockRepo.On("GetTransactionsByCustomer", ctx, uint32(101), "", 50).Return(txs, nil)

	res, err := svc.GetRecentTransactions(ctx, brandID, email, "", 50)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, 50.00, res[0].AmountFormatted)
	assert.Equal(t, "completed", res[0].HelpwiseStatus)
	assert.Equal(t, "EUR", res[0].Currency)

	mockRepo.AssertExpectations(t)
}
