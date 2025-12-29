# 🚀 MODIFICATIONS MODULE PLAINTES - BACKEND API

## 📅 Date : 17 Décembre 2024

---

## 📋 RÉSUMÉ DES MODIFICATIONS

Ce document décrit toutes les modifications apportées au module **Plaintes** du backend pour supporter les nouvelles fonctionnalités du frontend.

---

## 🆕 NOUVEAUX ENDPOINTS AJOUTÉS

### 1. **Alertes Actives**
```
GET /api/plaintes/alertes?commissariat_id={id}
```
**Retourne :** Liste des alertes actives (SLA dépassé, sans action, etc.)

**Response :**
```json
[
  {
    "id": "uuid",
    "plainte_id": "uuid",
    "plainte_numero": "PLT-2024-001",
    "type_alerte": "SLA_DEPASSE",
    "message": "Le délai SLA a été dépassé de 5 jours",
    "niveau": "CRITICAL",
    "jours_retard": 5
  }
]
```

---

### 2. **Top Agents Performants**
```
GET /api/plaintes/top-agents?commissariat_id={id}
```
**Retourne :** Classement des agents les plus performants

**Response :**
```json
[
  {
    "id": "uuid",
    "nom": "Dupont",
    "prenom": "Jean",
    "matricule": "001",
    "plaintes_traitees": 47,
    "plaintes_resolues": 42,
    "score": 8.9,
    "delai_moyen": 3.5
  }
]
```

---

### 3. **Preuves d'une Plainte**

#### **GET - Liste des preuves**
```
GET /api/plaintes/:id/preuves
```
**Retourne :** Liste des preuves d'une plainte

**Response :**
```json
[
  {
    "id": "uuid",
    "numero_piece": "PCE-2024-001",
    "type": "MATERIELLE",
    "description": "Téléphone portable Samsung",
    "lieu_conservation": "Coffre 3",
    "date_collecte": "2024-12-15T10:00:00Z",
    "collecte_par": "Agent Dupont",
    "expertise_demandee": true,
    "expertise_type": "Analyse numérique",
    "statut": "COLLECTEE",
    "created_at": "2024-12-15T10:00:00Z"
  }
]
```

#### **POST - Ajouter une preuve**
```
POST /api/plaintes/:id/preuves
```
**Body :**
```json
{
  "numero_piece": "PCE-2024-001",
  "type": "MATERIELLE",
  "description": "Description de la preuve",
  "lieu_conservation": "Coffre 3",
  "date_collecte": "2024-12-15T10:00:00Z",
  "collecte_par": "Agent Dupont",
  "expertise_demandee": true,
  "expertise_type": "Analyse numérique"
}
```

**Types de preuve :** `MATERIELLE`, `NUMERIQUE`, `TESTIMONIALE`, `DOCUMENTAIRE`

---

### 4. **Actes d'Enquête**

#### **GET - Liste des actes**
```
GET /api/plaintes/:id/actes-enquete
```
**Retourne :** Liste des actes d'enquête d'une plainte

**Response :**
```json
[
  {
    "id": "uuid",
    "type": "AUDITION",
    "date": "2024-12-16T14:00:00Z",
    "heure": "14:00",
    "lieu": "Commissariat central, bureau 3",
    "officier_charge": "Agent Martin",
    "description": "Audition du plaignant",
    "pv_numero": "PV-2024-123",
    "created_at": "2024-12-16T14:00:00Z"
  }
]
```

#### **POST - Ajouter un acte**
```
POST /api/plaintes/:id/actes-enquete
```
**Body :**
```json
{
  "type": "AUDITION",
  "date": "2024-12-16T14:00:00Z",
  "heure": "14:00",
  "lieu": "Commissariat central",
  "officier_charge": "Agent Martin",
  "description": "Description de l'acte",
  "pv_numero": "PV-2024-123",
  "mandat_numero": "MAN-2024-456"
}
```

**Types d'acte :** `AUDITION`, `PERQUISITION`, `EXPERTISE`, `GARDE_A_VUE`, `CONFRONTATION`, `RECONSTITUTION`

---

### 5. **Timeline des Événements**

#### **GET - Liste des événements**
```
GET /api/plaintes/:id/timeline
```
**Retourne :** Timeline chronologique des événements

**Response :**
```json
[
  {
    "id": "uuid",
    "date": "2024-12-10T10:00:00Z",
    "heure": "10:00",
    "type": "DEPOT",
    "titre": "Dépôt de la plainte",
    "description": "Plainte déposée au commissariat",
    "acteur": "Agent d'accueil",
    "statut": "TERMINE",
    "created_at": "2024-12-10T10:00:00Z"
  }
]
```

#### **POST - Ajouter un événement**
```
POST /api/plaintes/:id/timeline
```
**Body :**
```json
{
  "date": "2024-12-17T15:00:00Z",
  "heure": "15:00",
  "type": "AUDITION",
  "titre": "Audition du témoin",
  "description": "Audition du témoin principal",
  "acteur": "Agent Dupont",
  "statut": "EN_COURS"
}
```

