# 🚀 Guide de Génération - Police Nationale CI API

## ✅ Étape actuelle: Schémas créés, prêt pour la génération

Les schémas de base de données ont été complètement refaits et sont maintenant parfaitement alignés avec le frontend TypeScript.

## 📝 Ce qui a été fait

1. ✅ **6 nouveaux schémas Ent créés**:
   - Agent (agents de police)
   - Commissariat (commissariats)
   - TypeInfraction (catalogue des infractions)
   - Controle (contrôles routiers)
   - ProcesVerbal (procès-verbaux)
   - Alerte (système d'alertes)

2. ✅ **Relations définies** entre toutes les entités
3. ✅ **Enums alignés** avec le frontend  
4. ✅ **Documentation** complète dans REFONTE_BDD.md

## 🔧 Commandes à exécuter MAINTENANT

### 1. Générer le code Ent

Ouvrez un terminal et exécutez:

\`\`\`bash
cd /Users/mat/Development/importants/police-traffic-back-front/police-trafic-api-frontend-aligned

# Générer le code Ent
make generate
\`\`\`

Ou directement:

\`\`\`bash
go generate ./ent
\`\`\`

Cette commande va générer:
- Les modèles d'entités
- Les builders (Create, Update, Query, Delete)
- Les mutations
- Les relations (Edges)
- Les migrations de base de données

### 2. Supprimer l'ancien fichier

Après la génération, supprimez l'ancien schéma:

\`\`\`bash
rm ent/schema/control.go
\`\`\`

### 3. Vérifier la génération

\`\`\`bash
# Lister les nouveaux fichiers générés
ls -la ent/*.go

# Vérifier qu'il y a bien:
# - agent.go
# - commissariat.go
# - controle.go  
# - procesverbal.go
# - typeinfraction.go
# - alerte.go
\`\`\`

## 📋 Après la génération

Une fois la génération terminée, il faudra:

### 1. Mettre à jour le module controles

Les fichiers à adapter:
- `internal/modules/controles/repository.go`
- `internal/modules/controles/dto.go`
- `internal/modules/controles/service.go`

**Changements nécessaires**:
- Remplacer `ent.Control` par `ent.Controle`
- Adapter les noms de champs aux nouveaux noms français
- Mettre à jour les requêtes

### 2. Mettre à jour le module infractions

Les fichiers à adapter:
- `internal/modules/infractions/repository.go` 
- `internal/modules/infractions/dto.go`
- `internal/modules/infractions/service.go`

**Changements nécessaires**:
- S'assurer que l'entité `TypeInfraction` est bien utilisée
- Vérifier les noms de champs

### 3. Créer les nouveaux modules

**Module agents** (à créer):
- Repository
- Service  
- Controller
- DTO
- Module

**Module commissariats** (à adapter):
- Mettre à jour avec le nouveau schéma
- Ajouter les nouvelles fonctionnalités

**Module pv** (à adapter):
- Utiliser `ProcesVerbal` au lieu de l'ancien schéma
- Implémenter la génération de PV
- Gérer les paiements

**Module alertes** (à adapter):
- Utiliser la nouvelle entité `Alerte`
- Implémenter le système d'alertes

## 🎯 Ordre d'exécution recommandé

1. **Maintenant**: `make generate`
2. **Ensuite**: Supprimer `control.go`
3. **Puis**: Adapter les modules existants
4. **Enfin**: Créer les nouveaux modules

## ⚠️ Important

- **NE PAS** modifier les fichiers dans `ent/` sauf ceux dans `ent/schema/` et `ent/mixin/`
- Les fichiers générés seront ÉCRASÉS à chaque génération
- Toujours modifier les schémas sources dans `ent/schema/`

## 🐛 En cas d'erreur

Si la génération échoue:

1. Vérifier que tous les imports sont corrects dans les schémas
2. Vérifier que go.mod est à jour: `go mod tidy`
3. Vérifier les erreurs de syntax Go
4. Consulter les logs d'erreur

## 📞 Besoin d'aide?

Si vous rencontrez des problèmes:
1. Copiez le message d'erreur complet
2. Vérifiez quel fichier pose problème
3. Je pourrai vous aider à corriger

---

## 🎉 Une fois terminé

Après génération et adaptation des modules, vous pourrez:
- Lancer l'API: `make run`
- Tester les endpoints
- Voir les migrations de base de données
- Vérifier l'alignement avec le frontend

**Prêt à continuer? Exécutez `make generate` maintenant!** 🚀
