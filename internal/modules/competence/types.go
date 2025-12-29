package competence

import "time"

// CompetenceResponse represents a competence in API responses
type CompetenceResponse struct {
	ID             string           `json:"id"`
	Nom            string           `json:"nom"`
	Type           string           `json:"type"`
	Description    string           `json:"description,omitempty"`
	Organisme      string           `json:"organisme,omitempty"`
	DateObtention  *time.Time       `json:"dateObtention,omitempty"`
	DateExpiration *time.Time       `json:"dateExpiration,omitempty"`
	Active         bool             `json:"active"`
	Agents         []AgentResponse  `json:"agents,omitempty"`
	NombreAgents   int              `json:"nombreAgents"`
	JoursRestants  *int             `json:"joursRestants,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// AgentResponse represents an agent in competence responses
type AgentResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
	Grade     string `json:"grade,omitempty"`
}

// CreateCompetenceRequest represents request to create a competence
type CreateCompetenceRequest struct {
	Nom            string     `json:"nom" validate:"required"`
	Type           string     `json:"type" validate:"required"`
	Description    string     `json:"description,omitempty"`
	Organisme      string     `json:"organisme,omitempty"`
	DateObtention  *time.Time `json:"dateObtention,omitempty"`
	DateExpiration *time.Time `json:"dateExpiration,omitempty"`
}

// UpdateCompetenceRequest represents request to update a competence
type UpdateCompetenceRequest struct {
	Nom            *string    `json:"nom,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Organisme      *string    `json:"organisme,omitempty"`
	DateExpiration *time.Time `json:"dateExpiration,omitempty"`
	Active         *bool      `json:"active,omitempty"`
}

// AssignCompetenceRequest represents request to assign a competence to an agent
type AssignCompetenceRequest struct {
	AgentID string `json:"agentId" validate:"required"`
}

// ListCompetencesFilters represents query filters for listing competences
type ListCompetencesFilters struct {
	Type      string `query:"type"`
	Active    string `query:"active"`
	Search    string `query:"search"`
	Organisme string `query:"organisme"`
}
