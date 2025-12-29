package repository

import (
	"context"
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/ent/enttest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func TestVehiculeRepository_Create(t *testing.T) {
	// Créer un client Ent de test avec SQLite en mémoire
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Test de création
	input := &CreateVehiculeInput{
		ID:              "test-vehicule-1",
		Immatriculation: "AB123CD",
		Marque:         "Peugeot",
		Modele:         "208",
		TypeVehicule:   "VP",
	}

	// Créer le véhicule
	vehicule, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, vehicule)

	// Vérifications
	assert.Equal(t, input.ID, vehicule.ID)
	assert.Equal(t, "AB123CD", vehicule.Immatriculation) // Normalisé en majuscules
	assert.Equal(t, input.Marque, vehicule.Marque)
	assert.Equal(t, input.Modele, vehicule.Modele)
	assert.Equal(t, input.TypeVehicule, vehicule.TypeVehicule)
	assert.True(t, vehicule.Active) // Valeur par défaut
	assert.False(t, vehicule.CreatedAt.IsZero())
}

func TestVehiculeRepository_GetByID(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Créer un véhicule d'abord
	input := &CreateVehiculeInput{
		ID:              "test-vehicule-2",
		Immatriculation: "XY789ZW",
		Marque:         "Renault",
		Modele:         "Clio",
		TypeVehicule:   "VP",
	}

	created, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	// Récupérer par ID
	found, err := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Immatriculation, found.Immatriculation)
	assert.Equal(t, created.Marque, found.Marque)
}

func TestVehiculeRepository_GetByImmatriculation(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Créer un véhicule
	input := &CreateVehiculeInput{
		ID:              "test-vehicule-3",
		Immatriculation: "FR123GH",
		Marque:         "Citroën",
		Modele:         "C3",
		TypeVehicule:   "VP",
	}

	created, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	// Récupérer par immatriculation (avec différentes casses)
	testCases := []string{
		"FR123GH",
		"fr123gh",
		"Fr123gH",
	}

	for _, immat := range testCases {
		found, err := repo.GetByImmatriculation(context.Background(), immat)
		require.NoError(t, err, "Should find vehicule for immatriculation: %s", immat)
		assert.Equal(t, created.ID, found.ID)
	}
}

func TestVehiculeRepository_List(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Créer plusieurs véhicules
	vehicles := []*CreateVehiculeInput{
		{ID: "v1", Immatriculation: "AA111BB", Marque: "Peugeot", Modele: "208", TypeVehicule: "VP"},
		{ID: "v2", Immatriculation: "CC222DD", Marque: "Peugeot", Modele: "308", TypeVehicule: "VP"},
		{ID: "v3", Immatriculation: "EE333FF", Marque: "Renault", Modele: "Clio", TypeVehicule: "VP"},
	}

	for _, v := range vehicles {
		_, err := repo.Create(context.Background(), v)
		require.NoError(t, err)
	}

	// Test sans filtres
	all, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Test avec filtre marque
	filters := &VehiculeFilters{
		Marque: stringPtr("Peugeot"),
	}
	peugeots, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, peugeots, 2)

	// Test avec limite
	filters = &VehiculeFilters{
		Limit: 2,
	}
	limited, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, limited, 2)
}

func TestVehiculeRepository_Update(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Créer un véhicule
	input := &CreateVehiculeInput{
		ID:              "test-vehicule-4",
		Immatriculation: "UV456WX",
		Marque:         "Ford",
		Modele:         "Focus",
		TypeVehicule:   "VP",
	}

	created, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	// Mettre à jour
	updateInput := &UpdateVehiculeInput{
		Couleur:                        stringPtr("Rouge"),
		ProprietaireNom:                stringPtr("Dupont"),
		ProprietairePrenom:             stringPtr("Jean"),
	}

	updated, err := repo.Update(context.Background(), created.ID, updateInput)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "Rouge", updated.Couleur)
	assert.Equal(t, "Dupont", updated.ProprietaireNom)
	assert.Equal(t, "Jean", updated.ProprietairePrenom)
	
	// Vérifier que les autres champs n'ont pas changé
	assert.Equal(t, created.Marque, updated.Marque)
	assert.Equal(t, created.Modele, updated.Modele)
}

func TestVehiculeRepository_Search(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := NewVehiculeRepository(client, logger)

	// Créer des véhicules avec différents critères de recherche
	vehicles := []*CreateVehiculeInput{
		{ID: "s1", Immatriculation: "SEARCH01", Marque: "Tesla", Modele: "Model3", TypeVehicule: "VP"},
		{ID: "s2", Immatriculation: "FIND999", Marque: "BMW", Modele: "X3", TypeVehicule: "VP"},
		{ID: "s3", Immatriculation: "OTHER123", Marque: "Audi", Modele: "A4", TypeVehicule: "VP", ProprietaireNom: stringPtr("Martin")},
	}

	for _, v := range vehicles {
		_, err := repo.Create(context.Background(), v)
		require.NoError(t, err)
	}

	// Test recherche par immatriculation partielle
	results, err := repo.Search(context.Background(), "SEARCH")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "SEARCH01", results[0].Immatriculation)

	// Test recherche par marque
	results, err = repo.Search(context.Background(), "tesla")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Tesla", results[0].Marque)

	// Test recherche par nom propriétaire
	results, err = repo.Search(context.Background(), "martin")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Martin", results[0].ProprietaireNom)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}