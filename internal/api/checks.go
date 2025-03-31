package api

import (
	"bytes"
	"context"
	"encoding/json"
	api_types "github.com/scayle/terraform-provider-pingdom/internal/api/types"
	"net/http"
	"net/url"
)

func (client *client) GetCheck(ctx context.Context, id string) (*api_types.Check, error) {
	uri, err := url.JoinPath(client.baseURL, "checks", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if err != nil {
		return nil, err
	}

	var res *struct {
		Check api_types.Check `json:"check"`
	}
	err = client.do(req, &res)
	if err != nil {
		return nil, err
	}

	return &res.Check, nil
}

type CreateCheckRequest struct {
	Name string `json:"name"`
	Host string `json:"host"`
	// Type needs to be empty for update requests
	Type                     string   `json:"type,omitempty"`
	Encryption               bool     `json:"encryption"`
	CustomMessage            string   `json:"custom_message,omitempty"`
	ProbeFilters             []string `json:"probe_filters,omitempty"`
	Resolution               float64  `json:"resolution"`
	Auth                     string   `json:"auth,omitempty"`
	NotifyAgainEvery         int64    `json:"notifyagainevery"`
	NotifyWhenBackup         bool     `json:"notifywhenbackup"`
	Paused                   bool     `json:"paused"`
	Port                     int64    `json:"port,omitempty"`
	ResponseTimeThreshold    int64    `json:"responsetime_threshold"`
	SendNotificationWhenDown int64    `json:"sendnotificationwhendown"`
	SSLDownDaysBefore        int64    `json:"ssl_down_days_before"`
	Url                      string   `json:"url"`
	Tags                     []string `json:"tags"`
	UserIds                  string   `json:"userids"`
	VerifyCertificate        bool     `json:"verify_certificate"`
}

func (client *client) CreateCheck(ctx context.Context, body CreateCheckRequest) (*int64, error) {
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	uri, err := url.JoinPath(client.baseURL, "checks")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(encodedBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	var res *struct {
		Check struct {
			Id int64 `json:"id"`
		} `json:"check"`
	}
	err = client.do(req, &res)
	if err != nil {
		return nil, err
	}

	return &res.Check.Id, nil
}

func (client *client) UpdateCheck(ctx context.Context, checkId string, body CreateCheckRequest) error {
	body.Type = ""
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	uri, err := url.JoinPath(client.baseURL, "checks", checkId)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, bytes.NewReader(encodedBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	var res *struct{}
	return client.do(req, &res)
}

func (client *client) DeleteCheck(ctx context.Context, checkId string) error {
	uri, err := url.JoinPath(client.baseURL, "checks", checkId)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, http.NoBody)
	if err != nil {
		return err
	}

	var res *struct{}
	return client.do(req, &res)
}
