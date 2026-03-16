# Variables
CLUSTER_NAME=storm
NAMESPACE=storm
GATEWAY_PORT=30080
IMAGES=storm/user-service:latest storm/gateway-service:latest storm/message-service:latest storm/media-service:latest
POSTGRES_USER=storm
MESSAGE_DB_NAME=message_db
USER_DB_NAME=user_db

.PHONY: up down clean build deploy import restart status logs logs-media \
	migrate-message migrate-message-legacy seed-message seed-user \
	migrate-message-docker migrate-message-legacy-docker seed-message-docker seed-user-docker \
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

# Seed DB Message (conversations + messages)
seed-message:
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=postgres-message -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$POD" ]; then \
		echo "Pod postgres-message introuvable dans le namespace $(NAMESPACE)."; \
		echo "Deploie d'abord K8s: kubectl apply -k infra/k8s/base/"; \
		exit 1; \
	fi; \
	kubectl exec -i -n $(NAMESPACE) $$POD -- psql -U $(POSTGRES_USER) -d $(MESSAGE_DB_NAME) < services/message/migrations/002_seed_data.sql

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

seed-message-docker:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/002_seed_data.sql

seed-user-docker:
	docker exec -i storm-postgres-user psql -U storm -d storm_user_db < infra/seed/001_seed_users.sql

# Régénère message.pb.go (copie dans api/v1 car protoc sort par go_package)
proto-message:
	cd services/message && docker run --rm -v $$(pwd):/workspace -w /workspace znly/protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto
	cp services/message/github.com/Mathis-brgs/storm-project/services/message/api/v1/message.pb.go services/message/api/v1/message.pb.go
	rm -rf services/message/github.com
