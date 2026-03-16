package main

import (
	"log"
	"net/http"
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
	log.Println("message starting...")

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url, nats.Name("message"))
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Drain()

	var messageRepo repo.MessageRepo
	var conversationRepo repo.ConversationRepo
	if strings.ToLower(os.Getenv("STORAGE")) == "postgres" {
		db, err := postgres.NewDB()
		if err != nil {
			log.Fatalf("postgres connect: %v", err)
		}
		defer db.Close()
		messageRepo = postgres.NewMessageRepo(db)
		conversationRepo = postgres.NewConversationRepo(db)
		log.Println("storage: postgres")
	} else {
		messageRepo = memory.NewMessageRepo()
		conversationRepo = memory.NewConversationRepo()
		log.Println("storage: memory")
	}

	messageSvc := service.NewMessageService(messageRepo)
	conversationSvc := service.NewConversationService(conversationRepo)
	handler := natsh.NewMessageHandler(messageSvc, conversationSvc)

	if err := handler.Listen(nc); err != nil {
		log.Fatalf("listen: %v", err)
	}

	startHTTPHealthServer()

	log.Println("ready, listening on NATS")
	select {}
}

func startHTTPHealthServer() {
	httpPort := os.Getenv("HTTP_PORT")
	if strings.TrimSpace(httpPort) == "" {
		httpPort = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	addr := ":" + httpPort
	go func() {
		log.Printf("health server listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("health server listen: %v", err)
		}
	}()
}
