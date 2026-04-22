# usilo-kcfg — Contexte projet

## Objectif
CLI Go pour normaliser et gérer les kubeconfigs en entreprise.
Multi-OS : Linux, macOS, Windows/WSL.

## Stack
- Go + `cobra` (CLI) + `k8s.io/client-go` (kubeconfig)
- Pas de dépendances externes runtime (binaire standalone)

## Convention de nommage
- Fichier source : `kubeconfig_{user}@{cluster}.yaml`
- Contexte kubectl : `{user}@{cluster}`
- Exemple : `kubeconfig_john@prod-payments.yaml` → `john@prod-payments`

## Structure cible
```
kcfg/
├── cmd/          # commandes cobra (import, merge, list, remove, use, status)
├── internal/
│   ├── config/   # logique kubeconfig via client-go
│   └── normalize/
└── main.go
```

## Commandes prévues
- `kcfg init`
- `kcfg import -f <fichier> -u <user> -c <cluster> [--ctx <nom>]`
- `kcfg merge / list / remove / use / status / lint`

## Décisions d'architecture
- Merge via `clientcmd.Load` + `clientcmd.Merge` (pas de shell kubectl)
- Backup automatique avant chaque merge
- Permissions 600 sur tous les fichiers kubeconfig

## Référence comportement
Voir `docs/usilo-kcfg-reference.sh` — implémentation Bash de référence
à porter en Go avec le même comportement.