package api

import (
	"context"
	api_types "github.com/scayle/terraform-provider-pingdom/internal/api/types"
	"net/http"
	"net/url"
)

func (client *client) GetContacts(ctx context.Context) (*api_types.Contacts, error) {
	uri, err := url.JoinPath(client.baseURL, "alerting/contacts")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	var res *api_types.Contacts
	err = client.do(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
