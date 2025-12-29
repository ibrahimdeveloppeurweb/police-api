package conducteur

import (
	"time"
)

// Request types

// CreateConducteurRequest represents the request to create a conducteur
type CreateConducteurRequest struct {
	Nom                 string     `json:"nom" validate:"required"`
	Prenom              string     `json:"prenom" validate:"required"`
	DateNaissance       time.Time  `json:"date_naissance" validate:"required"`
	LieuNaissance       *string    `json:"lieu_naissance,omitempty"`
	Adresse             *string    `json:"adresse,omitempty"`
	CodePostal          *string    `json:"code_postal,omitempty"`
	Ville               *string    `json:"ville,omitempty"`
	Telephone           *string    `json:"telephone,omitempty"`
	Email               *string    `json:"email,omitempty"`
	NumeroPermis        *string    `json:"numero_permis,omitempty"`
	PermisDelivreLe     *time.Time `json:"permis_delivre_le,omitempty"`
	PermisValideJusqu   *time.Time `json:"permis_valide_jusqu,omitempty"`
	CategoriesPermis    *string    `json:"categories_permis,omitempty"`
	PointsPermis        int        `json:"points_permis,omitempty"`
	Nationalite         string     `json:"nationalite,omitempty"`
}

// UpdateConducteurRequest represents the request to update a conducteur
type UpdateConducteurRequest struct {
	Nom                 *string    `json:"nom,omitempty"`
	Prenom              *string    `json:"prenom,omitempty"`
	DateNaissance       *time.Time `json:"date_naissance,omitempty"`
	LieuNaissance       *string    `json:"lieu_naissance,omitempty"`
	Adresse             *string    `json:"adresse,omitempty"`
	CodePostal          *string    `json:"code_postal,omitempty"`
	Ville               *string    `json:"ville,omitempty"`
	Telephone           *string    `json:"telephone,omitempty"`
	Email               *string    `json:"email,omitempty"`
	NumeroPermis        *string    `json:"numero_permis,omitempty"`
	PermisDelivreLe     *time.Time `json:"permis_delivre_le,omitempty"`
	PermisValideJusqu   *time.Time `json:"permis_valide_jusqu,omitempty"`
	CategoriesPermis    *string    `json:"categories_permis,omitempty"`
	PointsPermis        *int       `json:"points_permis,omitempty"`
	Nationalite         *string    `json:"nationalite,omitempty"`
	Active              *bool      `json:"active,omitempty"`
}

// ListConducteursRequest represents the request to list conducteurs
type ListConducteursRequest struct {
	Nom         *string `json:"nom,omitempty"`
	Prenom      *string `json:"prenom,omitempty"`
	Ville       *string `json:"ville,omitempty"`
	Nationalite *string `json:"nationalite,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}

// Response types

// ConducteurResponse represents a conducteur in API responses
type ConducteurResponse struct {
	ID                  string    `json:"id"`
	Nom                 string    `json:"nom"`
	Prenom              string    `json:"prenom"`
	DateNaissance       time.Time `json:"date_naissance"`
	LieuNaissance       string    `json:"lieu_naissance,omitempty"`
	Adresse             string    `json:"adresse,omitempty"`
	CodePostal          string    `json:"code_postal,omitempty"`
	Ville               string    `json:"ville,omitempty"`
	Telephone           string    `json:"telephone,omitempty"`
	Email               string    `json:"email,omitempty"`
	NumeroPermis        string    `json:"numero_permis,omitempty"`
	PermisDelivreLe     *time.Time `json:"permis_delivre_le,omitempty"`
	PermisValideJusqu   *time.Time `json:"permis_valide_jusqu,omitempty"`
	CategoriesPermis    string    `json:"categories_permis,omitempty"`
	PointsPermis        int       `json:"points_permis"`
	Nationalite         string    `json:"nationalite"`
	Active              bool      `json:"active"`
	NombreControles     int       `json:"nombre_controles"`
	NombreInfractions   int       `json:"nombre_infractions"`
	PermisValide        bool      `json:"permis_valide"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ListConducteursResponse represents the response for listing conducteurs
type ListConducteursResponse struct {
	Conducteurs []*ConducteurResponse `json:"conducteurs"`
	Total       int                   `json:"total"`
}

// SearchConducteursResponse represents the response for searching conducteurs
type SearchConducteursResponse struct {
	Query   string               `json:"query"`
	Results []*ConducteurResponse `json:"results"`
	Total   int                   `json:"total"`
}

// ConducteurStatisticsResponse represents statistics for a conducteur
type ConducteurStatisticsResponse struct {
	ConducteurID        string             `json:"conducteur_id"`
	NombreControles     int                `json:"nombre_controles"`
	NombreInfractions   int                `json:"nombre_infractions"`
	PointsRetires       int                `json:"points_retires"`
	PointsRestants      int                `json:"points_restants"`
	MontantAmendes      float64            `json:"montant_amendes"`
	DerniereInfraction  *time.Time         `json:"derniere_infraction,omitempty"`
	InfractionsParType  map[string]int     `json:"infractions_par_type"`
	PremierControle     *time.Time         `json:"premier_controle,omitempty"`
	DernierControle     *time.Time         `json:"dernier_controle,omitempty"`
}