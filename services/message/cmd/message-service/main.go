package main

import (
	"log"
	"os"
	"strings"

	natsh "github.com/Mathis-brgs/storm-project/services/message/internal/nats"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/memory"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/postgres"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/nats-io/nats.go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("message-service starting...")

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url, nats.Name("message-service"))
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Drain()

	var messageRepo repo.MessageRepo
	if strings.ToLower(os.Getenv("STORAGE")) == "postgres" {
		db, err := postgres.NewDB()
		if err != nil {
			log.Fatalf("postgres connect: %v", err)
		}
		defer db.Close()
		messageRepo = postgres.NewMessageRepo(db)
		log.Println("storage: postgres")
	} else {
		messageRepo = memory.NewMessageRepo()
		log.Println("storage: memory")
	}

	svc := service.NewMessageService(messageRepo)
	handler := natsh.NewMessageHandler(svc)

	if err := handler.Listen(nc); err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Println("ready, listening on NEW_MESSAGE")
	select {}
}
