package objectif

import "time"

// ObjectifResponse represents an objectif in API responses
type ObjectifResponse struct {
	ID             string         `json:"id"`
	Titre          string         `json:"titre"`
	Description    string         `json:"description,omitempty"`
	Periode        string         `json:"periode"`
	DateDebut      time.Time      `json:"dateDebut"`
	DateFin        time.Time      `json:"dateFin"`
	Statut         string         `json:"statut"`
	ValeurCible    int            `json:"valeurCible,omitempty"`
	ValeurActuelle int            `json:"valeurActuelle"`
	Progression    float64        `json:"progression"`
	Agent          *AgentResponse `json:"agent,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// AgentResponse represents an agent in objectif responses
type AgentResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
}

// CreateObjectifRequest represents request to create an objectif
type CreateObjectifRequest struct {
	Titre       string    `json:"titre" validate:"required"`
	Description string    `json:"description,omitempty"`
	Periode     string    `json:"periode" validate:"required"`
	DateDebut   time.Time `json:"dateDebut" validate:"required"`
	DateFin     time.Time `json:"dateFin" validate:"required"`
	ValeurCible int       `json:"valeurCible,omitempty"`
	AgentID     string    `json:"agentId,omitempty"`
}

// UpdateObjectifRequest represents request to update an objectif
type UpdateObjectifRequest struct {
	Titre          *string    `json:"titre,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Statut         *string    `json:"statut,omitempty"`
	ValeurCible    *int       `json:"valeurCible,omitempty"`
	ValeurActuelle *int       `json:"valeurActuelle,omitempty"`
	DateFin        *time.Time `json:"dateFin,omitempty"`
}

// UpdateProgressionRequest represents request to update progression
type UpdateProgressionRequest struct {
	ValeurActuelle int `json:"valeurActuelle" validate:"required,min=0"`
}

// ListObjectifsFilters represents query filters for listing objectifs
type ListObjectifsFilters struct {
	AgentID   string `query:"agentId"`
	Periode   string `query:"periode"`
	Statut    string `query:"statut"`
	DateDebut string `query:"dateDebut"`
	DateFin   string `query:"dateFin"`
}
