# Message Service - Commandes locales
.PHONY: run run-postgres build proto proto-new migrate migrate-legacy seed migrate-docker migrate-legacy-docker seed-docker test clean

NAMESPACE=storm
MESSAGE_DB_NAME=message_db
POSTGRES_USER=storm

# Lance le service avec stockage mémoire (défaut)
run:
	go run ./cmd/main.go

# Lance le service avec PostgreSQL (nécessite docker compose postgres-chat + nats)
run-postgres:
	STORAGE=postgres go run ./cmd/main.go

# Compile le binaire
build:
	go build -o message-service ./cmd/main.go

# Régénère message.pb.go avec l'ancien générateur (Docker znly/protoc). Génère du code utilisant github.com/golang/protobuf (déprécié).
proto:
	docker run --rm -v $$(pwd):/workspace -w /workspace znly/protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto
	@if [ -d github.com ]; then \
		cp github.com/Mathis-brgs/storm-project/services/message/api/v1/message.pb.go api/v1/message.pb.go; \
		rm -rf github.com; \
	fi

# Régénère message.pb.go avec le générateur actuel (google.golang.org/protobuf). Prérequis: protoc + protoc-gen-go.
#   brew install protobuf
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# PATH inclut $(go env GOPATH)/bin pour que protoc trouve protoc-gen-go.
proto-new:
	PATH="$$(go env GOPATH)/bin:$$PATH" protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto

# Migrations standard (création schéma cible) - K8s
migrate:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Déploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < migrations/001_create_tables.sql

# Migration legacy (optionnelle) pour nettoyer/convertir un ancien schéma - K8s.
migrate-legacy:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < migrations/005_conversations_refactor.sql

# Seed - K8s
seed:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < migrations/002_seed_data.sql

# Fallback Docker Compose
migrate-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/001_create_tables.sql

migrate-legacy-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/005_conversations_refactor.sql

seed-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/002_seed_data.sql

# Tests
test:
	go test ./...

# Nettoie les binaires et artefacts protoc
clean:
	rm -f main
	rm -rf github.com