**Types d'événement :** `DEPOT`, `AUDITION`, `PERQUISITION`, `EXPERTISE`, `CONVOCATION`, `DECISION`, `AUTRE`

---

## 📂 FICHIERS MODIFIÉS

### 1. **controller.go**
- ✅ Ajout de 8 nouvelles routes
- ✅ Ajout de 8 nouveaux handlers

### 2. **types.go**
- ✅ Ajout de `AlerteResponse`
- ✅ Ajout de `TopAgentResponse`
- ✅ Ajout de `PreuveResponse` et `AddPreuveRequest`
- ✅ Ajout de `ActeEnqueteResponse` et `AddActeEnqueteRequest`
- ✅ Ajout de `TimelineEventResponse` et `AddTimelineEventRequest`

### 3. **service.go**
- ✅ Ajout de 8 nouvelles méthodes à l'interface Service

### 4. **service_extended.go** (NOUVEAU FICHIER)
- ✅ Implémentation des 8 nouvelles méthodes
- ✅ Données factices pour tests (à remplacer par vraie logique BDD)

---

## 🎯 STATUT D'IMPLÉMENTATION

### ✅ **Terminé**
- Routes API définies
- Handlers créés
- Types de requête/réponse définis
- Interface Service mise à jour
- Implémentations avec données factices

### 🔄 **À FAIRE (Prochaine étape)**
Pour rendre le backend complètement fonctionnel, il faut :

1. **Créer les schémas Ent** pour :
   - `Preuve`
   - `ActeEnquete`
   - `TimelineEvent`

2. **Implémenter la vraie logique** dans `service_extended.go` :
   - Remplacer les données factices
   - Ajouter les requêtes à la base de données
   - Gérer les relations entre entités

3. **Ajouter les validations** :
   - Validation des dates
   - Validation des types
   - Vérification des permissions

---

## 🧪 TESTS

### **Tester avec curl :**

```bash
# Alertes
curl http://localhost:8080/api/plaintes/alertes?commissariat_id=xxx

# Top Agents
curl http://localhost:8080/api/plaintes/top-agents?commissariat_id=xxx

# Preuves
curl http://localhost:8080/api/plaintes/{id}/preuves

# Ajouter une preuve
curl -X POST http://localhost:8080/api/plaintes/{id}/preuves \
  -H "Content-Type: application/json" \
  -d '{
    "numero_piece": "PCE-2024-001",
    "type": "MATERIELLE",
    "description": "Test",
    "date_collecte": "2024-12-17T10:00:00Z"
  }'

# Timeline
curl http://localhost:8080/api/plaintes/{id}/timeline

# Actes d'enquête
curl http://localhost:8080/api/plaintes/{id}/actes-enquete
```

---

## 📊 ARCHITECTURE

```
Module Plaintes Backend
├── controller.go          (Routes + Handlers)
├── service.go            (Interface Service)
├── service_extended.go   (Nouvelles méthodes) ⭐ NOUVEAU
├── types.go              (Types Request/Response)
└── module.go             (Module initialization)
```

---

## 🔗 INTÉGRATION FRONTEND

Le frontend peut maintenant appeler ces endpoints via :

```typescript
// Dans le frontend
import api from '@/lib/axios'

// Alertes
const alertes = await api.get('/plaintes/alertes')

// Top Agents
const agents = await api.get('/plaintes/top-agents')

// Preuves
const preuves = await api.get(`/plaintes/${id}/preuves`)

// Ajouter preuve
await api.post(`/plaintes/${id}/preuves`, data)

// Timeline
const timeline = await api.get(`/plaintes/${id}/timeline`)

// Actes
const actes = await api.get(`/plaintes/${id}/actes-enquete`)
```

---

## ✅ RÉSUMÉ

**8 nouveaux endpoints créés :**
1. ✅ GET /plaintes/alertes
2. ✅ GET /plaintes/top-agents
3. ✅ GET /plaintes/:id/preuves
4. ✅ POST /plaintes/:id/preuves
5. ✅ GET /plaintes/:id/actes-enquete
6. ✅ POST /plaintes/:id/actes-enquete
7. ✅ GET /plaintes/:id/timeline
8. ✅ POST /plaintes/:id/timeline

**Fichiers créés/modifiés :**
- ✅ controller.go (modifié)
- ✅ types.go (modifié)
- ✅ service.go (modifié)
- ✅ service_extended.go (créé) ⭐

---

**🎉 Le module Plaintes est maintenant prêt pour l'intégration frontend !**

Pour activer complètement, il faut simplement :
1. Redémarrer le serveur backend
2. Les endpoints retourneront des données factices
3. Plus tard, remplacer par la vraie logique BDD

---

**Auteur :** Assistant Claude  
**Date :** 17 Décembre 2024  
**Version :** 1.0
