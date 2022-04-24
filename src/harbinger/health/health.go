package health

import (
	"fmt"
	"net/http"

	"github.com/sheodox/harbinger/config"
	"github.com/sheodox/harbinger/discord"
)

type ServiceStatus struct {
	Service config.Service
	Online  bool
	discord discord.Webhook
}

type Checker struct {
	Services []*ServiceStatus
}

func NewChecker(services []config.Service, d discord.Webhook) Checker {
	serviceStatuses := make([]*ServiceStatus, len(services))

	for i, service := range services {
		// assume online at first
		serviceStatuses[i] = &ServiceStatus{service, true, discord.NewWebhook(service.Webhook)}
	}

	return Checker{serviceStatuses}
}

func (c *Checker) Check() {
	for _, service := range c.Services {
		// ping
		online, statusCode, err := c.checkServiceStatus(service.Service)

		if !online && service.Online {
			// the service has gone offline since we last checked
			if err != nil {
				service.discord.Send(fmt.Sprintf("%v has gone offline (%v)\nError: %v", service.Service.Name, statusCode, err))
			} else {
				service.discord.Send(fmt.Sprintf("%v has gone offline (%v)", service.Service.Name, statusCode))
			}
		} else if online && !service.Online {
			// the service has recovered
			service.discord.Send(fmt.Sprintf("%v is back online", service.Service.Name))
		}

		service.Online = online
	}
}

func (c Checker) checkServiceStatus(service config.Service) (bool, int, error) {
	resp, err := http.Get(service.Endpoint)

	if err != nil {
		return false, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, resp.StatusCode, nil
	}

	return true, resp.StatusCode, nil
}
