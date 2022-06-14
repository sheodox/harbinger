package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sheodox/harbinger/config"
)

const (
	MSG_REQUEUE_TIMEOUT = time.Minute
)

type Webhook struct {
	url  string
	msgs chan string
}

func NewWebhook(url string) Webhook {
	w := Webhook{url, make(chan string, 10)}

	go func() {
		for {
			select {
			case msg := <-w.msgs:
				fmt.Printf("Discord: %v\n", msg)
				message := WebhookMessage{msg}

				payload, err := json.Marshal(message)
				if err != nil {
					fmt.Printf("Error marshalling json for Discord message payload. msg: %q\n%v\n", msg, err)
				}

				resp, err := http.Post(w.url, "application/json", bytes.NewBuffer(payload))

				if err != nil {
					fmt.Println("Error encountered trying to send message")
					w.requeue(msg)
				} else if resp.StatusCode > 299 {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						fmt.Println("Error reading body: ", err)
					}
					resp.Body.Close()

					fmt.Printf("Error sending discord message: %v\n%v\n", resp.Status, string(body))

					w.requeue(msg)
				} else {
					fmt.Println("Successfully sent message")
				}
			}
		}
	}()

	return w
}

func (w Webhook) requeue(msg string) {
	fmt.Printf("Requeueing message %q\n", msg)
	go func() {
		<-time.After(MSG_REQUEUE_TIMEOUT)
		w.Send(msg)
	}()
}

type WebhookMessage struct {
	Content string `json:"content"`
}

func (w Webhook) Send(msg string) {
	w.msgs <- msg
}

type Discord struct {
	Harbinger Webhook
	Services  map[config.Service]Webhook
}

func NewDiscord(cfg config.Config) Discord {
	services := make(map[config.Service]Webhook)

	for _, service := range cfg.Services {
		services[service] = NewWebhook(service.Webhook)
	}

	return Discord{
		Harbinger: NewWebhook(cfg.Harbinger.Webhook),
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
