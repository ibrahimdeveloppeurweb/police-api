# Nettoyage des Schémas Ent

## ✅ Schémas Conservés

Les schémas suivants sont conservés car ils sont utilisés dans le projet aligné avec le frontend :

1. **control.go** - Contrôles routiers
2. **proces_verbal.go** - Procès-verbaux
3. **alerte.go** - Alertes sécuritaires
4. **commissariat.go** - Commissariats
5. **agent.go** - Agents
6. **type_infraction.go** - Types d'infractions

## 🗑️ Schémas Supprimés

Les schémas suivants ont été supprimés car ils ne sont pas utilisés dans le nouveau projet :

### Anciens schémas remplacés
- `alert.go` → remplacé par `alerte.go`
- `officier.go` → remplacé par `agent.go`
- `police_station.go` → remplacé par `commissariat.go`
- `ticket.go` → remplacé par `proces_verbal.go`

### Schémas géographiques (non utilisés)
- `arrondissement.go`
- `city.go`
- `district.go`
- `municipality.go`
- `neighborhood.go`
- `region.go`
- `subprefecture.go`

### Schémas métier (non utilisés)
- `alert_notification.go`
- `checkitem.go`
- `checkoption.go`
- `commissioner.go`
- `driver.go`
- `file.go`
- `inspection.go`
- `payment.go`
- `permission.go`
- `role.go`
- `ticket_violation.go`
- `user.go`
- `vehicle.go`
- `violation.go`

## 📝 Notes

- Tous les schémas supprimés étaient des schémas de l'ancien projet
- Les nouveaux schémas sont alignés avec les types TypeScript du frontend
- Après suppression, il faudra régénérer le code Ent : `go generate ./ent`
- Les fichiers générés dans `ent/` (hors `schema/`) seront automatiquement régénérés




