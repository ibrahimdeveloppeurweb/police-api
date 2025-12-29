# Guide d'Intégration du Mode Contenant avec Inventaire - Objets Perdus

## 📋 Résumé des Modifications

Ce guide vous aide à intégrer la fonctionnalité de contenants avec inventaire pour les objets perdus. Les modifications permettent de :

- ✅ Déclarer un objet simple (mode classique)
- ✅ Déclarer un contenant (sac, valise, portefeuille) avec inventaire détaillé
- ✅ Enregistrer tous les objets contenus avec leurs détails spécifiques
- ✅ Gérer les pièces d'identité et cartes bancaires avec leurs numéros

## 🔧 Modifications Effectuées

### 1. **Schéma Ent** (`ent/schema/objet_perdu.go`)
✅ Ajout de 2 nouveaux champs :
- `is_container` (bool) : Indique si c'est un contenant
- `container_details` (JSON) : Détails du contenant + inventaire complet

### 2. **Types** (`internal/modules/objets-perdus/types.go`)
✅ Ajout des structures :
- `InventoryItem` : Structure d'un objet dans l'inventaire
- `ContainerDetails` : Détails du contenant avec inventaire
- Mise à jour de `CreateObjetPerduRequest` et `UpdateObjetPerduRequest`

### 3. **Repository** (`internal/infrastructure/repository/objet_perdu_repository.go`)
✅ Mise à jour des méthodes Create et Update pour gérer les nouveaux champs

### 4. **Service** (`internal/modules/objets-perdus/service.go`)
✅ Logique de création et formatage pour gérer :
- Le mode contenant
- La sérialisation/désérialisation de l'inventaire
- La conversion des détails

## 🚀 Étapes d'Application

### Étape 1 : Régénérer le Code Ent

```bash
# Se placer dans le dossier du projet backend
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Régénérer le code Ent
go generate ./ent

# Vérifier qu'il n'y a pas d'erreurs
go build ./...
```

### Étape 2 : Créer et Appliquer la Migration

```bash
# Créer une nouvelle migration
go run cmd/migrate/main.go

# OU utiliser le Makefile si disponible
make migrate

# La migration ajoutera automatiquement les colonnes :
# - is_container (BOOLEAN DEFAULT FALSE)
# - container_details (JSONB)
```

### Étape 3 : Vérifier la Migration

```bash
# Se connecter à PostgreSQL
psql -U postgres -d police_trafic_db

# Vérifier les colonnes
\d objets_perdus

# Vous devriez voir :
# - is_container | boolean | NOT NULL DEFAULT false
# - container_details | jsonb |
```

### Étape 4 : Redémarrer le Backend

```bash
# Arrêter le serveur actuel (Ctrl+C)

# Compiler et démarrer
go run cmd/server/main.go

# OU si vous avez un Makefile
make run
```

### Étape 5 : Tester l'API

#### Test 1 : Créer un objet simple (mode classique)

```bash
curl -X POST http://localhost:8080/api/objets-perdus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "typeObjet": "Téléphone portable",
    "description": "iPhone 13 noir",
    "isContainer": false,
    "declarant": {
      "nom": "KOUASSI",
      "prenom": "Jean",
      "telephone": "+225 07 00 00 00 00"
    },
    "lieuPerte": "Plateau",
    "datePerte": "2024-12-10"
  }'
```

#### Test 2 : Créer un contenant avec inventaire

```bash
curl -X POST http://localhost:8080/api/objets-perdus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "typeObjet": "Sac à dos",
    "description": "Sac à dos noir contenant plusieurs objets",
    "isContainer": true,
    "containerDetails": {
      "type": "sac_dos",
      "couleur": "NOIR",
      "marque": "NIKE",
      "taille": "MOYEN",
      "signesDistinctifs": "LOGO NIKE BLANC",
      "inventory": [
        {
          "category": "telephone",
          "name": "IPHONE 13 PRO",
          "color": "NOIR",
          "brand": "APPLE",
          "serial": "IMEI123456789"
        },
        {
          "category": "identite",
          "name": "CARTE NATIONALE D'\''IDENTITE",
          "color": "BLEU",
          "identityType": "CNI",
          "identityNumber": "CI20240001",
          "identityName": "KOUASSI JEAN"
        },
        {
          "category": "carte",
          "name": "CARTE VISA",
          "color": "BLEU",
          "cardType": "VISA",
          "cardBank": "SGBCI",
          "cardLast4": "1234"
        },
        {
          "category": "portefeuille",
          "name": "PORTEFEUILLE CUIR",
          "color": "MARRON",
          "brand": "LOUIS VUITTON"
        }
      ]
    },
    "declarant": {
      "nom": "KOUASSI",
      "prenom": "Jean",
      "telephone": "+225 07 00 00 00 00",
      "email": "jean.kouassi@example.com"
    },
    "lieuPerte": "Autoroute du Nord",
    "adresseLieu": "Sortie Abobo",
    "datePerte": "2024-12-10",
    "heurePerte": "14:30"
  }'
```

