package vehicule

import (
	"time"
)

// Request types

// CreateVehiculeRequest represents the request to create a vehicule
type CreateVehiculeRequest struct {
	Immatriculation                string  `json:"immatriculation" validate:"required"`
	Marque                         string  `json:"marque" validate:"required"`
	Modele                         string  `json:"modele" validate:"required"`
	Couleur                        *string `json:"couleur,omitempty"`
	TypeVehicule                   string  `json:"type_vehicule,omitempty"`
	Energie                        *string `json:"energie,omitempty"`
	NumeroChassis                  *string `json:"numero_chassis,omitempty"`
	ProprietaireNom                *string `json:"proprietaire_nom,omitempty"`
	ProprietairePrenom             *string `json:"proprietaire_prenom,omitempty"`
	ProprietaireAdresse            *string `json:"proprietaire_adresse,omitempty"`
	AssuranceCompagnie             *string `json:"assurance_compagnie,omitempty"`
	AssuranceNumero                *string `json:"assurance_numero,omitempty"`
}

// UpdateVehiculeRequest represents the request to update a vehicule
type UpdateVehiculeRequest struct {
	Marque                         *string `json:"marque,omitempty"`
	Modele                         *string `json:"modele,omitempty"`
	Couleur                        *string `json:"couleur,omitempty"`
	TypeVehicule                   *string `json:"type_vehicule,omitempty"`
	Energie                        *string `json:"energie,omitempty"`
	NumeroChassis                  *string `json:"numero_chassis,omitempty"`
	ProprietaireNom                *string `json:"proprietaire_nom,omitempty"`
	ProprietairePrenom             *string `json:"proprietaire_prenom,omitempty"`
	ProprietaireAdresse            *string `json:"proprietaire_adresse,omitempty"`
	AssuranceCompagnie             *string `json:"assurance_compagnie,omitempty"`
	AssuranceNumero                *string `json:"assurance_numero,omitempty"`
	Active                         *bool   `json:"active,omitempty"`
}

// ListVehiculesRequest represents the request to list vehicules
type ListVehiculesRequest struct {
	Marque          *string `json:"marque,omitempty"`
	Modele          *string `json:"modele,omitempty"`
	TypeVehicule    *string `json:"type_vehicule,omitempty"`
	Active          *bool   `json:"active,omitempty"`
	ProprietaireNom *string `json:"proprietaire_nom,omitempty"`
	Limit           int     `json:"limit,omitempty"`
	Offset          int     `json:"offset,omitempty"`
}

// Response types

// VehiculeResponse represents a vehicule in API responses
type VehiculeResponse struct {
	ID                             string    `json:"id"`
	Immatriculation                string    `json:"immatriculation"`
	Marque                         string    `json:"marque"`
	Modele                         string    `json:"modele"`
	Couleur                        string    `json:"couleur,omitempty"`
	TypeVehicule                   string    `json:"type_vehicule"`
	Energie                        string    `json:"energie,omitempty"`
	NumeroChassis                  string    `json:"numero_chassis,omitempty"`
	ProprietaireNom                string    `json:"proprietaire_nom,omitempty"`
	ProprietairePrenom             string    `json:"proprietaire_prenom,omitempty"`
	ProprietaireAdresse            string    `json:"proprietaire_adresse,omitempty"`
	AssuranceCompagnie             string    `json:"assurance_compagnie,omitempty"`
	AssuranceNumero                string    `json:"assurance_numero,omitempty"`
	Active                         bool      `json:"active"`
	NombreControles                int       `json:"nombre_controles"`
	NombreInfractions              int       `json:"nombre_infractions"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

// ListVehiculesResponse represents the response for listing vehicules
type ListVehiculesResponse struct {
	Vehicules []*VehiculeResponse `json:"vehicules"`
	Total     int                 `json:"total"`
}

// SearchVehiculesResponse represents the response for searching vehicules
type SearchVehiculesResponse struct {
	Query   string               `json:"query"`
	Results []*VehiculeResponse `json:"results"`
	Total   int                  `json:"total"`
}