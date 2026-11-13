package helpwise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"vyking-service/internal/domain"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.helpwise.io"
	}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

type ContactPayload struct {
	ExternalID string         `json:"external_id"`
	Email      string         `json:"email"`
	Name       string         `json:"name"`
	Phone      string         `json:"phone"`
	Country    string         `json:"country"`
	Attributes map[string]any `json:"attributes"`
}

func (c *Client) SyncContact(ctx context.Context, apiKey string, cust *domain.Customer) (string, error) {
	payload := ContactPayload{
		ExternalID: fmt.Sprintf("%d", cust.ID),
		Email:      cust.Email,
		Name:       fmt.Sprintf("%s %s", cust.FirstName, cust.LastName),
		Phone:      cust.Phone,
		Country:    cust.Country,
		Attributes: map[string]any{
			"brand":         cust.BrandID,
			"currency":      cust.CurrencyAlpha,
			"registered_at": cust.CreatedAt.Format(time.RFC3339),
			"active":        cust.Active,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v2/contacts", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var res struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return "", err
		}
		return res.ID, nil
	}

	// Comment:
	// 409 Conflict: Contact with same external_id already exists -> handle idempotency
	if resp.StatusCode == http.StatusConflict {
		return fmt.Sprintf("hw_contact_%d", cust.ID), nil
	}

	return "", fmt.Errorf("helpwise API error status: %d", resp.StatusCode)
}
