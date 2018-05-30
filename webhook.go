package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const WebhooksURL = "https://api.ciscospark.com/v1/webhooks"

type Webhook struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	TargetURL string                 `json:"targetUrl"`
	Resource  string                 `json:"resource"`
	Event     string                 `json:"event"`
	Filter    string                 `json:"filter,omitempty"`
	Secret    string                 `json:"secret,omitempty"`
	OrgID     string                 `json:"orgId,omitempty"`
	CreatedBy string                 `json:"createdBy,omitempty"`
	AppID     string                 `json:"appId,omitempty"`
	OwnedBy   string                 `json:"ownedBy,omitempty"`
	Status    string                 `json:"active,omitempty"`
	ActorID   string                 `json:"actorId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"` // TODO: what is this?  Is it needed? Not in the docs
}

type WebhookList struct {
	Items []*Webhook
}

type NewWebhook struct {
	Name      string `json:"name"`             // required
	TargetURL string `json:"targetUrl"`        // required
	Resource  string `json:"resource"`         // required
	Event     string `json:"event"`            // required
	Filter    string `json:"filter,omitempty"` // optional
	Secret    string `json:"secret,omitempty"` // optional
}

// https://developer.webex.com/endpoint-webhooks-webhookId-get.html
func (c *client) GetWebhook(webhookID string) (*Webhook, error) {
	if webhookID == "" {
		return nil, fmt.Errorf("no webhook ID specified")
	}

	resp, err := c.getRequest(fmt.Sprintf("%s/%s", WebhooksURL, webhookID), nil)
	if err != nil {
		return nil, err
	}

	var webhook Webhook
	if err := json.Unmarshal(resp, &webhook); err != nil {
		return nil, err
	}
	return &webhook, err
}

// https://developer.webex.com/endpoint-webhooks-post.html
func (c *client) CreateWebhook(w *NewWebhook) (*Webhook, error) {
	if w == nil {
		return nil, fmt.Errorf("nil webhook")
	}
	if w.Name == "" {
		return nil, fmt.Errorf("no webhook name specified")
	}
	if w.TargetURL == "" {
		return nil, fmt.Errorf("no webhook target URL specified")
	}
	if w.Resource == "" {
		return nil, fmt.Errorf("no webhook resource specified")
	}
	if w.Event == "" {
		return nil, fmt.Errorf("no webhook event specified")
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(w); err != nil {
		return nil, err
	}
	resp, err := c.postRequest(WebhooksURL, b)

	if err != nil {
		return nil, err
	}

	var rwh Webhook
	err = json.Unmarshal(resp, &rwh)
	return &rwh, err
}

// https://developer.webex.com/endpoint-webhooks-webhookId-put.html
func (c *client) UpdateWebhook(w *Webhook) (*Webhook, error) {
	if w == nil {
		return nil, fmt.Errorf("nil webhook")
	}
	if w.ID == "" {
		return nil, fmt.Errorf("no webhook ID specified")
	}
	if w.Name == "" {
		return nil, fmt.Errorf("no webhook name specified")
	}
	if w.TargetURL == "" {
		return nil, fmt.Errorf("no webhook target URL specified")
	}
	// weirdly, Resource and Event aren't required, despite the fact that they are required for *new* webhooks

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(w); err != nil {
		return nil, err
	}
	resp, err := c.putRequest(fmt.Sprintf("%s/%s", WebhooksURL, w.ID), b)
	if err != nil {
		return nil, err
	}

	var rwh Webhook
	err = json.Unmarshal(resp, &rwh)
	return &rwh, err
}

// https://developer.webex.com/endpoint-webhooks-webhookId-delete.html
func (c *client) DeleteWebhook(hookID string) error {
	if hookID == "" {
		return fmt.Errorf("no webhook ID specified")
	}

	_, err := c.deleteRequest(fmt.Sprintf("%s/%s", WebhooksURL, hookID))
	return err
}

// https://developer.webex.com/endpoint-webhooks-get.html
func (c *client) ListWebhooks(max int) ([]*Webhook, error) {
	resp, reqErr := c.getRequestWithPaging(WebhooksURL, nil, max)
	if reqErr != nil && len(resp) == 0 { // if we got an error *and* results, parse them and return them
		return nil, reqErr
	}

	var webhooks []*Webhook
	for _, r := range resp {
		var w WebhookList
		if jsonErr := json.Unmarshal(r, &w); jsonErr != nil {
			return webhooks, fmt.Errorf("%v && %v", reqErr, jsonErr)
		}
		webhooks = append(webhooks, w.Items...)
	}
	return webhooks, reqErr
}
