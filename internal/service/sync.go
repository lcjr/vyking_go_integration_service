package service

import (
	"context"
	"log/slog"
	"time"

	"vyking-service/internal/domain"
	"vyking-service/internal/helpwise"
)

type SyncService struct {
	repo          domain.Repository
	hwClient      *helpwise.Client
	brandAPICreds map[uint32]string // BrandID -> API Key
}

func NewSyncService(repo domain.Repository, hwClient *helpwise.Client, creds map[uint32]string) *SyncService {
	return &SyncService{
		repo:          repo,
		hwClient:      hwClient,
		brandAPICreds: creds,
	}
}

func (s *SyncService) ProcessQueueBatch(ctx context.Context, batchSize int) error {
	items, err := s.repo.FetchPendingSyncQueue(ctx, batchSize)
	if err != nil || len(items) == 0 {
		return err
	}

	processedIDs := make([]int64, 0, len(items))

	for _, item := range items {
		cust, err := s.repo.GetCustomerWithCurrency(ctx, item.CustomerID)
		if err != nil {
			slog.Error("failed to get customer", "customer_id", item.CustomerID, "err", err)
			continue
		}

		apiKey, exists := s.brandAPICreds[cust.BrandID]
		if !exists {
			slog.Error("missing API credentials for brand", "brand_id", cust.BrandID)
			continue
		}

		contactID, err := s.hwClient.SyncContact(ctx, apiKey, cust)
		if err != nil {
			slog.Error("helpwise sync failed", "customer_id", cust.ID, "err", err)
			continue
		}

		if err := s.repo.UpsertPartnerContact(ctx, cust.ID, contactID, time.Now()); err != nil {
			slog.Error("failed to record partner contact", "customer_id", cust.ID, "err", err)
			continue
		}

		processedIDs = append(processedIDs, item.ID)
	}

	if len(processedIDs) > 0 {
		return s.repo.MarkSyncQueueProcessed(ctx, processedIDs, time.Now())
	}

	return nil
}