#### Test 3 : Récupérer un objet avec inventaire

```bash
curl -X GET http://localhost:8080/api/objets-perdus/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

La réponse inclura :
```json
{
  "id": "...",
  "numero": "OBP-ABI-COM-2024-0001",
  "isContainer": true,
  "containerDetails": {
    "type": "sac_dos",
    "couleur": "NOIR",
    "marque": "NIKE",
    "inventory": [
      {
        "id": 1,
        "category": "telephone",
        "name": "IPHONE 13 PRO",
        ...
      }
    ]
  }
}
```

## 🗄️ Structure de la Base de Données

### Nouvelle Structure JSON de `container_details`

```json
{
  "type": "sac_dos",           // Type de contenant
  "couleur": "NOIR",           // Couleur du contenant
  "marque": "NIKE",            // Marque
  "taille": "MOYEN",           // Taille
  "signesDistinctifs": "LOGO", // Signes distinctifs
  "inventory": [               // Inventaire des objets
    {
      "id": 1,
      "category": "telephone",
      "icon": "smartphone",
      "name": "IPHONE 13 PRO",
      "color": "NOIR",
      "brand": "APPLE",
      "serial": "IMEI123456789",
      
      // Champs spécifiques pour identité
      "identityType": "CNI",
      "identityNumber": "CI20240001",
      "identityName": "KOUASSI JEAN",
      
      // Champs spécifiques pour cartes
      "cardType": "VISA",
      "cardBank": "SGBCI",
      "cardLast4": "1234"
    }
  ]
}
```

## 📊 Requêtes SQL Utiles

### Rechercher les contenants
```sql
SELECT * FROM objets_perdus WHERE is_container = true;
```

### Rechercher par type de contenant
```sql
SELECT * FROM objets_perdus 
WHERE is_container = true 
AND container_details->>'type' = 'sac_dos';
```

### Rechercher les objets avec un item spécifique dans l'inventaire
```sql
SELECT * FROM objets_perdus 
WHERE is_container = true 
AND container_details->'inventory' @> '[{"category": "telephone"}]';
```

### Compter les items dans l'inventaire
```sql
SELECT 
  numero,
  jsonb_array_length(container_details->'inventory') as nb_items
FROM objets_perdus 
WHERE is_container = true;
```

## 🐛 Dépannage

### Erreur : "column is_container does not exist"
**Solution :** La migration n'a pas été appliquée
```bash
go run cmd/migrate/main.go
```

### Erreur : "cannot unmarshal"
**Solution :** Vérifier le format JSON de l'inventaire dans la requête

### Erreur de compilation Go
**Solution :** Régénérer le code Ent
```bash
go generate ./ent
go mod tidy
```

### L'inventaire n'est pas sauvegardé
**Solution :** Vérifier que `container_details` est bien un champ JSONB dans PostgreSQL

## 📝 Exemples de Types de Contenants

| Type | Label | Icône |
|------|-------|-------|
| `sac` | Sac / Sacoche | ShoppingBag |
| `valise` | Valise / Bagage | Briefcase |
| `portefeuille` | Portefeuille | Wallet |
| `mallette` | Mallette professionnelle | Briefcase |
| `sac_dos` | Sac à dos | Backpack |

## 📝 Exemples de Catégories d'Items

| Catégorie | Label | Champs Spécifiques |
|-----------|-------|-------------------|
| `telephone` | Téléphone | brand, serial |
| `identite` | Identité | identityType, identityNumber, identityName |
| `carte` | Carte | cardType, cardBank, cardLast4 |
| `portefeuille` | Portefeuille | brand |
| `papiers` | Papiers | identityType, identityNumber |
| `ordinateur` | Ordinateur | brand, serial |
| `cles` | Clés | description |
| `argent` | Argent | description |

## ✅ Checklist de Validation

- [ ] Le schéma Ent est modifié
- [ ] Les types sont mis à jour
- [ ] Le repository est modifié
- [ ] Le service est modifié
- [ ] Le code Ent est régénéré (`go generate ./ent`)
- [ ] La migration est créée et appliquée
- [ ] Le backend compile sans erreur
- [ ] Test API : Créer un objet simple
- [ ] Test API : Créer un contenant avec inventaire
- [ ] Test API : Récupérer un objet avec inventaire
- [ ] Le frontend peut créer et afficher les objets avec inventaire

## 🎉 Félicitations !

Vous avez maintenant un système complet de gestion des objets perdus avec support des contenants et inventaires détaillés. Les agents peuvent désormais enregistrer précisément le contenu des sacs, valises et portefeuilles retrouvés avec tous leurs détails !

## 📞 Support

En cas de problème, vérifier :
1. Les logs du backend
2. La structure de la base de données
3. Les requêtes API avec les bons formats JSON
4. Que tous les fichiers ont bien été modifiés et sauvegardés
