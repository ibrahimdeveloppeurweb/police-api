package observation

import "time"

// ObservationResponse represents an observation in API responses
type ObservationResponse struct {
	ID           string         `json:"id"`
	Contenu      string         `json:"contenu"`
	Type         string         `json:"type"`
	Categorie    string         `json:"categorie,omitempty"`
	VisibleAgent bool           `json:"visibleAgent"`
	Agent        *AgentResponse `json:"agent,omitempty"`
	Auteur       *AgentResponse `json:"auteur,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// AgentResponse represents an agent in observation responses
type AgentResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
	Grade     string `json:"grade,omitempty"`
}

// CreateObservationRequest represents request to create an observation
type CreateObservationRequest struct {
	Contenu      string `json:"contenu" validate:"required"`
	Type         string `json:"type" validate:"required"`
	Categorie    string `json:"categorie,omitempty"`
	VisibleAgent bool   `json:"visibleAgent"`
	AgentID      string `json:"agentId" validate:"required"`
	AuteurID     string `json:"auteurId,omitempty"`
}

// UpdateObservationRequest represents request to update an observation
type UpdateObservationRequest struct {
	Contenu      *string `json:"contenu,omitempty"`
	Type         *string `json:"type,omitempty"`
	Categorie    *string `json:"categorie,omitempty"`
	VisibleAgent *bool   `json:"visibleAgent,omitempty"`
}

// ListObservationsFilters represents query filters for listing observations
type ListObservationsFilters struct {
	AgentID      string `query:"agentId"`
	AuteurID     string `query:"auteurId"`
	Type         string `query:"type"`
	Categorie    string `query:"categorie"`
	VisibleAgent string `query:"visibleAgent"`
}
