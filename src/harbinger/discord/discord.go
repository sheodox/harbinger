package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Webhook struct {
	url string
}

func NewWebhook(url string) Webhook {
	return Webhook{url}
}

type WebhookMessage struct {
	Content string `json:"content"`
}

func (w Webhook) Send(msg string) error {
	message := WebhookMessage{msg}
	fmt.Printf("Discord: %v\n", msg)

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = http.Post(w.url, "application/json", bytes.NewBuffer(payload))

	return err
}
