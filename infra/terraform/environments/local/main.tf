# ==============================================================================
# ENVIRONNEMENT LOCAL - Plus de code Terraform actif
# ==============================================================================
#
# LocalStack (émulateur AWS) a été supprimé lors de la migration vers Azure.
# Il n'existe pas d'équivalent Azure complet pour tester Terraform localement.
#
# Le développement local utilise docker-compose à la racine du projet :
#   docker compose up -d
#
# Services locaux disponibles :
#   - PostgreSQL (ports 5432, 5433)
#   - Redis      (port 6379)
#   - NATS       (port 4222)
#   - MinIO      (port 9000) ← équivalent local de Azure Blob Storage
#
# Pour valider la syntaxe Terraform Azure sans accès Azure :
#   cd infra/terraform/environments/dev
#   tofu init
#   tofu validate
#
# Pour déployer en vrai, voir infra/terraform/environments/dev/
# ==============================================================================
