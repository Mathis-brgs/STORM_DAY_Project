#!/usr/bin/env bash
# kill-pods.sh — Tue aléatoirement des pods et mesure le temps de recovery
# Usage: ./kill-pods.sh [namespace] [nombre_de_kills]
# Exemple: ./kill-pods.sh storm-prod 3

set -euo pipefail

NAMESPACE="${1:-storm-prod}"
KILLS="${2:-1}"

echo "=== STORM Chaos — Kill Random Pods ==="
echo "Namespace : $NAMESPACE"
echo "Nombre de kills : $KILLS"
echo ""

# Exclure les pods système et de monitoring
EXCLUDE_LABELS="app in (prometheus,grafana,alertmanager,nats)"

for i in $(seq 1 "$KILLS"); do
  POD=$(kubectl get pods -n "$NAMESPACE" \
    --field-selector=status.phase=Running \
    -o jsonpath='{.items[*].metadata.name}' \
    | tr ' ' '\n' \
    | grep -v -E '^(prometheus|grafana|alertmanager|nats)' \
    | shuf -n 1)

  if [[ -z "$POD" ]]; then
    echo "Aucun pod éligible trouvé."
    exit 1
  fi

  echo "[$(date '+%H:%M:%S')] Kill #$i — Pod : $POD"
  START=$(date +%s%3N)

  kubectl delete pod "$POD" -n "$NAMESPACE" --grace-period=0 --force 2>/dev/null || true

  # Attendre que le pod soit recréé et Running
  echo "  En attente du redémarrage..."
  kubectl wait pod \
    -n "$NAMESPACE" \
    -l "$(kubectl get pod "$POD" -n "$NAMESPACE" -o jsonpath='{.metadata.labels}' 2>/dev/null \
         | python3 -c "import sys,json; d=json.load(sys.stdin); print(','.join(f'{k}={v}' for k,v in d.items() if k=='app'))" 2>/dev/null || echo "app=unknown")" \
    --for=condition=Ready \
    --timeout=120s 2>/dev/null || true

  END=$(date +%s%3N)
  RECOVERY=$(( END - START ))
  echo "  Recovery time : ${RECOVERY}ms"
  echo ""

  # Petite pause entre les kills
  if [[ $i -lt $KILLS ]]; then
    sleep 5
  fi
done

echo "=== Terminé ==="
echo "Résultats à reporter dans infra/chaos/README.md"
