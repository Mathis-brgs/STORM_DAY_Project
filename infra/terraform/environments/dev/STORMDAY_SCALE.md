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
