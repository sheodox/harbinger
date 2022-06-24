package health

import (
	"testing"
	"time"

	"github.com/sheodox/harbinger/config"
)

type DiscordMock struct {
	sentAlert bool
}

func (d *DiscordMock) SendAsService(service config.Service, messages string) {
	d.sentAlert = true
}

func TestHealthCheck(t *testing.T) {
	mockOnline := true
	mockStatus := 200
	var mockErr error
	checkStatus := func(_ config.Service) (bool, int, error) {
		return mockOnline, mockStatus, mockErr
	}

	fakeServices := make([]*ServiceStatus, 0)
	fakeService := config.Service{DisplayName: "test", ServiceName: "test", Endpoint: "test", Webhooks: []string{"test"}}
	fakeServices = append(fakeServices, &ServiceStatus{Service: fakeService, Online: true, Alerted: true, ConsecutiveOfflineChecks: 0, OfflineAt: time.Now()})
	d := DiscordMock{false}
	health := Checker{fakeServices, &d, checkStatus}

	health.Check()

	if d.sentAlert {
		t.Fatal("online services shouldn't send messages")
	}

	mockOnline = false

	health.Check()
	if d.sentAlert {
		t.Fatal("shouldn't send message for first offline, in case it's a network blip")
	}

	health.Check()

	if !d.sentAlert {
		t.Fatal("should have sent message after confirmed offline by two consecutive failures")
	}

	d.sentAlert = false

	health.Check()

	mockOnline = true
	health.Check()

	if !d.sentAlert {
		t.Fatal("back online, should have sent a message")
	}

	d.sentAlert = false
	mockOnline = false
	health.Check()
	mockOnline = true
	health.Check()

	if d.sentAlert {
		t.Fatal("back online after only one failed check should not send a message")
	}

	mockOnline = false
	health.Check()

	if d.sentAlert {
		t.Fatal("shouldn't send alert after just one")
	}

	health.Check()
	if !d.sentAlert {
		t.Fatal("after having recovered previously, it should alert again after another two consecutive failures")
	}
}
