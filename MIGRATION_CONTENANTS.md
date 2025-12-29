# Migration des Objets Perdus en Mode Contenant

## Problème

Les objets perdus créés avant la mise en place du système de contenants (Sac, Valise, Portefeuille, etc.) ont `isContainer = false` dans la base de données. Ils ne bénéficient donc pas de l'affichage amélioré avec inventaire.

## Solution

Deux options pour migrer les objets existants :

### Option 1 : Migration SQL directe (Recommandé pour les développeurs)

```bash
# Se connecter à la base de données PostgreSQL
psql -h localhost -U votre_utilisateur -d police_trafic

# Exécuter le script SQL
\i scripts/migrate_containers.sql
```

### Option 2 : Migration via script Node.js (Plus sûr)

```bash
# Depuis le dossier police-trafic-api-frontend-aligned

# 1. S'assurer que le backend est démarré
./server

# 2. Dans un autre terminal, exécuter le script
cd scripts
node migrate-containers-to-new-format.js
```

## Vérification

Après la migration, vérifiez dans l'interface web :

1. Ouvrir un objet perdu de type "Sac / Sacoche", "Valise", etc.
2. Vous devriez voir :
   - Badge "Contenant avec inventaire" (violet)
   - Section "Description du contenant" au lieu de "Description de l'objet"
   - Section "Inventaire du contenant" (vide par défaut)

## Créer un nouvel objet avec inventaire

Pour tester complètement la nouvelle fonctionnalité :

1. Aller dans "Objets Perdus" → "Nouveau"
2. Sélectionner "Oui, c'est un contenant" 
3. Choisir le type de contenant (Sac, Valise, etc.)
4. Remplir les détails du contenant
5. Ajouter des objets à l'inventaire via le bouton "+"
6. Enregistrer

Vous verrez alors l'affichage complet avec l'inventaire des objets contenus.

## Rollback (Annuler la migration)

Si besoin d'annuler la migration :

```sql
UPDATE objets_perdus
SET 
    is_container = false,
    container_details = NULL,
    updated_at = NOW()
WHERE is_container = true;
```

## Notes

- Les objets migrés auront un inventaire vide par défaut
- La description originale est déplacée dans "signesDistinctifs" du contenant
- La couleur est préservée dans les détails du contenant
- Le type de contenant est déduit automatiquement du `typeObjet`
