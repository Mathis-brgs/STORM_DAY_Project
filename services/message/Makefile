# Message Service - Commandes locales
.PHONY: run run-postgres build proto migrate seed test clean

# Lance le service avec stockage mémoire (défaut)
run:
	go run ./cmd/message-service/

# Lance le service avec PostgreSQL (nécessite docker compose postgres-chat + nats)
run-postgres:
	STORAGE=postgres go run ./cmd/message-service/

# Compile le binaire
build:
	go build -o message-service ./cmd/message-service/

# Régénère message.pb.go (nécessite Docker + znly/protoc)
proto:
	docker run --rm -v $$(pwd):/workspace -w /workspace znly/protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto
	@if [ -d github.com ]; then \
		cp github.com/Mathis-brgs/storm-project/services/message/api/v1/message.pb.go api/v1/message.pb.go; \
		rm -rf github.com; \
	fi

# Migrations (depuis la racine du projet: make migrate-message)
# Nécessite: docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/001_create_tables.sql
migrate:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/001_create_tables.sql

# Seed (depuis la racine: make seed-message)
seed:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < migrations/002_seed_data.sql

# Tests
test:
	go test ./...

# Nettoie les binaires et artefacts protoc
clean:
	rm -f message-service
	rm -rf github.com
