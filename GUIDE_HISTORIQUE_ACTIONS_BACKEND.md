// GUIDE DE MISE À JOUR BACKEND POUR HISTORIQUE DES ACTIONS

## ÉTAPE 1 : Générer les entités Ent

Exécutez le script de génération :
```bash
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
chmod +x generer_historique.sh
./generer_historique.sh
```

## ÉTAPE 2 : Ajouter les méthodes dans service_extended.go

Ajoutez cette fonction à la fin du fichier `internal/modules/plainte/service_extended.go` :

```go
// ========================
// HISTORIQUE ACTIONS IMPLEMENTATION
// ========================

// GetHistoriqueActions returns historique actions for a plainte from database
func (s *service) GetHistoriqueActions(ctx context.Context, plainteID string) ([]HistoriqueActionResponse, error) {
	s.logger.Info("Getting historique actions from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query historique actions from database
	actions, err := s.client.HistoriqueActionPlainte.Query().
		Where(historiqueactionplainte.PlainteIDEQ(uid)).
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query historique actions", zap.Error(err))
		return nil, fmt.Errorf("failed to query historique actions: %w", err)
	}

	// Convert to response format
	var responses []HistoriqueActionResponse
	for _, action := range actions {
		resp := HistoriqueActionResponse{
			ID:              action.ID.String(),
			PlainteID:       action.PlainteID.String(),
			TypeAction:      action.TypeAction,
			AncienneValeur:  ptrString(action.AncienneValeur),
			NouvelleValeur:  action.NouvelleValeur,
			Observations:    ptrString(action.Observations),
			EffectuePar:     ptrUUIDString(action.EffectuePar),
			EffectueParNom:  ptrString(action.EffectueParNom),
			CreatedAt:       action.CreatedAt,
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched historique actions",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// CreateHistoriqueAction creates a new historique action entry
func (s *service) CreateHistoriqueAction(ctx context.Context, req CreateHistoriqueActionRequest) error {
	s.logger.Info("Creating historique action",
		zap.String("plainte_id", req.PlainteID),
		zap.String("type_action", req.TypeAction))

	// Convert IDs to UUID
	plainteUID, err := uuid.Parse(req.PlainteID)
	if err != nil {
		return fmt.Errorf("invalid plainte ID: %w", err)
	}

	builder := s.client.HistoriqueActionPlainte.Create().
		SetPlainteID(plainteUID).
		SetTypeAction(req.TypeAction).
		SetNouvelleValeur(req.NouvelleValeur)

	// Set optional fields
	if req.AncienneValeur != nil {
		builder.SetAncienneValeur(*req.AncienneValeur)
	}
	if req.Observations != nil {
		builder.SetObservations(*req.Observations)
	}
	if req.EffectuePar != nil {
		effectueParUID, err := uuid.Parse(*req.EffectuePar)
		if err == nil {
			builder.SetEffectuePar(effectueParUID)
		}
	}
	if req.EffectueParNom != nil {
		builder.SetEffectueParNom(*req.EffectueParNom)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create historique action", zap.Error(err))
		return fmt.Errorf("failed to create historique action: %w", err)
	}

	s.logger.Info("Successfully created historique action")
	return nil
}

// Helper function for optional UUID to string
func ptrUUIDString(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}
```

## ÉTAPE 3 : Ajouter les types dans types.go

Ajoutez ces types dans `internal/modules/plainte/types.go` :

```go
// HistoriqueActionResponse represents a historique action response
type HistoriqueActionResponse struct {
	ID              string     `json:"id"`
	PlainteID       string     `json:"plainte_id"`
	TypeAction      string     `json:"type_action"`
	AncienneValeur  *string    `json:"ancienne_valeur,omitempty"`
	NouvelleValeur  string     `json:"nouvelle_valeur"`
	Observations    *string    `json:"observations,omitempty"`
	EffectuePar     *string    `json:"effectue_par,omitempty"`
	EffectueParNom  *string    `json:"effectue_par_nom,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateHistoriqueActionRequest represents a request to create a historique action
