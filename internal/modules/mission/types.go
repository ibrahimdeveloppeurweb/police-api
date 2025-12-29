package mission

import "time"

// MissionResponse represents a mission in API responses
type MissionResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Titre        string                `json:"titre,omitempty"`
	DateDebut    time.Time             `json:"dateDebut"`
	DateFin      *time.Time            `json:"dateFin,omitempty"`
	Duree        string                `json:"duree,omitempty"`
	Zone         string                `json:"zone,omitempty"`
	Statut       string                `json:"statut"`
	Rapport      string                `json:"rapport,omitempty"`
	Agents       []AgentResponse       `json:"agents,omitempty"`
	Equipe       *EquipeResponse       `json:"equipe,omitempty"`
	Commissariat *CommissariatResponse `json:"commissariat,omitempty"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

// AgentResponse represents an agent in mission responses
type AgentResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
	Grade     string `json:"grade,omitempty"`
}

// EquipeResponse represents an equipe in mission responses
type EquipeResponse struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// CommissariatResponse represents a commissariat in mission responses
type CommissariatResponse struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// CreateMissionRequest represents request to create a mission
type CreateMissionRequest struct {
	Type           string     `json:"type" validate:"required"`
	Titre          string     `json:"titre,omitempty"`
	DateDebut      time.Time  `json:"dateDebut" validate:"required"`
	DateFin        *time.Time `json:"dateFin,omitempty"`
	Duree          string     `json:"duree,omitempty"`
	Zone           string     `json:"zone,omitempty"`
	AgentIDs       []string   `json:"agentIds,omitempty"`
	EquipeID       string     `json:"equipeId,omitempty"`
	CommissariatID string     `json:"commissariatId,omitempty"`
}

// AddAgentsRequest represents request to add agents to a mission
type AddAgentsRequest struct {
	AgentIDs []string `json:"agentIds" validate:"required,min=1"`
}

// RemoveAgentRequest represents request to remove an agent from a mission
type RemoveAgentRequest struct {
	AgentID string `json:"agentId" validate:"required"`
}

// CancelMissionRequest represents request to cancel a mission
type CancelMissionRequest struct {
	Raison string `json:"raison,omitempty"`
}

// UpdateMissionRequest represents request to update a mission
type UpdateMissionRequest struct {
	Titre   *string    `json:"titre,omitempty"`
	Zone    *string    `json:"zone,omitempty"`
	Duree   *string    `json:"duree,omitempty"`
	Statut  *string    `json:"statut,omitempty"`
	Rapport *string    `json:"rapport,omitempty"`
	DateFin *time.Time `json:"dateFin,omitempty"`
}

// EndMissionRequest represents request to end a mission
type EndMissionRequest struct {
	Rapport string `json:"rapport" validate:"required"`
}

// ListMissionsFilters represents query filters for listing missions
type ListMissionsFilters struct {
	AgentID        string `query:"agentId"`
	EquipeID       string `query:"equipeId"`
	CommissariatID string `query:"commissariatId"`
	Statut         string `query:"statut"`
	Type           string `query:"type"`
	DateDebut      string `query:"dateDebut"`
	DateFin        string `query:"dateFin"`
}
