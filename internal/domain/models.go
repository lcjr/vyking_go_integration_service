package domain

import (
	"context"
	"time"
)

type SyncQueueItem struct {
	ID         int64
	CustomerID uint32
	QueuedAt   time.Time
}

type Customer struct {
	ID            uint32
	BrandID       uint32
	Email         string
	FirstName     string
	LastName      string
	Phone         string
	Country       string
	CurrencyAlpha string
	MinorUnits    uint8
	Active        bool
	CreatedAt     time.Time
}

type Transaction struct {
	ID              uint64     `json:"id"`
	CustomerID      uint32     `json:"customer_id"`
	TypeName        string     `json:"type"`
	StatusID        uint32     `json:"-"`
	HelpwiseStatus  string     `json:"status"`
	AmountMinor     int64      `json:"-"`
	AmountFormatted float64    `json:"amount"`
	Currency        string     `json:"currency"`
	PaymentMethod   string     `json:"payment_method"`
	Created         time.Time  `json:"created_at"`
	Booked          *time.Time `json:"settled_at,omitempty"`
}

const (
	HWStatusPending   = "pending"
	HWStatusCompleted = "completed"
	HWStatusFailed    = "failed"
	HWStatusCancelled = "cancelled"
)

type Repository interface {
	FetchPendingSyncQueue(ctx context.Context, batchSize int) ([]SyncQueueItem, error)
	GetCustomerWithCurrency(ctx context.Context, customerID uint32) (*Customer, error)
	GetCustomerByEmailAndBrand(ctx context.Context, email string, brandID uint32) (*Customer, error)
	UpsertPartnerContact(ctx context.Context, customerID uint32, contactID string, syncedAt time.Time) error
	MarkSyncQueueProcessed(ctx context.Context, queueIDs []int64, processedAt time.Time) error
	GetTransactionsByCustomer(ctx context.Context, customerID uint32, status string, limit int) ([]Transaction, error)
	GetBrandIDByPartnerToken(ctx context.Context, token string) (uint32, error)
}