type CreateHistoriqueActionRequest struct {
	PlainteID       string  `json:"plainte_id"`
	TypeAction      string  `json:"type_action"`
	AncienneValeur  *string `json:"ancienne_valeur,omitempty"`
	NouvelleValeur  string  `json:"nouvelle_valeur"`
	Observations    *string `json:"observations,omitempty"`
	EffectuePar     *string `json:"effectue_par,omitempty"`
	EffectueParNom  *string `json:"effectue_par_nom,omitempty"`
}
```

## ÉTAPE 4 : Modifier le contrôleur

Dans `internal/modules/plainte/controller.go`, remplacez la méthode GetHistorique existante par :

```go
// GetHistorique returns historique actions for a plainte
func (c *Controller) GetHistorique(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	historique, err := c.service.GetHistoriqueActions(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, historique)
}
```

## ÉTAPE 5 : Modifier les endpoints existants pour enregistrer l'historique

Dans `service_extended.go`, modifiez les méthodes suivantes :

### ChangerEtape
```go
func (s *service) ChangerEtape(ctx context.Context, plainteID string, req ChangerEtapeRequest) (*PlainteResponse, error) {
	// ... code existant ...
	
	// AJOUTER AVANT LE RETURN:
	// Enregistrer dans l'historique
	_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
		PlainteID:      plainteID,
		TypeAction:     "CHANGEMENT_ETAPE",
		AncienneValeur: ptrString(string(oldEtape)),
		NouvelleValeur: string(req.Etape),
		Observations:   req.Observations,
		// EffectuePar et EffectueParNom à récupérer du contexte auth
	})
	
	return response, nil
}
```

### ChangerStatut
```go
func (s *service) ChangerStatut(ctx context.Context, plainteID string, req ChangerStatutRequest) (*PlainteResponse, error) {
	// ... code existant ...
	
	// AJOUTER AVANT LE RETURN:
	typeAction := "CHANGEMENT_STATUT"
	if req.Statut == "CONVOCATION" {
		typeAction = "CONVOCATION"
	}
	
	_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
		PlainteID:      plainteID,
		TypeAction:     typeAction,
		AncienneValeur: ptrString(string(oldStatut)),
		NouvelleValeur: string(req.Statut),
		Observations:   req.DecisionFinale,
	})
	
	return response, nil
}
```

### AssignerAgent
```go
func (s *service) AssignerAgent(ctx context.Context, plainteID string, req AssignerAgentRequest) (*PlainteResponse, error) {
	// ... code existant après avoir récupéré l'agent ...
	
	// AJOUTER AVANT LE RETURN:
	nouvelleValeur := fmt.Sprintf("%s %s", agent.Nom, agent.Prenom)
	_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
		PlainteID:      plainteID,
		TypeAction:     "ASSIGNATION_AGENT",
		NouvelleValeur: nouvelleValeur,
		Observations:   req.Observations,
	})
	
	return response, nil
}
```

## ÉTAPE 6 : Redémarrer le backend

```bash
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
make restart
# ou
./restart-backend.sh
```

## ÉTAPE 7 : Tester

```bash
# Tester l'endpoint historique
curl http://localhost:8080/api/plaintes/{ID}/historique

# Devrait retourner un tableau vide [] au lieu de null
```

## RÉSUMÉ DES FICHIERS À MODIFIER

1. ✅ `ent/schema/historique_action_plainte.go` - CRÉÉ
2. ✅ `ent/schema/plainte.go` - MODIFIÉ (edge ajouté)
3. ⏳ Générer les entités avec `./generer_historique.sh`
4. ⏳ `internal/modules/plainte/types.go` - AJOUTER les types
5. ⏳ `internal/modules/plainte/service_extended.go` - AJOUTER les méthodes
6. ⏳ `internal/modules/plainte/controller.go` - MODIFIER GetHistorique
7. ⏳ Modifier ChangerEtape, ChangerStatut, AssignerAgent
8. ⏳ Redémarrer le backend
