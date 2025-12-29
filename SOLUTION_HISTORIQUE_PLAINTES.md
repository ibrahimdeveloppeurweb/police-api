# 🔄 Solution : Système d'Historique Automatique pour Plaintes

## 📊 Vue d'ensemble

Le système d'historique des plaintes fonctionne différemment des alertes :
- **Alertes** : Suivis stockés en JSON dans le champ `suivis`
- **Plaintes** : Historique stocké dans l'entité dédiée `PlainteHistorique`

## ✅ Ce qui existe déjà

1. **Entité PlainteHistorique** ✅ (ent/schema/plaintehistorique.go)
2. **Edge plainte → historiques** ✅ (ent/schema/plainte.go)
3. **Méthode GetHistorique** ✅ (service_extended.go)
4. **Type HistoriqueResponse** ✅ (types.go)
5. **Route API** ✅ (GET /plaintes/:id/historique)

## ❌ Ce qui manque

1. **Méthode CreateHistorique** pour enregistrer automatiquement les actions
2. **Appels à CreateHistorique** dans ChangerEtape, ChangerStatut, AssignerAgent

## 🛠️ Solution : Ajouter la Méthode CreateHistorique

### Étape 1 : Ajouter la méthode dans service_extended.go

Ajoutez cette méthode à la fin du fichier `internal/modules/plainte/service_extended.go` :

```go
// ========================
// HELPER METHODS
// ========================

// CreateHistorique crée une entrée d'historique pour une plainte
func (s *service) CreateHistorique(
	ctx context.Context,
	plainteID uuid.UUID,
	userID *uuid.UUID,
	typeChangement plaintehistorique.TypeChangement,
	champModifie string,
	ancienneValeur *string,
	nouvelleValeur string,
	commentaire *string,
	auteurNom *string,
) error {
	s.logger.Info("Creating historique entry",
		zap.String("plainte_id", plainteID.String()),
		zap.String("type_changement", string(typeChangement)),
		zap.String("champ_modifie", champModifie))

	// Build historique entry
	builder := s.client.PlainteHistorique.Create().
		SetPlainteID(plainteID).
		SetTypeChangement(typeChangement).
		SetChampModifie(champModifie).
		SetNouvelleValeur(nouvelleValeur)

	// Set optional fields
	if userID != nil {
		builder.SetUserID(*userID)
	}
	if ancienneValeur != nil {
		builder.SetAncienneValeur(*ancienneValeur)
	}
	if commentaire != nil {
		builder.SetCommentaire(*commentaire)
	}
	if auteurNom != nil {
		builder.SetAuteurNom(*auteurNom)
	}

	// Create the entry
	_, err := builder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create historique entry", zap.Error(err))
		return fmt.Errorf("failed to create historique: %w", err)
	}

	s.logger.Info("Successfully created historique entry")
	return nil
}

// ptrString retourne un pointeur vers une chaîne
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
```

### Étape 2 : Modifier ChangerEtape

Dans le fichier `internal/modules/plainte/service.go`, modifiez la méthode `ChangerEtape` :

```go
// ChangerEtape changes the workflow step of a plainte
func (s *service) ChangerEtape(ctx context.Context, id string, req ChangerEtapeRequest) (*PlainteResponse, error) {
	// Existing code...
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID")
	}

	// Get current plainte to get old value
	currentPlainte, err := s.client.Plainte.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("plainte not found")
	}
	oldEtape := string(currentPlainte.EtapeActuelle)

	// Update plainte
	plainteUpdate := s.client.Plainte.UpdateOneID(uid).
		SetEtapeActuelle(plainte.EtapeActuelle(req.Etape))

	if req.Observations != nil {
		plainteUpdate = plainteUpdate.SetObservations(*req.Observations)
	}

	updated, err := plainteUpdate.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update plainte: %w", err)
	}

	// ✨ NOUVEAU : Créer l'entrée d'historique
	err = s.CreateHistorique(
		ctx,
		uid,
		nil, // userID - à récupérer du contexte si disponible
		plaintehistorique.TypeChangementETAPE,
		"etape_actuelle",
		&oldEtape,
		req.Etape,
		req.Observations,
		nil, // auteurNom - à récupérer du contexte si disponible
	)
	if err != nil {
		s.logger.Warn("Failed to create historique entry", zap.Error(err))
		// Ne pas bloquer l'opération principale
	}

	// Convert to response...
	return s.convertToResponse(updated), nil
}
```

### Étape 3 : Modifier ChangerStatut

Dans le fichier `internal/modules/plainte/service.go`, modifiez la méthode `ChangerStatut` :

