package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sheodox/harbinger/config"
)

const (
	// need to get at least this many sequential offline statuses in a
	// row before considering the service down otherwise we'll cry wolf at a network blip
	OFFLINE_COUNT_THRESHOLD = 2
	HEALTH_CHECK_TIMEOUT    = time.Second * 10
)

type ServiceStatus struct {
	Service                  config.Service
	Online                   bool
	Alerted                  bool
	ConsecutiveOfflineChecks int
	OfflineAt                time.Time
}

type Notifier interface {
	SendAsService(config.Service, string)
}

type ServiceStatusChecker func(config.Service) (bool, int, error)

type Checker struct {
	Services           []*ServiceStatus
	Discord            Notifier
	checkServiceStatus ServiceStatusChecker
}

func NewChecker(services []config.Service, d Notifier) Checker {
	serviceStatuses := make([]*ServiceStatus, len(services))

	for i, service := range services {
		// assume online at first
		serviceStatuses[i] = &ServiceStatus{Service: service, Online: true, Alerted: false, ConsecutiveOfflineChecks: 0, OfflineAt: time.Now()}
	}

	return Checker{serviceStatuses, d, checkServiceStatus}
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
			// we won't alert at the first sign a service is offline, but we want
			// to know when we first noticed it. if the service is offline and
			// we go on to alert about it, this lets us show a more accurate downtime
			service.OfflineAt = time.Now()
		} else if !online && service.ConsecutiveOfflineChecks == OFFLINE_COUNT_THRESHOLD {
			// the service has gone offline since we last checked
			if err != nil {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)\n`Error: %v`", service.Service.DisplayName, statusCode, err))
			} else {
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":red_circle: %v has gone offline (%v)", service.Service.DisplayName, statusCode))
			}

			service.Alerted = true
		} else if online && !service.Online {
			// only log if we've actually notified of the service being down
			if service.Alerted {
				// the service has recovered
				downtime := time.Now().Sub(service.OfflineAt).Round(time.Second)
				c.Discord.SendAsService(service.Service, fmt.Sprintf(":green_circle: %v is back online (down %v)", service.Service.DisplayName, downtime))
			}

			service.Alerted = false
			service.ConsecutiveOfflineChecks = 0
		}

		service.Online = online
	}
}

var (
	healthCheckerClient = http.Client{
		Timeout: HEALTH_CHECK_TIMEOUT,
	}
)

func checkServiceStatus(service config.Service) (bool, int, error) {
	resp, err := healthCheckerClient.Get(service.Endpoint)

	if err != nil {
		return false, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, resp.StatusCode, nil
	}

	return true, resp.StatusCode, nil
}
