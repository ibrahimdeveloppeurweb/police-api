package commissariat

import "time"

// DashboardResponse represents commissariat dashboard
type DashboardResponse struct {
	Statistiques    *DashboardStats      `json:"statistiques"`
	ControleRecents []*ControleRecent    `json:"controleRecents"`
	AlertesRecentes []*AlerteRecente     `json:"alertesRecentes"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	ControlesTotal   int `json:"controlesTotal"`
	PvTotal          int `json:"pvTotal"`
	AgentsActifs     int `json:"agentsActifs"`
	AlertesEnCours   int `json:"alertesEnCours"`
}

// ControleRecent represents recent control
type ControleRecent struct {
	ID           string    `json:"id"`
	DateControle time.Time `json:"dateControle"`
	LieuControle string    `json:"lieuControle"`
	TypeControle string    `json:"typeControle"`
	Statut       string    `json:"statut"`
	AgentNom     string    `json:"agentNom"`
}

// AlerteRecente represents recent alert
type AlerteRecente struct {
	ID         string    `json:"id"`
	Titre      string    `json:"titre"`
	Niveau     string    `json:"niveau"`
	Statut     string    `json:"statut"`
	DateAlerte time.Time `json:"dateAlerte"`
}

// AgentResponse represents agent in commissariat context
type AgentResponse struct {
	ID        string    `json:"id"`
	Matricule string    `json:"matricule"`
	Nom       string    `json:"nom"`
	Prenom    string    `json:"prenom"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Actif     bool      `json:"actif"`
	CreatedAt time.Time `json:"createdAt"`
}

// ControleResponse represents control in commissariat context
type ControleResponse struct {
	ID            string             `json:"id"`
	DateControle  time.Time          `json:"dateControle"`
	LieuControle  string             `json:"lieuControle"`
	TypeControle  string             `json:"typeControle"`
	Statut        string             `json:"statut"`
	Observations  string             `json:"observations,omitempty"`
	Agent         *AgentSummary      `json:"agent,omitempty"`
	Vehicule      *VehiculeSummary   `json:"vehicule,omitempty"`
	Conducteur    *ConducteurSummary `json:"conducteur,omitempty"`
	NbInfractions int                `json:"nbInfractions"`
	CreatedAt     time.Time          `json:"createdAt"`
}

// AgentSummary represents agent summary
type AgentSummary struct {
	ID        string `json:"id"`
	Matricule string `json:"matricule"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
}

// VehiculeSummary represents vehicle summary
type VehiculeSummary struct {
	ID             string `json:"id"`
	Immatriculation string `json:"immatriculation"`
	Marque         string `json:"marque,omitempty"`
	Modele         string `json:"modele,omitempty"`
}

// ConducteurSummary represents conductor summary
type ConducteurSummary struct {
	ID     string `json:"id"`
	Nom    string `json:"nom"`
	Prenom string `json:"prenom"`
}

// StatistiquesResponse represents commissariat statistics
type StatistiquesResponse struct {
	Controles   *ControleStats   `json:"controles"`
	PV          *PVStats         `json:"pv"`
	Infractions *InfractionStats `json:"infractions"`
}

// ControleStats represents control statistics
type ControleStats struct {
	Total     int               `json:"total"`
	ParType   map[string]int    `json:"parType"`
	Evolution []EvolutionEntry  `json:"evolution"`
}

// EvolutionEntry represents evolution entry
type EvolutionEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// PVStats represents PV statistics
type PVStats struct {
	Total        int     `json:"total"`
	MontantTotal float64 `json:"montantTotal"`
	TauxPaiement float64 `json:"tauxPaiement"`
}

// InfractionStats represents infraction statistics
type InfractionStats struct {
	Total   int            `json:"total"`
	ParType map[string]int `json:"parType"`
}

// ListControlesResponse represents paginated controls
type ListControlesResponse struct {
	Data       []*ControleResponse `json:"data"`
	Pagination *Pagination         `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// CommissariatResponse represents commissariat in list
type CommissariatResponse struct {
	ID           string    `json:"id"`
	Nom          string    `json:"nom"`
	Code         string    `json:"code"`
	Ville        string    `json:"ville"`
	Region       string    `json:"region,omitempty"`
	Adresse      string    `json:"adresse,omitempty"`
	Telephone    string    `json:"telephone,omitempty"`
	Email        string    `json:"email,omitempty"`
	Actif        bool      `json:"actif"`
	NbAgents     int       `json:"nbAgents,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ListCommissariatsResponse represents paginated commissariats
type ListCommissariatsResponse struct {
	Data       []*CommissariatResponse `json:"data"`
	Pagination *Pagination             `json:"pagination"`
}
