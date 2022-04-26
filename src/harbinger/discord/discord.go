package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sheodox/harbinger/config"
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

type Discord struct {
	Harbinger Webhook
	Services  map[config.Service]Webhook
}

func NewDiscord(cfg config.Config) Discord {
	services := make(map[config.Service]Webhook)

	for _, service := range cfg.Services {
		services[service] = Webhook{service.Webhook}
	}

	return Discord{
		Harbinger: Webhook{cfg.Harbinger.Webhook},
		Services:  services,
	}
}

func (d Discord) SendAsServiceByName(serviceName, msg string) {
	for service, webhook := range d.Services {
		if service.ServiceName == serviceName {
			webhook.Send(msg)
			return
		}
	}

	d.Harbinger.Send(fmt.Sprintf("[%v] %v", serviceName, msg))
}

func (d Discord) SendAsService(service config.Service, msg string) {
	for s, webhook := range d.Services {
		if s == service {
			webhook.Send(msg)
			return
		}
	}
}
