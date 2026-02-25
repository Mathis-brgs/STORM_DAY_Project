# Message Service - Commandes locales
.PHONY: run run-postgres build proto proto-new migrate seed test clean

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

# Migrations (depuis la racine: make migrate-message)
# 001 = tables, 003 = messages (id int, sender_id uuid, group_id int), 004 = groups.user_id uuid
migrate:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/001_create_tables.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/003_messages_uuid.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/004_groups_user_id_uuid.sql

# Seed (depuis la racine: make seed-message)
seed:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/002_seed_data.sql

# Tests
test:
	go test ./...

# Nettoie les binaires et artefacts protoc
clean:
	rm -f main
	rm -rf github.com
