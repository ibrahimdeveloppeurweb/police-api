# 🔄 Refonte Complète de la Base de Données

**Date**: 26 Novembre 2024  
**Statut**: ✅ Schémas créés, prêt pour génération

## 📋 Travail effectué

### 1. Création des nouveaux schémas Ent

Tous les schémas ont été créés en parfait alignement avec le frontend TypeScript.

#### ✅ Schémas créés:

1. **agent.go** - Agents de police
   - Matricule unique
   - Informations personnelles
   - Grade (10 grades différents)
   - Statut (actif, repos, mission, formation, congé)
   - Spécialités
   - Rattachement au commissariat

2. **commissariat.go** - Commissariats
   - Nom et localisation GPS
   - Responsable (nom, grade, téléphone)
   - Statistiques des agents
   - Statistiques des contrôles
   - Revenus et taux de conformité
   - Statut (actif, maintenance, urgence)

3. **type_infraction.go** - Types d'infractions
   - Code unique
   - Libellé et catégorie (6 catégories)
   - Gravité (1 à 5)
   - Amende (min/max en FCFA)
   - Points du permis
   - Description et sanctions
   - Gestion de la récidive

4. **controle.go** - Contrôles routiers
   - Numéro unique
   - Date, heure et lieu
   - Agent et commissariat
   - Véhicule (immatriculation, marque, modèle, couleur, type)
   - Conducteur (nom, prénoms, téléphone)
   - Permis de conduire (numéro, expiration, points)
   - CNI (numéro, expiration)
   - Infractions constatées (JSON)
   - Montant total
   - Statut (en_cours, terminé, avec_infractions, conforme)
   - Observations et photos
   - Lien vers PV

5. **proces_verbal.go** - Procès-verbaux
   - Numéro unique
   - Lien vers contrôle
   - Date de génération
   - Statut (généré, notifié, payé, impayé, contentieux, annulé)
   - Infractions détaillées (JSON)
   - Montant total
   - Mode de paiement (espèces, mobile money, virement, chèque)
   - Date et référence de transaction
   - Délai de paiement

6. **alerte.go** - Système d'alertes
   - Type (5 types d'alertes)
   - Titre et message
   - Urgence (faible, moyen, élevé, critique)
   - Date
   - Commissariat concerné
   - Véhicule (immatriculation, marque, modèle)
   - Suspect (nom, description)
   - Statut (active, résolue, archivée)
   - Actions à entreprendre

### 2. Relations entre entités

```
Commissariat (1) ──> (N) Agent
Commissariat (1) ──> (N) Controle
Agent (1) ──> (N) Controle  
Controle (1) ──> (1) ProcesVerbal
```

### 3. Mixins utilisés

Tous les schémas utilisent:
- **UUIDMixin**: ID en UUID v4
- **TimeMixin**: created_at, updated_at
- **SoftDeleteMixin**: deleted_at (suppression logique)

### 4. Fichiers créés

✅ `/ent/schema/agent.go`
✅ `/ent/schema/commissariat.go`
✅ `/ent/schema/type_infraction.go`
✅ `/ent/schema/controle.go`
✅ `/ent/schema/proces_verbal.go`
✅ `/ent/schema/alerte.go`
✅ `/ent/schema/README.md`
✅ `/scripts/regenerate-ent.sh`

## 🚀 Prochaines étapes

### Étape 1: Générer le code Ent

```bash
cd /Users/mat/Development/importants/police-traffic-back-front/police-trafic-api-frontend-aligned

# Option 1: Via le Makefile
make generate

# Option 2: Via le script
chmod +x scripts/regenerate-ent.sh
./scripts/regenerate-ent.sh

# Option 3: Directement
go generate ./ent
```

### Étape 2: Nettoyer l'ancien fichier control.go

Après génération, supprimer manuellement:
- `/ent/schema/control.go`

Ou le renommer en `.old`

### Étape 3: Adapter les modules

Une fois Ent régénéré, il faudra adapter:

1. **Module controles**
   - Mettre à jour le repository pour utiliser `ent.Controle`
   - Adapter les DTO aux nouveaux champs
   - Mettre à jour les requêtes

2. **Module infractions** (déjà fait)
   - Utilise `TypeInfraction`
   - Déjà aligné avec le frontend

3. **Créer le module agents**
   - Repository, Service, Controller
   - Gestion des agents de police

4. **Créer le module commissariats**
   - Repository, Service, Controller
   - Gestion des commissariats

5. **Créer le module pv**
   - Repository, Service, Controller
   - Génération et gestion des PV

6. **Créer le module alertes**
   - Repository, Service, Controller
   - Système d'alertes

## 📊 Alignement Frontend-Backend

### Types TypeScript → Schémas Ent

| Frontend (TypeScript) | Backend (Ent) | Statut |
|----------------------|---------------|---------|
| `Agent` | `Agent` | ✅ Aligné |
| `Commissariat` | `Commissariat` | ✅ Aligné |
| `TypeInfraction` | `TypeInfraction` | ✅ Aligné |
| `Controle` | `Controle` | ✅ Aligné |
| `ProcesVerbal` | `ProcesVerbal` | ✅ Aligné |
| `Alerte` | `Alerte` | ✅ Aligné |

### Enums alignés

- **GradeAgent**: 10 valeurs identiques
- **StatusAgent**: 5 valeurs identiques
- **CategorieInfraction**: 6 valeurs identiques
- **StatusControle**: 4 valeurs identiques
- **TypeAlerte**: 5 valeurs identiques
- **NiveauUrgence**: 4 valeurs identiques
- **StatusPV**: 6 valeurs identiques
- **ModePaiement**: 4 valeurs identiques

## ⚠️ Points d'attention

1. **Ancien fichier control.go**
   - À supprimer après génération
   - Remplacé par `controle.go`

2. **Modules à adapter**
   - Le module controles actuel utilise l'ancienne structure
   - À mettre à jour après génération

3. **Base de données**
   - Les migrations seront automatiquement créées
   - Tester sur une base de données de développement d'abord

## ✨ Avantages de la refonte

1. **Alignement parfait** avec le frontend TypeScript
2. **Structure cohérente** et bien organisée
3. **Relations claires** entre entités
4. **Enums strictement typés**
5. **Documentation intégrée** dans les schémas
6. **Indexation optimisée** pour les requêtes
7. **Suppression logique** (soft delete) sur toutes les entités

## 📝 Notes

- Tous les montants sont en **FCFA** (Franc CFA)
- Les dates sont stockées en **format date** PostgreSQL
- Les timestamps en **timestamptz** (avec timezone)
- Les coordonnées GPS en **float** (latitude/longitude)
- Les listes complexes en **JSON** (infractions, photos, actions)

---

**Auteur**: Claude  
**Projet**: Police Nationale CI - API Backend  
**Version**: 1.0
