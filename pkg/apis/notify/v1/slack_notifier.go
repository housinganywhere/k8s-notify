package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

var client httpClient = &http.Client{Timeout: 2*time.Second}

// SlackNotifier defines the spec for integrating with Slack
// +k8s:openapi-gen=true
type SlackNotifier struct {
	WebhookUrl string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
}

type slackSendMessageRequest struct {
	Channel   string `json:"channel,omitempty"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Text      string `json:"text"`
}

func (sn *SlackNotifier) Send(message string) error {
	msg := slackSendMessageRequest{
		Channel:   sn.Channel,
		Username:  sn.Username,
		IconEmoji: sn.IconEmoji,
		Text:      escapeString(message),
	}

	b, err := json.Marshal(&msg)
	if err != nil {
		return errors.Wrap(err, "failed to serialize the request data")
	}
	req, err := http.NewRequest("POST", sn.WebhookUrl, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "failed to create the request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make a request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
}
