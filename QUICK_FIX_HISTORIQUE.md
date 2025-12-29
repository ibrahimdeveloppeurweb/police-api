# ⚡ Quick Fix - Système d'Historique des Plaintes

## 🎯 Problème
Le frontend des plaintes a un système de suivi (comme les alertes), mais le backend n'enregistre pas automatiquement les actions dans l'historique.

## ✅ Ce qui existe déjà
- Frontend avec onglet "Suivi" ✅
- Entité `PlainteHistorique` ✅
- Méthode `GetHistorique` ✅
- Route API `/plaintes/:id/historique` ✅

## ❌ Ce qui manque
- Enregistrement automatique des actions (changement d'étape, statut, assignation)

## 🛠️ Solution Rapide

### 1. Ajouter la méthode CreateHistorique

Ajoutez à la fin de `internal/modules/plainte/service_extended.go` :

```go
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
		zap.String("type_changement", string(typeChangement)))

	builder := s.client.PlainteHistorique.Create().
		SetPlainteID(plainteID).
		SetTypeChangement(typeChangement).
		SetChampModifie(champModifie).
		SetNouvelleValeur(nouvelleValeur)

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

	_, err := builder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create historique entry", zap.Error(err))
		return fmt.Errorf("failed to create historique: %w", err)
	}

	return nil
}
```

### 2. Ajouter l'import nécessaire

En haut de `service_extended.go` :

```go
import (
	"police-trafic-api-frontend-aligned/ent/plaintehistorique"
	// ... autres imports
)
```

### 3. Appeler CreateHistorique dans les méthodes

Dans `internal/modules/plainte/service.go` :

#### ChangerEtape :
```go
// Après l'update de la plainte
err = s.CreateHistorique(
	ctx,
	uid,
	nil,
	plaintehistorique.TypeChangementETAPE,
	"etape_actuelle",
	&oldEtape,
	req.Etape,
	req.Observations,
	nil,
)
```

#### ChangerStatut :
```go
// Après l'update de la plainte
err = s.CreateHistorique(
	ctx,
	uid,
	nil,
	plaintehistorique.TypeChangementSTATUT,
	"statut",
	&oldStatut,
	req.Statut,
	req.DecisionFinale,
	nil,
)
```

#### AssignerAgent :
```go
// Après l'update de la plainte
agentNom := fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
err = s.CreateHistorique(
	ctx,
	uid,
	&agentUID,
	plaintehistorique.TypeChangementASSIGNATION,
	"agent_assigne_id",
	oldAgentID,
	req.AgentID,
	nil,
	&agentNom,
)
```

## 🚀 Test Rapide

```bash
# 1. Compiler
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
go build -o server cmd/api/main.go

# 2. Lancer le serveur
./server

# 3. Tester (dans un autre terminal)
# Changer l'étape
curl -X PATCH http://localhost:8080/api/plaintes/{id}/etape \
  -H "Content-Type: application/json" \
  -d '{"etape": "ENQUETE"}'

# Voir l'historique
curl http://localhost:8080/api/plaintes/{id}/historique
```

## ✅ Résultat

Après ces modifications :
- ✅ Chaque changement d'étape est enregistré
- ✅ Chaque changement de statut est enregistré
- ✅ Chaque assignation d'agent est enregistrée
- ✅ Le frontend affiche tout dans l'onglet "Suivi"

## 📄 Documentation Complète

Voir `SOLUTION_HISTORIQUE_PLAINTES.md` pour tous les détails.

**Temps d'implémentation : 10-15 minutes** ⏱️
