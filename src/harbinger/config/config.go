package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type ServiceReminder struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type Service struct {
	DisplayName  string          `json:"displayName"`
	ServiceName  string          `json:"serviceName"`
	Endpoint     string          `json:"endpoint"`
	Webhooks     []string        `json:"webhooks"`
	Reminder     ServiceReminder `json:"reminder"`
	ReminderTime time.Duration
}

type Harbinger struct {
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
}

type Config struct {
	Services  []Service `json:"services"`
	Harbinger Harbinger `json:"harbinger"`
}

var (
	config       Config
	configLoaded = false
)

func LoadConfig() (Config, error) {
	if configLoaded {
		return config, nil
	}

	config = Config{}

	data, err := ioutil.ReadFile("./config.json")

	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)

	if err == nil {
		configLoaded = true
	}

	validateConfig(config)

	for index, service := range config.Services {
		service.ReminderTime = time.Hour*time.Duration(service.Reminder.Hours) + time.Minute*time.Duration(service.Reminder.Minutes) + time.Second*time.Duration(service.Reminder.Seconds)

		// need a minimum time
		if service.ReminderTime <= 0 {
			service.ReminderTime = time.Hour * 8
		}

		config.Services[index] = service
	}

	return config, err
}

func validateConfig(cfg Config) {
	validatePropertyUnique := func(propertyGetter func(Service) string, servicePropName string) {
		unique := make(map[string]bool)
		for _, service := range cfg.Services {
			prop := propertyGetter(service)
			if prop != "" {
				unique[prop] = true
			}
		}

		if len(unique) < len(cfg.Services) {
			log.Fatalf("Service config invalid! Each service must have a unique %q", servicePropName)
		}
	}

	validatePropertyUnique(func(s Service) string {
		return s.ServiceName
	}, "serviceName")

	validatePropertyUnique(func(s Service) string {
		return s.DisplayName
	}, "displayName")

	validatePropertyUnique(func(s Service) string {
		return s.Endpoint
	}, "endpoint")

	validatePropertyUnique(func(s Service) string {
		return strings.Join(s.Webhooks, ",")
	}, "webhooks")
}
