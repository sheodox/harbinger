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
	urls []string
	msgs chan string
}

func NewWebhook(urls []string) Webhook {
	w := Webhook{urls, make(chan string, 10)}

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

				for _, url := range w.urls {
					resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))

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
	Services  map[string]Webhook
}

func NewDiscord(cfg config.Config) Discord {
	services := make(map[string]Webhook)

	for _, service := range cfg.Services {
		services[service.ServiceName] = NewWebhook(service.Webhooks)
	}

	return Discord{
		Harbinger: NewWebhook([]string{cfg.Harbinger.Webhook}),
		Services:  services,
	}
}

func (d Discord) SendAsServiceByName(serviceName, msg string) {
	for service, webhook := range d.Services {
		if service == serviceName {
			webhook.Send(msg)
			return
		}
	}

	d.Harbinger.Send(fmt.Sprintf("[%v] %v", serviceName, msg))
}

func (d Discord) SendAsService(service config.Service, msg string) {
	d.SendAsServiceByName(service.ServiceName, msg)
}
