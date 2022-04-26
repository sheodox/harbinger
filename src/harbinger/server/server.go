package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sheodox/harbinger/config"
	"github.com/sheodox/harbinger/discord"
)

type LogMessage struct {
	Service   string `json:"service"`
	Concern   string `json:"concern"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func StartServer(cfg config.Config, dis discord.Discord) {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(200, "")
	})

	e.POST("/logs", func(c echo.Context) error {
		logs := make([]LogMessage, 0)

		if err := c.Bind(&logs); err != nil {
			return err
		}

		for _, log := range logs {
			dis.SendAsServiceByName(log.Service, fmt.Sprintf("`%v: %v`", log.Concern, log.Message))
		}

		return c.String(200, "")
	})

	e.Start(":3000")
}
