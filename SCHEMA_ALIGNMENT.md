# Alignement Schéma Ent avec Frontend

## ✅ Schémas Créés

### 1. Control (Controle)
**Aligné avec** : `Controle` interface frontend

**Champs principaux** :
- `numero` - Numéro unique du contrôle
- `date` - Date du contrôle (séparée de l'heure)
- `heure` - Heure du contrôle (format HH:MM)
- `lieu` - Lieu du contrôle
- `status` - Status (en_cours, termine, avec_infractions, conforme)
- `vehicule_*` - Informations véhicule (immatriculation, marque, modele, couleur, type)
- `conducteur_*` - Informations conducteur (nom, prenoms, telephone)
- `permis_*` - Permis de conduire (numero, date_expiration, points_restants)
- `cni_*` - CNI (numero, date_expiration)
- `infractions` - JSON array des infractions constatées
- `montant_total` - Montant total des amendes
- `observations` - Observations de l'agent
- `photos` - JSON array des URLs de photos
- `pv_*` - Informations PV (numero, genere, date_generation)

### 2. ProcesVerbal (PV)
**Aligné avec** : `ProcesVerbal` interface frontend

**Champs principaux** :
- `numero` - Numéro unique du PV
- `controle_id` - ID du contrôle source
- `date_generation` - Date de génération
- `statut` - Statut (genere, notifie, paye, impaye, contentieux, annule)
- `infractions` - JSON array des infractions avec type, libelle, montant, points
- `montant_total` - Montant total
- `mode_paiement` - Mode de paiement (especes, mobile_money, virement, cheque)
- `date_paiement` - Date de paiement
- `reference_transaction` - Référence transaction
- `delai_paiement` - Délai de paiement

### 3. Commissariat
**Aligné avec** : `Commissariat` interface frontend

**Champs principaux** :
- `nom` - Nom du commissariat
- `localisation` - Adresse
- `latitude`, `longitude` - Coordonnées GPS
- `responsable_*` - Responsable (nom, grade, telephone)
- `agents_*` - Statistiques agents (total, presents, en_mission)
- `statistiques_*` - Statistiques (controles_jour, controles_semaine, controles_mois, revenus, taux_conformite)
- `status` - Status (actif, maintenance, urgence)

### 4. Agent
**Aligné avec** : `Agent` interface frontend

**Champs principaux** :
- `matricule` - Numéro matricule
- `nom`, `prenoms` - Nom et prénoms
- `grade` - Grade (Gardien de la Paix, Brigadier, etc.)
- `commissariat_id` - Commissariat assigné
- `telephone`, `email` - Contact
- `status` - Status (actif, repos, mission, formation, conge)
- `specialites` - JSON array des spécialités
- `date_recrutement` - Date de recrutement
- `derniere_activite` - Dernière activité

### 5. Alerte
**Aligné avec** : `Alerte` interface frontend

**Champs principaux** :
- `type` - Type (vehicule_vole, suspect_recherche, urgence_securite, alerte_generale, maintenance_systeme)
- `titre`, `message` - Titre et message
- `urgence` - Niveau urgence (faible, moyen, eleve, critique)
- `date` - Date de l'alerte
- `commissariat_id` - Commissariat concerné
- `vehicule_*` - Informations véhicule si applicable
- `suspect_*` - Informations suspect si applicable
- `status` - Status (active, resolue, archivee)
- `actions` - JSON array des actions

### 6. TypeInfraction
**Aligné avec** : `TypeInfraction` interface frontend

**Champs principaux** :
- `code` - Code infraction (DOC-001, VIT-001, etc.)
- `libelle` - Libellé
- `categorie` - Catégorie (Documents, Vitesse, Securite, Stationnement, Comportement, Vehicule)
- `gravite` - Gravité (1-5)
- `amende_min`, `amende_max` - Montants amende
- `devise` - Devise (FCFA)
- `points` - Points retirés
- `description` - Description
- `sanctions` - JSON array des sanctions
- `recidive_*` - Informations récidive

## 🔄 Différences avec l'ancien schéma

1. **Control** : 
   - Date/heure séparées au lieu d'un timestamp
   - Permis et CNI avec expiration
   - Photos en JSON array
   - Infractions en JSON array structuré
   - Status aligné avec frontend (en_cours, termine, etc.)

2. **ProcesVerbal** :
   - Statut aligné avec frontend (genere, notifie, paye, etc.)
   - Infractions avec points
   - Mode de paiement aligné

3. **Commissariat** :
   - Responsable intégré
   - Statistiques intégrées
   - Status aligné

4. **Agent** :
   - Grade aligné avec frontend
   - Status aligné
   - Spécialités en JSON array

## 📝 Prochaines Étapes

1. Générer le code Ent : `go generate ./ent`
2. Mettre à jour les repositories pour utiliser les nouveaux champs
3. Mettre à jour les services pour mapper correctement
4. Tester avec les données mock du frontend




