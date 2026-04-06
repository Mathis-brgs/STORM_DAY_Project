# Post-Mortem — Storm Day STORM

> **Date** : [À compléter — date du Storm Day]
> **Durée de l'incident** : [HH:MM – HH:MM]
> **Sévérité** : [P1 / P2 / P3]
> **Statut** : [Résolu / En cours]

---

## Résumé exécutif

[2-3 phrases décrivant ce qui s'est passé, l'impact et comment c'a été résolu.]

Exemple : *Lors du Storm Day, le service Message a subi une saturation de la connexion PostgreSQL à partir de 15 000 messages/s, causant une dégradation de 30% du throughput pendant 12 minutes. Le problème a été résolu par le scale-up manuel du nombre de répliques de 3 à 8.*

---

## Timeline

| Heure | Événement |
|-------|-----------|
| HH:MM | Début du test de charge k6 (100 VUs) |
| HH:MM | Passage à 1 000 VUs — latence p95 passe de 50ms à 200ms |
| HH:MM | Passage à 5 000 VUs — premières erreurs 5xx détectées |
| HH:MM | Alerte Prometheus déclenchée (error_rate > 1%) |
| HH:MM | Identification du goulot (PostgreSQL connections saturées) |
| HH:MM | Scale-up Message Service : 3 → 8 répliques |
| HH:MM | Retour à la normale (error_rate < 0.1%) |
| HH:MM | Pic de 100k connexions WS atteint / non atteint |
| HH:MM | Fin du test |

---

## Impact

| Métrique | Valeur |
|----------|--------|
| Durée de dégradation | ___ minutes |
| Taux d'erreur max | ___% |
| Connexions WS max atteintes | ___ / 100 000 |
| Throughput messages max | ___ msg/s / 500 000 |
| Services impactés | Gateway, Message, [autres] |

---

## Causes racines

### Cause primaire

**[Exemple]** : La limite de connexions PostgreSQL (`max_connections=100`) a été atteinte avec 10 répliques de Message Service × 10 connexions/réplique = 100 connexions.

### Causes contributives

1. **[Exemple]** : Absence de connection pooler (pgBouncer) devant PostgreSQL
2. **[Exemple]** : `synchronous_commit=on` (défaut PostgreSQL) causant des writes lents sous charge
3. **[Exemple]** : NATS single-node créant un goulot sur le bus de messages

---

## Ce qui a fonctionné

- HPA Gateway a scalé automatiquement de 2 à 10 pods en moins de 3 minutes
- Les alertes Prometheus ont déclenché dans les 30 secondes suivant la dégradation
- Le BatchWriter Message a absorbé les pics de charge sans perte de message
- Le JWT local validation Gateway a tenu sans dégradation
- La procédure de scale-up manuel était documentée et exécutée en < 2 min

---

## Ce qui n'a pas fonctionné

- [Exemple] NATS single-node a atteint sa limite de throughput à ~200k msg/s
- [Exemple] L'overlay load-test n'avait pas été appliqué avant le test
- [Exemple] Manque de data sur les connexions PostgreSQL dans Grafana (dashboard incomplet)
- [Exemple] Le Media Service n'avait pas de tests de charge — comportement inconnu sous load

---

## Métriques observées

### Gateway
- Connexions WebSocket actives (pic) : ___
- p95 latence WS handshake : ___ ms
- CPU utilisation (pic) : ___%
- Mémoire utilisation (pic) : ___ Mi

### Message Service
- Messages reçus/s (pic) : ___
- Batch inserts/s (pic) : ___
- Erreurs base de données : ___
- Latence p95 batch insert : ___ ms

### NATS
- Messages publiés/s (pic) : ___
- Mémoire utilisée (pic) : ___ Mi
- Erreurs de publication : ___

### PostgreSQL Message
- Connexions actives (pic) : ___ / max ___
- Queries/s (pic) : ___
- Temps moyen requête INSERT : ___ ms

---

## Actions correctives

### Immédiates (avant prochain Storm Day)

| # | Action | Responsable | Délai | Statut |
|---|--------|-------------|-------|--------|
| 1 | Appliquer overlay load-test avant le test | Infra | J-1 | ☐ |
| 2 | Augmenter `max_connections` PostgreSQL → 200 | Infra | J-1 | ☐ |
| 3 | Déployer NATS en cluster 3 nœuds | Infra | J-1 | ☐ |
| 4 | Ajouter dashboard PostgreSQL connections dans Grafana | Monitoring | J-2 | ☐ |

### Moyen terme (semaines suivantes)

| # | Action | Priorité | Description |
|---|--------|----------|-------------|
| 1 | Installer pgBouncer | Haute | Connection pooler devant PostgreSQL pour absorber les pics |
| 2 | Tests de charge Media Service | Moyenne | Identifier les limites du service avant production |
| 3 | Chaos engineering automatisé | Moyenne | Scripts kill-pods.sh, inject-latency.sh |
| 4 | Alertes sur file descriptors | Basse | Prévenir l'épuisement des FD sur les nodes K8s |

---

## Leçons apprises

1. **Tester l'overlay load-test en avance** : ne pas appliquer les changements de ressources le jour J
2. **Le goulot est toujours dans la base de données** : PostgreSQL est le composant le moins scalable horizontalement
3. **Monitorer les connexions DB** : métrique critique manquante dans les dashboards initiaux
4. **La validation JWT locale est un vrai gain** : aucune dégradation du User Service pendant le pic
5. **Le BatchWriter est efficace** : 500k msg/s réduit à ~1k inserts/s, PostgreSQL peut absorber

---

## Résultats finaux Storm Day

| Objectif | Cible | Atteint | Observation |
|----------|-------|---------|-------------|
| Connexions WS simultanées | 100 000 | ___ | ___ |
| Throughput messages | 500 000 msg/s | ___ msg/s | ___ |
| Latence WS p95 | < 200ms | ___ ms | ___ |
| Taux d'erreur | < 1% | ___% | ___ |
| Temps de récupération | < 5 min | ___ min | ___ |
