# Variables
CLUSTER_NAME=storm
NAMESPACE=storm
GATEWAY_PORT=30080
IMAGES=storm/user-service:latest storm/gateway-service:latest storm/message-service:latest storm/media-service:latest

.PHONY: up down clean build deploy import restart status logs logs-media

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