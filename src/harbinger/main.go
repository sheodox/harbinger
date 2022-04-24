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

	harbingerWebhook := discord.NewWebhook(cfg.Harbinger.Webhook)

	checker := health.NewChecker(cfg.Services, harbingerWebhook)

	quit := make(chan any)
	startChecker(checker, quit)

	harbingerWebhook.Send(fmt.Sprintf("Harbinger %v started", cfg.Harbinger.Name))

	fmt.Println("Harbinger started")

	server.StartServer(harbingerWebhook)

	<-quit
}

func startChecker(checker health.Checker, quit chan any) {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				checker.Check()
			}
		}
	}()
}
