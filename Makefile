# Variables
CLUSTER_NAME=storm
NAMESPACE=storm
GATEWAY_PORT=30080
IMAGES=storm/user-service:latest storm/gateway-service:latest storm/message-service:latest storm/media-service:latest
POSTGRES_USER=storm
MESSAGE_DB_NAME=storm_message_db
USER_DB_NAME=storm_user_db

.PHONY: up down clean build deploy import restart status logs logs-media \
	migrate-message migrate-message-legacy migrate-message-006 seed-message seed-user \
	migrate-message-docker migrate-message-legacy-docker migrate-message-006-docker seed-message-docker seed-user-docker \
	dev-infra-up dev-migrate-all-docker dev-setup-docker k8s-reset-postgres-message \
	proto-message

# Lance tout : cluster, build, import et déploiement
up:
	k3d cluster create $(CLUSTER_NAME) -p "$(GATEWAY_PORT):$(GATEWAY_PORT)@server:0" --wait || true
	$(MAKE) build
	$(MAKE) import
	$(MAKE) deploy

# Arrête le cluster
down:
	k3d cluster stop $(CLUSTER_NAME)

# Supprime le cluster
clean:
	k3d cluster delete $(CLUSTER_NAME)

# Build toutes les images Docker (même cible que la CI : une seule commande local = CI)
# Note: On utilise :latest pour coller à la variable IMAGES
build:
	docker build -t storm/user-service:latest services/user/
	docker build -f services/gateway/Dockerfile -t storm/gateway-service:latest .
	docker build -t storm/message-service:latest services/message/
	docker build -t storm/media-service:latest services/media/

# Importe les images dans k3d
import:
	k3d image import $(IMAGES) -c $(CLUSTER_NAME)

# Applique les manifestes K8s
deploy:
	kubectl apply -k infra/k8s/base/

# Redémarre les services applicatifs
restart:
	kubectl rollout restart deployment -n $(NAMESPACE)

# État des pods
status:
	@echo "--- État des Pods dans $(NAMESPACE) ---"
	kubectl get pods -n $(NAMESPACE)

# Logs de tous les services (si le label part-of existe)
logs:
	kubectl logs -n $(NAMESPACE) -l "app.kubernetes.io/part-of=storm" --tail=100 -f

# LOGS SPÉCIFIQUES MEDIA (Ta zone de test)
logs-media:
	kubectl logs -n $(NAMESPACE) -f -l app=media-service

# --- Migrations & Seeds (K8s par defaut) ---

# Crée la DB message si elle n'existe pas (volume existant créé avec un ancien nom)
create-message-db:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable."; exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d postgres -c "CREATE DATABASE $(MESSAGE_DB_NAME);" 2>/dev/null || true

# Migrations DB Message (schéma cible)
migrate-message:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < services/message/migrations/001_create_tables.sql

# Migration DB Message legacy (optionnelle)
migrate-message-legacy:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < services/message/migrations/005_conversations_refactor.sql

# Migration 006: reply_to_id, status, forward_from_id, message_seen_by
migrate-message-006:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < services/message/migrations/006_message_reply_status_forward_seen.sql

# Seed DB Message (conversations + messages)
seed-message:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < services/message/migrations/002_seed_data.sql

# Crée les tables user (users, jwt) dans storm_user_db
create-user-tables:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-user -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then echo "Pod postgres-user introuvable."; exit 1; fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(USER_DB_NAME) < infra/seed/000_create_user_tables.sql

# Seed DB User (nécessite que le user-service ait créé les tables)
seed-user:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-user -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-user introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(USER_DB_NAME) < infra/seed/001_seed_users.sql

# --- Fallback Docker Compose ---

migrate-message-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/001_create_tables.sql

migrate-message-legacy-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/005_conversations_refactor.sql

migrate-message-006-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/006_message_reply_status_forward_seen.sql

seed-message-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/002_seed_data.sql

seed-user-docker:
	docker exec -i storm-postgres-user psql -U storm -d storm_user_db < infra/seed/001_seed_users.sql

# --- Dev local (Docker Compose) : idéal pour tester le front sans k8s ---

# Infra minimale : Postgres user + message, NATS, Redis
dev-infra-up:
	docker compose up -d postgres-user postgres-chat nats redis

# Applique toutes les migrations + seed user (conteneurs déjà démarrés)
dev-migrate-all-docker:
	@echo "→ Migrations message DB (001 + 005 + 006)..."
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/001_create_tables.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/005_conversations_refactor.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/006_message_reply_status_forward_seen.sql
	@echo "→ Schéma + seed user DB..."
	docker exec -i storm-postgres-user psql -U storm -d storm_user_db < infra/seed/000_create_user_tables.sql
	docker exec -i storm-postgres-user psql -U storm -d storm_user_db < infra/seed/001_seed_users.sql
	@echo "OK — voir docs/DEV-FRONT-LOCAL.md pour lancer les services."

# Infra + migrations (une commande)
dev-setup-docker: dev-infra-up
	@echo "Attente Postgres (message)..."
	@until docker exec storm-postgres-chat pg_isready -U storm -d storm_message_db 2>/dev/null; do sleep 2; done
	@echo "Attente Postgres (user)..."
	@until docker exec storm-postgres-user pg_isready -U storm -d storm_user_db 2>/dev/null; do sleep 2; done
	$(MAKE) dev-migrate-all-docker

# k8s : repartir d’un volume message DB vierge (après corruption checkpoint, etc.)
k8s-reset-postgres-message:
	@echo "⚠️  Supprime deployment + PVC postgres-message — PERTE des données message DB"
	kubectl delete deployment postgres-message -n $(NAMESPACE) --ignore-not-found
	kubectl delete pvc postgres-message-pvc -n $(NAMESPACE) --ignore-not-found
	kubectl apply -k infra/k8s/base/
	@echo "→ Surveille: kubectl get pods -n $(NAMESPACE) -l app=postgres-message -w"
	@echo "→ Puis: make migrate-message && make migrate-message-legacy && make migrate-message-006"

# Régénère message.pb.go (copie dans api/v1 car protoc sort par go_package)
proto-message:
	cd services/message && docker run --rm -v $$(pwd):/workspace -w /workspace znly/protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto
	cp services/message/github.com/Mathis-brgs/storm-project/services/message/api/v1/message.pb.go services/message/api/v1/message.pb.go
	rm -rf services/message/github.com
