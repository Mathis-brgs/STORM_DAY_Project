# Chaos Engineering — STORM

Scripts de chaos pour valider la résilience du cluster avant et pendant le STORM DAY.

---

## Scripts disponibles

### `kill-pods.sh` — Kill aléatoire de pods

```bash
./infra/chaos/kill-pods.sh [namespace] [nombre_de_kills]
# Exemple :
./infra/chaos/kill-pods.sh storm-prod 3
```

**Ce que ça teste :** la capacité du cluster à redémarrer les pods et maintenir la disponibilité du service (liveness/readiness probes, restart policy, réplication).

**Métriques à observer :**
- Recovery time (affiché dans la console)
- Erreurs HTTP 5xx dans Grafana pendant le kill
- Connexions WS actives (dashboard Gateway WS)

---

### `inject-latency.sh` — Injection de latence réseau (tc netem)

```bash
./infra/chaos/inject-latency.sh <pod-name> [namespace] [delay_ms] [jitter_ms] [duration_s]
# Exemple :
./infra/chaos/inject-latency.sh gateway-7d9f8b-xxx storm-prod 200 50 60
```

**Ce que ça teste :** le comportement sous latence réseau — timeouts NATS, retry logic, dégradation des p95.

**Prérequis :** le container doit avoir `NET_ADMIN` capability. Ajouter dans le manifest K8s si nécessaire :
```yaml
securityContext:
  capabilities:
    add: ["NET_ADMIN"]
```

---

## Résultats des sessions chaos

### Session 1 — _à compléter_

| Date | Script | Cible | Paramètres | Recovery / Impact | Notes |
|------|--------|-------|-----------|-------------------|-------|
| JJ/MM | kill-pods | gateway | 1 kill | _ms | |
| JJ/MM | kill-pods | message-service | 1 kill | _ms | |
| JJ/MM | inject-latency | gateway | 200ms/50ms, 30s | p95 WS +_ms | |

---

## Résultats attendus (objectifs STORM DAY)

| Scénario | Objectif |
|----------|----------|
| Kill gateway pod | Recovery < 15s, 0 message perdu |
| Kill message-service | Recovery < 30s, NATS buffer les messages |
| Kill user-service | Auth dégradée < 10s |
| Latence 100ms | p95 WS < 300ms |
| Latence 500ms | p95 WS < 1000ms, pas de déconnexion massive |

---

## Commandes utiles pendant les tests

```bash
# Observer les pods en temps réel
kubectl get pods -n storm-prod -w

# Logs d'un service
kubectl logs -n storm-prod -l app=gateway -f

# Métriques Prometheus (connexions WS actives)
curl http://localhost:9090/api/v1/query?query=storm_ws_connections_active

# Grafana
open http://localhost:3001
```
