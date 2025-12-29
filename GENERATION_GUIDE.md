# Guide de Génération du Code Ent

## 📋 Prérequis

1. **Go installé** (version 1.21 ou supérieure)
2. **Ent CLI installé** :
   ```bash
   go install entgo.io/ent/cmd/ent@latest
   ```

## 🔧 Génération du Code Ent

Une fois que Go est installé et dans votre PATH, exécutez :

```bash
cd police-trafic-api-frontend-aligned
go generate ./ent
```

Cette commande va :
1. Lire tous les schémas dans `ent/schema/`
2. Générer le code Ent dans `ent/` (entités, clients, queries, etc.)
3. Créer les fichiers nécessaires pour utiliser Ent avec PostgreSQL

## 📁 Schémas Créés

Les schémas suivants ont été créés et sont prêts à être générés :

1. **Control** (`ent/schema/control.go`)
   - Contrôles routiers avec tous les champs frontend
   - Relations : Agent, Commissariat, ProcesVerbal

2. **ProcesVerbal** (`ent/schema/proces_verbal.go`)
   - Procès-verbaux avec statuts alignés frontend
   - Relation : Control

3. **Alerte** (`ent/schema/alerte.go`)
   - Alertes sécuritaires
   - Relation : Commissariat

4. **Commissariat** (`ent/schema/commissariat.go`)
   - Commissariats avec responsable et statistiques
   - Relations : Agents, Controls, Alertes

5. **Agent** (`ent/schema/agent.go`)
   - Agents avec grade et spécialités
   - Relation : Commissariat

6. **TypeInfraction** (`ent/schema/type_infraction.go`)
   - Types d'infractions avec catégories et amendes

## ⚠️ Après la Génération

Après avoir généré le code Ent, vous devrez :

1. **Vérifier les imports** dans les repositories et services
2. **Tester la compilation** :
   ```bash
   go build ./...
   ```
3. **Créer les migrations** (si nécessaire) :
   ```bash
   go run -mod=mod entgo.io/ent/cmd/ent migrate generate ./schema
   ```

## 🔍 Vérification

Pour vérifier que tout est correct :

```bash
# Vérifier la compilation
go build ./...

# Vérifier les imports
go mod tidy

# Lancer les tests (si disponibles)
go test ./...
```

## 📝 Notes

- Les schémas sont alignés **exactement** avec les types TypeScript du frontend
- Tous les enums et status correspondent aux valeurs frontend
- Les relations entre entités sont définies correctement
- Les champs JSON (infractions, photos, actions, etc.) sont configurés

Une fois le code généré, le projet sera prêt à être utilisé avec le frontend !




