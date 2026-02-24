package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mathis-brgs/storm-project/services/notification/internal/service"
	"github.com/Mathis-brgs/storm-project/services/notification/internal/subscribers"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("connexion NATS: %v", err)
	}
	defer nc.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	notifService := service.NewNotificationService(rdb)

	if err := subscribers.StartNotificationSubscribers(nc, notifService); err != nil {
		log.Fatalf("démarrage subscribers: %v", err)
	}

	fmt.Printf("Notification service démarré — NATS: %s | Redis: %s\n", natsURL, redisAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Notification service arrêté")
}