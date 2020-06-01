package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

type HangoutsChatNotifier struct {
	WebhookUrl string `json:"webhook_url"`
}

func (n *HangoutsChatNotifier) Send(message string) error {
	url := n.WebhookUrl

	var jsonStr = []byte(fmt.Sprintf(`{"text":"%s"}`, escapeString(message)))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json.RawMessage(jsonStr)))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil

}

func (n *HangoutsChatNotifier) SendRich(e corev1.Event) error {
	return n.Send(e.Message)
}
