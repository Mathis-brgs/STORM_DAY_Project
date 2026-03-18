# Scale-up Storm Day — 8 avril 2026

## La veille (7 avril), faire dans l'ordre :

### 1. Modifier main.tf — PostgreSQL
```hcl
# Ligne ~103 dans main.tf, changer :
sku_name = "B_Standard_B1ms"
# → par :
sku_name = "GP_Standard_D4s_v3"
storage_mb = 131072  # 128 Go
```

### 2. Modifier main.tf — Redis
```hcl
# Ligne ~122 dans main.tf, changer :
capacity = 0
sku_name = "Basic"
# → par :
capacity = 2
sku_name = "Standard"
```

### 3. Modifier main.tf — AKS
```hcl
# Dans le module "aks", changer :
vm_size    = "Standard_B2s"
# → par :
vm_size    = "Standard_D4s_v3"

# Et dans modules/aks/variables.tf, changer les defaults :
min_node_count = 2
max_node_count = 6
```

### 4. Appliquer
```bash
cd infra/terraform/environments/dev
tofu plan   # vérifier les changements
tofu apply  # ~30 min (PostgreSQL et AKS prennent du temps)
```

### 5. Vérifier après apply
```bash
# Vérifier les nodes AKS
az aks get-credentials --resource-group storm-dev --name storm-aks-dev
kubectl get nodes

# Redéployer les services si besoin (les pods redémarrent automatiquement)
kubectl rollout restart deployment -n storm
```

### 6. Scale message-service → 20 replicas + augmenter HPA max

```bash
# Augmenter HPA max de 5 à 20
kubectl patch hpa message-hpa -n storm --type='merge' \
  -p '{"spec":{"maxReplicas":20}}'

# Scale immédiat à 10 replicas
kubectl scale deployment message-service -n storm --replicas=10
kubectl rollout status deployment/message-service -n storm
```

**Capacité attendue (infra seule, sans batching) :**
- 10 replicas × ~400 msg/s = **~4 000 msg/s** en DB
- 20 replicas × ~400 msg/s = **~8 000 msg/s** en DB

**Capacité attendue (avec batching — voir Prio 8 plan) :**
- 10 replicas × ~5 000 msg/s = **~50 000 msg/s** en DB ✅

---

### 7. Scale gateway → 8 replicas + déployer fix SO_REUSEPORT

> ⚠️ Prérequis : fix SO_REUSEPORT doit être mergé dans l'image gateway avant Storm Day (voir PLAN_2026-03-18.md Prio 7)

```bash
# Scale gateway à 8 replicas
kubectl scale deployment gateway-service -n storm --replicas=8
kubectl rollout status deployment/gateway-service -n storm

# Scale user-service à 8 replicas
kubectl scale deployment user-service -n storm --replicas=8
kubectl rollout status deployment/user-service -n storm

# Vérifier la distribution sur les nodes
kubectl get pods -n storm -l app=gateway-service -o wide
```

### 7. Test de charge final avant Storm Day

```bash
# Modifier le target dans ws-stress-max.js → 100 000 VUs, ramp 30 min
# Puis lancer
k6 run /tmp/ws-stress-max.js
```

Objectifs à valider :
- **100 000 connexions WS simultanées** tenues ≥ 2 min
- **50 000 messages/s** écrits en DB (minimum)
- **500 000 livraisons/s** via NATS broadcast → WS clients
- Gateway UP, CPU < 70%, RSS < 80% de la limit
- Résilience : simuler des incidents (pod kill, node drain) et mesurer le temps de recovery

**Prérequis code avant Storm Day :**
- ✅ SO_REUSEPORT gateway (Prio 7 — 100k WS)
- ✅ Write batching message-service (Prio 8 — 50k msg/s DB)

---

## Après le Storm Day (8 avril soir), revenir aux valeurs dev :

Remettre les valeurs initiales dans main.tf et relancer `tofu apply`.
Ou faire `tofu destroy` si le projet est terminé.

---

## Coûts estimés Storm Day (~8h)

| Ressource | Dev | Storm Day | Delta/h |
|---|---|---|---|
| PostgreSQL GP_D4s_v3 | $0.17/h | $0.80/h | +$0.63/h |
| Redis Standard C2 | $0.22/h | $0.58/h | +$0.36/h |
| AKS 4x D4s_v3 | $0.30/h | $1.20/h | +$0.90/h |
| **Total delta** | | | **+$1.89/h** |

**Coût total Storm Day : ~$15 pour 8h** (largement dans le budget).
