package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sheodox/harbinger/config"
	"github.com/sheodox/harbinger/discord"
)

type ServiceStatus struct {
	Service   config.Service
	Online    bool
	OfflineAt time.Time
}

type Checker struct {
	Services []*ServiceStatus
	Discord  discord.Discord
}

func NewChecker(services []config.Service, d discord.Discord) Checker {
	serviceStatuses := make([]*ServiceStatus, len(services))

	for i, service := range services {
		// assume online at first
		serviceStatuses[i] = &ServiceStatus{service, true, time.Now()}
	}

	return Checker{serviceStatuses, d}
}

func (c *Checker) Check() {
	for _, service := range c.Services {
		// ping
		online, statusCode, err := c.checkServiceStatus(service.Service)

		if !online && service.Online {
			// the service has gone offline since we last checked
			if err != nil {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)\n`Error: %v`", service.Service.DisplayName, statusCode, err))
			} else {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)", service.Service.DisplayName, statusCode))
			}

			service.OfflineAt = time.Now()
		} else if online && !service.Online {
			// the service has recovered
			downtime := time.Now().Sub(service.OfflineAt).Round(time.Second)
			c.Discord.SendAsService(service.Service, fmt.Sprintf(":green_circle: %v is back online (down %v)", service.Service.DisplayName, downtime))
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
