package equipe

import "time"

// EquipeResponse represents an equipe in API responses
type EquipeResponse struct {
	ID             string                   `json:"id"`
	Nom            string                   `json:"nom"`
	Code           string                   `json:"code"`
	Zone           string                   `json:"zone,omitempty"`
	Description    string                   `json:"description,omitempty"`
	Active         bool                     `json:"active"`
	Commissariat   *CommissariatResponse    `json:"commissariat,omitempty"`
	ChefEquipe     *MembreResponse          `json:"chefEquipe,omitempty"`
	Membres        []MembreResponse         `json:"membres,omitempty"`
	Missions       []MissionSummaryResponse `json:"missions,omitempty"`
	NombreMembres  int                      `json:"nombreMembres"`
	MissionsActives int                     `json:"missionsActives"`
	CreatedAt      time.Time                `json:"createdAt"`
	UpdatedAt      time.Time                `json:"updatedAt"`
}

// CommissariatResponse represents a commissariat in equipe responses
type CommissariatResponse struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// MembreResponse represents a team member
type MembreResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
	Grade     string `json:"grade,omitempty"`
	Role      string `json:"role"`
}

// MissionSummaryResponse represents a mission summary
type MissionSummaryResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Titre     string    `json:"titre,omitempty"`
	DateDebut time.Time `json:"dateDebut"`
	Statut    string    `json:"statut"`
}

// CreateEquipeRequest represents request to create an equipe
type CreateEquipeRequest struct {
	Nom            string `json:"nom" validate:"required"`
	Code           string `json:"code" validate:"required"`
	Zone           string `json:"zone,omitempty"`
	Description    string `json:"description,omitempty"`
	CommissariatID string `json:"commissariatId,omitempty"`
}

// UpdateEquipeRequest represents request to update an equipe
type UpdateEquipeRequest struct {
	Nom         *string `json:"nom,omitempty"`
	Zone        *string `json:"zone,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// AddMembreRequest represents request to add a member
type AddMembreRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// SetChefEquipeRequest represents request to set team leader
type SetChefEquipeRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// ListEquipesFilters represents query filters for listing equipes
type ListEquipesFilters struct {
	CommissariatID string `query:"commissariatId"`
	Active         string `query:"active"`
	Search         string `query:"search"`
}