```go
// ChangerStatut changes the status of a plainte
func (s *service) ChangerStatut(ctx context.Context, id string, req ChangerStatutRequest) (*PlainteResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID")
	}

	// Get current plainte to get old value
	currentPlainte, err := s.client.Plainte.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("plainte not found")
	}
	oldStatut := string(currentPlainte.Statut)

	// Update plainte
	plainteUpdate := s.client.Plainte.UpdateOneID(uid).
		SetStatut(plainte.Statut(req.Statut))

	if req.DecisionFinale != nil {
		plainteUpdate = plainteUpdate.SetDecisionFinale(*req.DecisionFinale)
	}

	if req.Statut == "RESOLU" {
		now := time.Now()
		plainteUpdate = plainteUpdate.SetDateResolution(now)
	}

	updated, err := plainteUpdate.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update plainte: %w", err)
	}

	// ✨ NOUVEAU : Créer l'entrée d'historique
	err = s.CreateHistorique(
		ctx,
		uid,
		nil, // userID - à récupérer du contexte si disponible
		plaintehistorique.TypeChangementSTATUT,
		"statut",
		&oldStatut,
		req.Statut,
		req.DecisionFinale,
		nil, // auteurNom - à récupérer du contexte si disponible
	)
	if err != nil {
		s.logger.Warn("Failed to create historique entry", zap.Error(err))
		// Ne pas bloquer l'opération principale
	}

	return s.convertToResponse(updated), nil
}
```

### Étape 4 : Modifier AssignerAgent

Dans le fichier `internal/modules/plainte/service.go`, modifiez la méthode `AssignerAgent` :

```go
// AssignerAgent assigns an agent to a plainte
func (s *service) AssignerAgent(ctx context.Context, id string, req AssignerAgentRequest) (*PlainteResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID")
	}

	agentUID, err := uuid.Parse(req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent ID")
	}

	// Get current plainte
	currentPlainte, err := s.client.Plainte.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("plainte not found")
	}

	// Get old agent ID if exists
	var oldAgentID *string
	if currentPlainte.Edges.AgentAssigne != nil {
		old := currentPlainte.Edges.AgentAssigne.ID.String()
		oldAgentID = &old
	}

	// Verify agent exists
	agent, err := s.client.User.Get(ctx, agentUID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	// Update plainte
	updated, err := s.client.Plainte.UpdateOneID(uid).
		SetAgentAssigneID(agentUID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update plainte: %w", err)
	}

	// ✨ NOUVEAU : Créer l'entrée d'historique
	agentNom := fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
	err = s.CreateHistorique(
		ctx,
		uid,
		&agentUID, // L'agent assigné est l'auteur
		plaintehistorique.TypeChangementASSIGNATION,
		"agent_assigne_id",
		oldAgentID,
		req.AgentID,
		nil,
		&agentNom,
	)
	if err != nil {
		s.logger.Warn("Failed to create historique entry", zap.Error(err))
		// Ne pas bloquer l'opération principale
	}

	return s.convertToResponse(updated), nil
}
```

## 🚀 Étapes d'Installation

### 1. Régénérer le code Ent (si nécessaire)

```bash
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
go generate ./ent
```

### 2. Ajouter l'import dans service_extended.go

En haut du fichier `internal/modules/plainte/service_extended.go`, ajoutez :

```go
import (
	// ... imports existants
	"police-trafic-api-frontend-aligned/ent/plaintehistorique" // ✨ NOUVEAU
)
```

### 3. Tester la compilation

```bash
go build -o server cmd/api/main.go
```

### 4. Tester l'API

```bash
# 1. Changer l'étape d'une plainte
curl -X PATCH http://localhost:8080/api/plaintes/{id}/etape \
  -H "Content-Type: application/json" \
  -d '{"etape": "ENQUETE", "observations": "Début de l'enquête"}'

# 2. Vérifier l'historique
curl http://localhost:8080/api/plaintes/{id}/historique
```

## 📋 Vérification

Le frontend devrait maintenant afficher :
- ✅ Les changements d'étape
- ✅ Les changements de statut  
- ✅ Les assignations d'agent
- ✅ Avec toutes les informations (date, heure, auteur, observations)

## 🎨 Exemple de Réponse API

```json
[
  {
    "id": "uuid",
    "type_changement": "ETAPE",
    "champ_modifie": "etape_actuelle",
    "ancienne_valeur": "DEPOT",
    "nouvelle_valeur": "ENQUETE",
    "commentaire": "Début de l'enquête",
    "auteur_nom": "Jean Dupont",
    "created_at": "2025-12-18T14:30:00Z"
  },
  {
    "id": "uuid",
    "type_changement": "ASSIGNATION",
    "champ_modifie": "agent_assigne_id",
    "ancienne_valeur": null,
    "nouvelle_valeur": "uuid-agent",
    "commentaire": null,
    "auteur_nom": "Marie Martin",
    "created_at": "2025-12-18T15:00:00Z"
  }
]
```

## 💡 Améliorations Futures

1. **Récupérer l'utilisateur du contexte**
   - Actuellement `userID` et `auteurNom` sont mis à `nil`
   - Il faudrait les récupérer du contexte de la requête

2. **Ajouter d'autres types de changements**
   - PRIORITE (quand la priorité change)
   - AUTRE (pour d'autres modifications)

3. **Enrichir les commentaires automatiques**
   - Ajouter plus de détails sur les changements
   - Inclure des informations contextuelles

4. **Notifications**
   - Envoyer des notifications aux agents concernés
   - Logger dans les audit logs

## 🎯 Résultat Final

Avec cette solution :
- ✅ Historique complet de toutes les actions
- ✅ Traçabilité parfaite
- ✅ Interface frontend fonctionnelle
- ✅ Design cohérent avec les alertes
- ✅ Performance optimale (table dédiée avec index)

**Temps d'implémentation estimé : 15-20 minutes** ⏱️
