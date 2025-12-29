package vehicule

import (
	"context"
	"testing"

	"police-trafic-api-frontend-aligned/ent/enttest"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func TestVehiculeService_Create(t *testing.T) {
	// Setup
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Test création avec validation métier
	input := &CreateVehiculeRequest{
		Immatriculation: "ab-123-cd", // Sera normalisé
		Marque:         "Peugeot",
		Modele:         "208",
		TypeVehicule:   "VP",
		Couleur:        stringPtr("Blanc"),
	}

	result, err := service.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Vérifications
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "AB123CD", result.Immatriculation) // Normalisé
	assert.Equal(t, input.Marque, result.Marque)
	assert.Equal(t, input.Modele, result.Modele)
	assert.Equal(t, "Blanc", result.Couleur)
	assert.True(t, result.Active)
	assert.Equal(t, 0, result.NombreControles) // Pas de contrôles encore
}

func TestVehiculeService_Create_Validation(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Test validation: immatriculation manquante
	input := &CreateVehiculeRequest{
		Marque:       "Peugeot",
		Modele:       "208",
		TypeVehicule: "VP",
	}

	_, err := service.Create(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "immatriculation is required")

	// Test validation: marque manquante
	input = &CreateVehiculeRequest{
		Immatriculation: "AB123CD",
		Modele:          "208",
		TypeVehicule:    "VP",
	}

	_, err = service.Create(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marque is required")
}

func TestVehiculeService_Create_UniqueImmatriculation(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Créer un premier véhicule
	input1 := &CreateVehiculeRequest{
		Immatriculation: "UNIQUE123",
		Marque:         "Peugeot",
		Modele:         "208",
	}

	_, err := service.Create(context.Background(), input1)
	require.NoError(t, err)

	// Tentative de création avec la même immatriculation
	input2 := &CreateVehiculeRequest{
		Immatriculation: "unique123", // Même après normalisation
		Marque:         "Renault",
		Modele:         "Clio",
	}

	_, err = service.Create(context.Background(), input2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestVehiculeService_GetByImmatriculation(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Créer un véhicule
	input := &CreateVehiculeRequest{
		Immatriculation: "FIND-ME-123",
		Marque:         "Tesla",
		Modele:         "Model 3",
	}

	created, err := service.Create(context.Background(), input)
	require.NoError(t, err)

	// Rechercher avec différents formats
	testCases := []string{
		"FIND-ME-123",
		"find-me-123",
		"FINDME123",
		"findme123",
	}

	for _, immat := range testCases {
		found, err := service.GetByImmatriculation(context.Background(), immat)
		require.NoError(t, err, "Should find vehicule for: %s", immat)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "FINDME123", found.Immatriculation) // Toujours normalisé
	}
}

func TestVehiculeService_Search(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Créer plusieurs véhicules
	vehicles := []*CreateVehiculeRequest{
		{Immatriculation: "SEARCH001", Marque: "Tesla", Modele: "Model S"},
		{Immatriculation: "FIND999", Marque: "BMW", Modele: "X5"},
		{Immatriculation: "OTHER123", Marque: "Audi", Modele: "Q7", ProprietaireNom: stringPtr("Durand")},
	}

	for _, v := range vehicles {
		_, err := service.Create(context.Background(), v)
		require.NoError(t, err)
	}

	// Test recherche par différents critères
	testCases := []struct {
		query    string
		expected int
	}{
		{"tesla", 1},
		{"SEARCH", 1},
		{"999", 1},
		{"durand", 1},
		{"unknown", 0},
		{"", 0}, // Requête vide
	}

	for _, tc := range testCases {
		results, err := service.Search(context.Background(), tc.query)
		require.NoError(t, err)
		assert.Len(t, results.Results, tc.expected, "Query: %s", tc.query)
		assert.Equal(t, tc.query, results.Query)
		assert.Equal(t, tc.expected, results.Total)
	}
}

func TestVehiculeService_List_WithFilters(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	repo := repository.NewVehiculeRepository(client, logger)
	service := NewService(repo, logger)

	// Créer plusieurs véhicules
	vehicles := []*CreateVehiculeRequest{
		{Immatriculation: "FILTER01", Marque: "Peugeot", Modele: "208", TypeVehicule: "VP"},
		{Immatriculation: "FILTER02", Marque: "Peugeot", Modele: "308", TypeVehicule: "VP"},
		{Immatriculation: "FILTER03", Marque: "Renault", Modele: "Clio", TypeVehicule: "VP"},
		{Immatriculation: "FILTER04", Marque: "Mercedes", Modele: "Sprinter", TypeVehicule: "PL"},
	}

	for _, v := range vehicles {
		_, err := service.Create(context.Background(), v)
		require.NoError(t, err)
	}

	// Test sans filtres
	results, err := service.List(context.Background(), &ListVehiculesRequest{})
	require.NoError(t, err)
	assert.Len(t, results.Vehicules, 4)
	assert.Equal(t, 4, results.Total)

	// Test filtre par marque
	results, err = service.List(context.Background(), &ListVehiculesRequest{
		Marque: stringPtr("Peugeot"),
	})
	require.NoError(t, err)
	assert.Len(t, results.Vehicules, 2)

	// Test filtre par type
	results, err = service.List(context.Background(), &ListVehiculesRequest{
		TypeVehicule: stringPtr("PL"),
	})
	require.NoError(t, err)
	assert.Len(t, results.Vehicules, 1)
	assert.Equal(t, "Mercedes", results.Vehicules[0].Marque)

	// Test avec limite
	results, err = service.List(context.Background(), &ListVehiculesRequest{
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, results.Vehicules, 2)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}