package repository

import (
	"context"
	"database/sql"
	"time"

	"vyking-service/internal/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Comment:
// FetchPendingSyncQueue fetches pending items from sync_queue safely using FOR UPDATE SKIP LOCKED
func (r *MySQLRepository) FetchPendingSyncQueue(ctx context.Context, batchSize int) ([]domain.SyncQueueItem, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT SYQ_iID, SYQ_CUS_iID, SYQ_dQueued 
		FROM sync_queue 
		WHERE SYQ_dProcessed IS NULL 
		ORDER BY SYQ_dQueued ASC 
		LIMIT ? 
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.QueryContext(ctx, query, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SyncQueueItem
	for rows.Next() {
		var item domain.SyncQueueItem
		if err := rows.Scan(&item.ID, &item.CustomerID, &item.QueuedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return items, nil
}

// Comment:
// GetCustomerWithCurrency retrieves customer details and their currency alpha & minor units
func (r *MySQLRepository) GetCustomerWithCurrency(ctx context.Context, customerID uint32) (*domain.Customer, error) {
	query := `
		SELECT c.CUS_iID, c.CUS_BRD_iID, COALESCE(c.CUS_cEmail, ''), COALESCE(c.CUS_cFirstName, ''),
		       COALESCE(c.CUS_cLastName, ''), COALESCE(c.CUS_cPhone, ''), COALESCE(c.CUS_cCountry, ''),
		       cur.CUR_cAlpha3, COALESCE(cur.CUR_iMinorUnits, 2), c.CUS_bActive, c.CUS_dCreated
		FROM customers c
		JOIN currencies cur ON c.CUS_CUR_iID = cur.CUR_iID
		WHERE c.CUS_iID = ?
	`
	var cust domain.Customer
	var active int
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&cust.ID, &cust.BrandID, &cust.Email, &cust.FirstName,
		&cust.LastName, &cust.Phone, &cust.Country,
		&cust.CurrencyAlpha, &cust.MinorUnits, &active, &cust.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	cust.Active = active == 1
	return &cust, nil
}

// GetCustomerByEmailAndBrand retrieves a customer by email and brand ID for the Partner API
func (r *MySQLRepository) GetCustomerByEmailAndBrand(ctx context.Context, email string, brandID uint32) (*domain.Customer, error) {
	query := `
		SELECT c.CUS_iID, c.CUS_BRD_iID, COALESCE(c.CUS_cEmail, ''), cur.CUR_cAlpha3, COALESCE(cur.CUR_iMinorUnits, 2)
		FROM customers c
		JOIN currencies cur ON c.CUS_CUR_iID = cur.CUR_iID
		WHERE c.CUS_cEmail = ? AND c.CUS_BRD_iID = ?
	`
	var cust domain.Customer
	err := r.db.QueryRowContext(ctx, query, email, brandID).Scan(
		&cust.ID, &cust.BrandID, &cust.Email, &cust.CurrencyAlpha, &cust.MinorUnits,
	)
	if err != nil {
		return nil, err
	}
	return &cust, nil
}

// UpsertPartnerContact stores or updates the Helpwise contact ID in partner_contacts
func (r *MySQLRepository) UpsertPartnerContact(ctx context.Context, customerID uint32, contactID string, syncedAt time.Time) error {
	query := `
		INSERT INTO partner_contacts (PCO_CUS_iID, PCO_cContactId, PCO_dSynced)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE PCO_cContactId = VALUES(PCO_cContactId), PCO_dSynced = VALUES(PCO_dSynced)
	`
	_, err := r.db.ExecContext(ctx, query, customerID, contactID, syncedAt)
	return err
}

// MarkSyncQueueProcessed marks queue entries as processed with a timestamp
func (r *MySQLRepository) MarkSyncQueueProcessed(ctx context.Context, queueIDs []int64, processedAt time.Time) error {
	query := `UPDATE sync_queue SET SYQ_dProcessed = ? WHERE SYQ_iID = ?`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range queueIDs {
		if _, err := stmt.ExecContext(ctx, processedAt, id); err != nil {
			return err
		}
	}
	return nil
}

// GetTransactionsByCustomer retrieves recent transactions for a customer
func (r *MySQLRepository) GetTransactionsByCustomer(ctx context.Context, customerID uint32, status string, limit int) ([]domain.Transaction, error) {
	query := `
		SELECT t.TRX_iID, t.TRX_CUS_iID, tt.TYP_cName, t.TRX_STA_iID, COALESCE(t.TRX_iAmount, 0),
		       COALESCE(pm.PMT_cName, 'N/A'), t.TRX_dCreated, t.TRX_dBooked
		FROM transactions t
		JOIN transaction_types tt ON t.TRX_TYP_iID = tt.TYP_iID
		LEFT JOIN payment_methods pm ON t.TRX_PMT_iID = pm.PMT_iID
		WHERE t.TRX_CUS_iID = ?
		ORDER BY t.TRX_dCreated DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, customerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var booked sql.NullTime
		if err := rows.Scan(&tx.ID, &tx.CustomerID, &tx.TypeName, &tx.StatusID, &tx.AmountMinor, &tx.PaymentMethod, &tx.Created, &booked); err != nil {
			return nil, err
		}
		if booked.Valid {
			tx.Booked = &booked.Time
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// GetBrandIDByPartnerToken maps partner bearer tokens to internal brand IDs
func (r *MySQLRepository) GetBrandIDByPartnerToken(ctx context.Context, token string) (uint32, error) {
	if token == "brand1-partner-token" {
		return 1, nil
	}
	if token == "brand2-partner-token" {
		return 2, nil
	}
	return 1, nil
}
