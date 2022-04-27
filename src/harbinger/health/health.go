package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sheodox/harbinger/config"
	"github.com/sheodox/harbinger/discord"
)

const (
	// need to get at least this many sequential offline statuses in a
	// row before considering the service down otherwise we'll cry wolf at a network blip
	offlineCountThreshold = 2
)

type ServiceStatus struct {
	Service                  config.Service
	Online                   bool
	ConsecutiveOfflineChecks int
	OfflineAt                time.Time
}

type Checker struct {
	Services []*ServiceStatus
	Discord  discord.Discord
}

func NewChecker(services []config.Service, d discord.Discord) Checker {
	serviceStatuses := make([]*ServiceStatus, len(services))

	for i, service := range services {
		// assume online at first
		serviceStatuses[i] = &ServiceStatus{service, true, 0, time.Now()}
	}

	return Checker{serviceStatuses, d}
}

func (c *Checker) Check() {
	for _, service := range c.Services {
		// ping
		online, statusCode, err := c.checkServiceStatus(service.Service)

		if !online {
			// keep track of how many times in a row we've checked and saw the service was offline
			service.ConsecutiveOfflineChecks++
		}

		if !online && service.ConsecutiveOfflineChecks == 1 {
			// we won't report at the first sign a service is offline, but we want
			// to know of the first time we noticed it. if the service is offline and
			// we go on to alert about it, this lets us show a more accurate downtime
			service.OfflineAt = time.Now()
		} else if !online && service.ConsecutiveOfflineChecks == offlineCountThreshold {
			// the service has gone offline since we last checked
			if err != nil {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)\n`Error: %v`", service.Service.DisplayName, statusCode, err))
			} else {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)", service.Service.DisplayName, statusCode))
			}
		} else if online && !service.Online {
			// the service has recovered
			downtime := time.Now().Sub(service.OfflineAt).Round(time.Second)
			c.Discord.SendAsService(service.Service, fmt.Sprintf(":green_circle: %v is back online (down %v)", service.Service.DisplayName, downtime))
			service.ConsecutiveOfflineChecks = 0
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
