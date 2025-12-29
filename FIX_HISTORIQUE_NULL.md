# 🔧 FIX: API /plaintes/:id/historique retourne null

## 🎯 Problème
L'endpoint `GET /api/plaintes/:id/historique` retourne `null` au lieu d'un tableau vide `[]`.

## ✅ Solution Rapide (Base de données uniquement)

### Option 1: Exécuter le script SQL

```bash
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Se connecter à PostgreSQL et exécuter le script
psql -U votre_user -d votre_database -f create_historique_table.sql
```

### Option 2: SQL Manuel

```sql
-- Créer la table
CREATE TABLE historique_action_plaintes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plainte_id UUID NOT NULL REFERENCES plaintes(id) ON DELETE CASCADE,
    type_action VARCHAR(50) NOT NULL,
    ancienne_valeur VARCHAR(255),
    nouvelle_valeur VARCHAR(255) NOT NULL,
    observations TEXT,
    effectue_par UUID REFERENCES users(id) ON DELETE SET NULL,
    effectue_par_nom VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_historique_plainte_id ON historique_action_plaintes(plainte_id);
CREATE INDEX idx_historique_created_at ON historique_action_plaintes(created_at DESC);

-- Ajouter champs manquants dans plaintes
ALTER TABLE plaintes ADD COLUMN IF NOT EXISTS nombre_convocations INTEGER DEFAULT 0;
ALTER TABLE plaintes ADD COLUMN IF NOT EXISTS decision_finale TEXT;
```

## 🚀 Solution Complète (Backend + BDD)

### Étape 1: Préparer les fichiers Ent

Les fichiers suivants ont déjà été créés :
- ✅ `ent/schema/historique_action_plainte.go`
- ✅ `ent/schema/plainte.go` (edge ajouté)

### Étape 2: Générer les entités

```bash
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
go generate ./ent
```

### Étape 3: Ajouter les types

Dans `internal/modules/plainte/types.go`, ajoutez à la fin :

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

### Étape 4: Ajouter les méthodes service

Dans `internal/modules/plainte/service_extended.go`, ajoutez à la fin :

```go
// GetHistoriqueActions returns historique actions for a plainte
func (s *service) GetHistoriqueActions(ctx context.Context, plainteID string) ([]HistoriqueActionResponse, error) {
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	actions, err := s.client.HistoriqueActionPlainte.Query().
		Where(historiqueactionplainte.PlainteIDEQ(uid)).
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query historique actions: %w", err)
	}

	var responses []HistoriqueActionResponse
	for _, action := range actions {
		resp := HistoriqueActionResponse{
			ID:              action.ID.String(),
			PlainteID:       action.PlainteID.String(),
			TypeAction:      action.TypeAction,
			AncienneValeur:  ptrString(action.AncienneValeur),
			NouvelleValeur:  action.NouvelleValeur,
			Observations:    ptrString(action.Observations),
			EffectueParNom:  ptrString(action.EffectueParNom),
			CreatedAt:       action.CreatedAt,
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *service) CreateHistoriqueAction(ctx context.Context, req CreateHistoriqueActionRequest) error {
	plainteUID, err := uuid.Parse(req.PlainteID)
	if err != nil {
		return fmt.Errorf("invalid plainte ID: %w", err)
	}

	builder := s.client.HistoriqueActionPlainte.Create().
		SetPlainteID(plainteUID).
		SetTypeAction(req.TypeAction).
		SetNouvelleValeur(req.NouvelleValeur)

	if req.AncienneValeur != nil {
		builder.SetAncienneValeur(*req.AncienneValeur)
	}
	if req.Observations != nil {
		builder.SetObservations(*req.Observations)
	}
	if req.EffectueParNom != nil {
		builder.SetEffectueParNom(*req.EffectueParNom)
	}

	_, err = builder.Save(ctx)
	return err
}
```

### Étape 5: Modifier le contrôleur

Dans `internal/modules/plainte/controller.go`, remplacez `GetHistorique` par :

```go
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

### Étape 6: Enregistrer l'historique automatiquement

Dans chaque méthode de modification (ChangerEtape, ChangerStatut, AssignerAgent), ajoutez :

```go
// À la fin de ChangerEtape
_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
	PlainteID:      plainteID,
	TypeAction:     "CHANGEMENT_ETAPE",
	AncienneValeur: ptrString(string(ancienneEtape)),
	NouvelleValeur: string(nouvelleEtape),
	Observations:   req.Observations,
})

// À la fin de ChangerStatut
typeAction := "CHANGEMENT_STATUT"
if req.Statut == "CONVOCATION" {
	typeAction = "CONVOCATION"
}
_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
	PlainteID:      plainteID,
	TypeAction:     typeAction,
	AncienneValeur: ptrString(string(ancienStatut)),
	NouvelleValeur: string(nouveauStatut),
	Observations:   req.DecisionFinale,
})

// À la fin de AssignerAgent
_ = s.CreateHistoriqueAction(ctx, CreateHistoriqueActionRequest{
	PlainteID:      plainteID,
	TypeAction:     "ASSIGNATION_AGENT",
	NouvelleValeur: fmt.Sprintf("%s %s", agent.Nom, agent.Prenom),
})
```

### Étape 7: Compiler et redémarrer

```bash
go build -o server cmd/api/main.go
./server
```

## 🧪 Tests

```bash
# Test 1: Vérifier que l'endpoint retourne un tableau vide au lieu de null
curl http://localhost:8080/api/plaintes/VOTRE-UUID/historique

# Résultat attendu: []
# Au lieu de: null

# Test 2: Après avoir changé l'étape d'une plainte
# L'historique devrait contenir une entrée
```

## 📝 Checklist

- [ ] Table `historique_action_plaintes` créée dans PostgreSQL
- [ ] Champs `nombre_convocations` et `decision_finale` ajoutés à la table `plaintes`
- [ ] Entités Ent générées avec `go generate ./ent`
- [ ] Types ajoutés dans `types.go`
- [ ] Méthodes ajoutées dans `service_extended.go`
- [ ] Contrôleur modifié dans `controller.go`
- [ ] Enregistrements automatiques ajoutés dans les 3 méthodes
- [ ] Backend compilé et redémarré
- [ ] Endpoint testé et retourne `[]` au lieu de `null`

## 🆘 Aide Rapide

Si vous voulez juste que l'API retourne `[]` au lieu de `null` MAINTENANT :

1. **Solution ultra-rapide** : Créez juste la table avec le SQL ci-dessus
2. **Redémarrez le backend**
3. L'endpoint devrait retourner `[]`

Les autres modifications permettront d'enregistrer automatiquement l'historique quand vous faites des actions.

## 📚 Documentation Complète

Voir `GUIDE_HISTORIQUE_ACTIONS_BACKEND.md` pour les détails complets.
