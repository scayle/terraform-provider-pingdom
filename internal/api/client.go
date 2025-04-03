package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api_types "github.com/scayle/terraform-provider-pingdom/internal/api/types"
	"io"
	"net/http"
)

type Client interface {
	GetCheck(ctx context.Context, id string) (*api_types.Check, error)
	CreateCheck(ctx context.Context, body CreateCheckRequest) (*int64, error)
	UpdateCheck(ctx context.Context, id string, body CreateCheckRequest) error
	DeleteCheck(ctx context.Context, id string) error

	GetContacts(ctx context.Context) (*api_types.Contacts, error)
}

type client struct {
	baseURL string
	token   string
}

func New(token string) Client {
	return &client{
		token:   token,
		baseURL: "https://api.pingdom.com/api/3.1",
	}
}

func (client *client) do(req *http.Request, r any) error {
	req.Header.Set("Authorization", "Bearer "+client.token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	tflog.Debug(req.Context(), "Received Response", map[string]interface{}{
		"body":   string(body),
		"status": res.StatusCode,
	})

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if err := json.Unmarshal(body, r); err != nil {
		return err
	}

	return nil
}
