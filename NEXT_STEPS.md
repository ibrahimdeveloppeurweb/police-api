# Prochaines Étapes

## 🎯 État Actuel

### ✅ Fait
1. **Schémas Ent créés** alignés avec frontend :
   - Control (avec date/heure séparées, permis, CNI, photos, etc.)
   - ProcesVerbal (avec statuts frontend)
   - Commissariat (avec responsable, statistiques)
   - Agent (avec grade, spécialités)
   - Alerte (avec type, urgence, véhicule, suspect)
   - TypeInfraction (avec catégorie, gravité, amendes)

2. **Module Controles** :
   - DTOs alignés avec interface Controle frontend
   - Repository mis à jour avec nouveau schéma
   - Service mis à jour avec mapping complet
   - Controller prêt

### ⏳ À Faire

1. **Générer le code Ent** :
   ```bash
   cd police-trafic-api-frontend-aligned
   go generate ./ent
   ```

2. **Mettre à jour les autres modules** (PV, Alertes, Commissariat, Agent, Admin) :
   - Mettre à jour les DTOs pour correspondre aux nouveaux schémas
   - Mettre à jour les repositories
   - Mettre à jour les services
   - Mettre à jour les controllers

3. **Tester avec les données mock du frontend** :
   - Vérifier que les structures correspondent
   - Vérifier que les mappings sont corrects
   - Vérifier que les endpoints fonctionnent

## 📋 Checklist par Module

### Module PV
- [ ] Mettre à jour DTOs avec nouveau schéma
- [ ] Mettre à jour Repository
- [ ] Mettre à jour Service
- [ ] Mettre à jour Controller

### Module Alertes
- [ ] Mettre à jour DTOs avec nouveau schéma
- [ ] Mettre à jour Repository
- [ ] Mettre à jour Service
- [ ] Mettre à jour Controller

### Module Commissariat
- [ ] Mettre à jour DTOs avec nouveau schéma
- [ ] Mettre à jour Repository
- [ ] Mettre à jour Service
- [ ] Mettre à jour Controller

### Module Agent/User
- [ ] Mettre à jour DTOs avec nouveau schéma
- [ ] Mettre à jour Repository
- [ ] Mettre à jour Service
- [ ] Mettre à jour Controller

### Module Admin
- [ ] Mettre à jour DTOs
- [ ] Mettre à jour Repository pour utiliser nouveaux schémas
- [ ] Mettre à jour Service
- [ ] Mettre à jour Controller

## 🔧 Commandes Utiles

```bash
# Générer le code Ent
go generate ./ent

# Installer les dépendances
go mod tidy

# Vérifier les erreurs
go build ./...

# Lancer l'application
go run cmd/server/main.go
```




