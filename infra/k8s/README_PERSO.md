# ⚡ Documentation Technique - Projet Storm (Gateway)

Ce document regroupe toutes les commandes pour installer, lancer et développer sur le Gateway.

## 🏗️ 1. Installation complète (De zéro)
*À faire uniquement si tu n'as pas de cluster ou si tu viens de le supprimer.*

### A. Créer le cluster
```powershell
k3d cluster create storm -p "30080:30080@server:0"
B. Builder et Importer l'image Gateway
PowerShell
docker build -t storm/gateway-service:latest services/gateway/
k3d image import storm/gateway-service:latest -c storm
C. Configurer la connexion (IMPORTANT)
Si kubectl get nodes échoue ou timeout :

Faire docker ps et noter le port local (ex: 58479).

Ouvrir la config : notepad $HOME\.kube\config.

Remplacer server: ... par server: https://127.0.0.1:<PORT>.

D. Déployer les ressources
Dans l'ordre précis :

PowerShell
# 1. Le namespace
kubectl apply -f infra/k8s/base/namespace.yaml

# 2. Les dépendances (Redis/Nats)
kubectl apply -f infra/k8s/base/redis.yaml
kubectl apply -f infra/k8s/base/nats.yaml

# 3. Le Gateway
kubectl apply -f infra/k8s/base/gateway-service.yaml
🟢 2. Démarrage Quotidien (Start)
Lancer Docker Desktop.

Démarrer le cluster :

PowerShell
k3d cluster start storm
Vérifier la connexion (Obligatoire) :

PowerShell
kubectl get nodes
✅ Si "Ready" : C'est bon.

❌ Si Erreur : Le port a changé. Voir section "Dépannage Connexion" plus bas.

Vérifier que les services tournent :

PowerShell
kubectl get pods -n storm
🔴 3. Fin de journée (Stop)
Ne supprime pas le cluster, mets-le en pause :

PowerShell
k3d cluster stop storm
🛠️ 4. Workflow de Développement (Boucle de dev)
À faire à chaque modification du code dans services/gateway/.

Re-builder l'image :

PowerShell
docker build -t storm/gateway-service:latest services/gateway/
Mettre à jour le cluster :

PowerShell
k3d image import storm/gateway-service:latest -c storm
Redémarrer le Gateway :

PowerShell
kubectl rollout restart deployment/gateway-service -n storm
Suivre les logs :

PowerShell
kubectl logs -f deployment/gateway-service -n storm
🚑 5. Dépannage Connexion (Si kubectl plante)
Si tu as une erreur connectex ou dial tcp, c'est que le port Docker a changé au redémarrage.

Trouver le nouveau port :

PowerShell
docker ps --format "table {{.Names}}\t{{.Ports}}"
Cherche la ligne k3d-storm-serverlb et note le port qui pointe vers 6443 (ex: 58479).

Éditer la config :

PowerShell
notepad $HOME\.kube\config
Modifier l'IP : Cherche la ligne server: https://... et remplace par :

YAML
server: [https://127.0.0.1](https://127.0.0.1):<TON_NOUVEAU_PORT>
(Exemple : server: https://127.0.0.1:58479). Sauvegarde (Ctrl+S).

Réessayer : kubectl get nodes

🔗 Tests Rapides
Url Healthcheck : http://localhost:30080/health

Voir tout : kubectl get all -n storm