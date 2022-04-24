package config

import (
	"encoding/json"
	"io/ioutil"
)

type Service struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Webhook  string `json:"webhook"`
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

	return config, err
}
