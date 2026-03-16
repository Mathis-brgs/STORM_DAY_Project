#!/usr/bin/env bash
# inject-latency.sh — Injecte de la latence réseau sur un pod via tc netem
# Usage: ./inject-latency.sh <pod-name> [namespace] [delay_ms] [jitter_ms] [duration_s]
# Exemple: ./inject-latency.sh gateway-7d9f8b-xxx storm-prod 200 50 60
#
# Prérequis : le container doit avoir NET_ADMIN capability, ou utiliser un debug container.
# Alternative sans CAP_NET_ADMIN : kubectl debug + nsenter

set -euo pipefail

POD="${1:?Usage: $0 <pod-name> [namespace] [delay_ms] [jitter_ms] [duration_s]}"
NAMESPACE="${2:-storm-prod}"
DELAY="${3:-100}"       # ms
JITTER="${4:-20}"       # ms
DURATION="${5:-30}"     # secondes

CONTAINER=$(kubectl get pod "$POD" -n "$NAMESPACE" \
  -o jsonpath='{.spec.containers[0].name}')

echo "=== STORM Chaos — Inject Latency ==="
echo "Pod       : $POD"
echo "Container : $CONTAINER"
echo "Namespace : $NAMESPACE"
echo "Latence   : ${DELAY}ms ± ${JITTER}ms"
echo "Durée     : ${DURATION}s"
echo ""

# Appliquer la latence
echo "[$(date '+%H:%M:%S')] Application de la latence..."
kubectl exec "$POD" -n "$NAMESPACE" -c "$CONTAINER" -- \
  tc qdisc add dev eth0 root netem delay "${DELAY}ms" "${JITTER}ms" distribution normal

echo "  Latence injectée. Observation pendant ${DURATION}s..."
echo "  → Surveiller Grafana : http://localhost:3001 (dashboard Gateway WS)"
echo ""

sleep "$DURATION"

# Supprimer la règle tc
echo "[$(date '+%H:%M:%S')] Suppression de la latence..."
kubectl exec "$POD" -n "$NAMESPACE" -c "$CONTAINER" -- \
  tc qdisc del dev eth0 root 2>/dev/null || true

echo "  Latence supprimée."
echo ""
echo "=== Terminé ==="
echo "Résultats à reporter dans infra/chaos/README.md"
