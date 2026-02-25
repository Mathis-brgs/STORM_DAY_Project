# Variables
CLUSTER_NAME=storm
NAMESPACE=storm
GATEWAY_PORT=30080
IMAGES=storm/user-service:latest storm/gateway-service:latest storm/message-service:latest storm/media-service:latest

.PHONY: up down clean build deploy import restart status logs logs-media migrate-message migrate-user seed-message seed-user proto-message

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

# Build toutes les images Docker
# Note: On utilise :latest pour coller à la variable IMAGES
build:
	docker build -t storm/user-service:latest services/user/
	docker build -t storm/gateway-service:latest services/gateway/
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

# --- Migrations & Seeds (dev local avec Docker Compose) ---

# Migrations DB Message (001 = tables, 003 = messages, 004 = groups.user_id uuid)
migrate-message:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/001_create_tables.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/003_messages_uuid.sql
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/004_groups_user_id_uuid.sql

# Seed DB Message (groups + messages)
seed-message:
	docker exec -i storm-postgres-chat psql -U storm -d storm_message_db < services/message/migrations/002_seed_data.sql

# Seed DB User (nécessite que le user-service ait créé les tables)
seed-user:
	docker exec -i storm-postgres-user psql -U storm -d storm_user_db < infra/seed/001_seed_users.sql

# Régénère message.pb.go (copie dans api/v1 car protoc sort par go_package)
proto-message:
	cd services/message && docker run --rm -v $$(pwd):/workspace -w /workspace znly/protoc -I. --go_out=. --go_opt=paths=source_relative api/v1/message.proto
	cp services/message/github.com/Mathis-brgs/storm-project/services/message/api/v1/message.pb.go services/message/api/v1/message.pb.go
	rm -rf services/message/github.com