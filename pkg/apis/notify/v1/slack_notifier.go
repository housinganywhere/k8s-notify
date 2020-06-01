package v1

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client httpClient = &http.Client{Timeout: 2 * time.Second, Transport: tr}

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

func (sn *SlackNotifier) SendRich(e corev1.Event) error {
	var msg = sn.richFormat(e)
	return sn.Send(msg)
}

func (sn *SlackNotifier) Send(message string) error {
	// msg := slackSendMessageRequest{
	// 	Channel:   sn.Channel,
	// 	Username:  sn.Username,
	// 	IconEmoji: sn.IconEmoji,
	// 	Text:      message,
	// }

	// b, err := json.Marshal(&message)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to serialize the request data")
	// }
	// req, err := http.NewRequest("POST", sn.WebhookUrl, bytes.NewBuffer(b))
	req, err := http.NewRequest("POST", sn.WebhookUrl, strings.NewReader(message))
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

func (sn *SlackNotifier) richFormat(e corev1.Event) string {
	const rich = `
{
		"channel": "{{ .Channel }}",
		"blocks": [
		{
			"type": "divider"
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "{{ .Type }} {{ .Msg }}"
			}
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "mrkdwn",
					"text": ":kibana: <https://logs.housinganywhere.com/app/kibana#/discover/8ecb25e0-a26b-11ea-988b-b7fac5cb3e38?_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:now-3d,to:now))&_a=(columns:!(involvedObject.kind,message),filters:!(),index:'50fa3090-94fa-11ea-988b-b7fac5cb3e38',interval:auto,query:(language:kuery,query:%22{{ .Uid }}%22),sort:!(!(firstTimestamp,desc)))|uid>"
				},
				{
					"type": "mrkdwn",
					"text": "<https://logs.housinganywhere.com/app/kibana#/discover/8ecb25e0-a26b-11ea-988b-b7fac5cb3e38?_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:now-3d,to:now))&_a=(columns:!(involvedObject.kind,message),filters:!(),index:'50fa3090-94fa-11ea-988b-b7fac5cb3e38',interval:auto,query:(language:kuery,query:%22{{ .Name }}%22),sort:!(!(firstTimestamp,desc)))|{{ .Name }}>"
				},
				{
					"type": "mrkdwn",
					"text": "{{ .Extra }}"
				}
			]
		}
	]
}
	`

	type slackEvent struct {
		Msg, Type, Uid, Extra, Name, Channel string
	}

	var se slackEvent
	se.Msg = e.Message
	if e.Type == "Normal" {
		se.Type = ":white_check_mark:"
	} else {
		se.Type = ":warning:"
	}
	se.Uid = string(e.InvolvedObject.UID)
	se.Extra = fmt.Sprintf("%s %s %s", e.InvolvedObject.Kind, e.Source.Component, e.Reason)
	se.Name = e.InvolvedObject.Name
	se.Channel = sn.Channel

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("rich").Parse(rich))
	err := t.Execute(buf, se)
	if err != nil {
		log.Println("error executing template:", err)
	}
	return buf.String()
}
