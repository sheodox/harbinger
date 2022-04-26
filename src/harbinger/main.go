package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sheodox/harbinger/config"
	"github.com/sheodox/harbinger/discord"
	"github.com/sheodox/harbinger/health"
	"github.com/sheodox/harbinger/server"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dis := discord.NewDiscord(cfg)

	checker := health.NewChecker(cfg.Services, dis)

	quit := make(chan any)
	startChecker(checker, quit)

	dis.Harbinger.Send(fmt.Sprintf("Harbinger %v started", cfg.Harbinger.Name))

	fmt.Println("Harbinger started")

	server.StartServer(cfg, dis)

	<-quit
}

func startChecker(checker health.Checker, quit chan any) {
	checker.Check()
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				checker.Check()
			}
		}
	}()
}
