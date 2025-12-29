package admin

import "time"

// StatistiquesNationales represents national statistics
type StatistiquesNationales struct {
	ControlesTotal        int                        `json:"controlesTotal"`
	PvTotal               int                        `json:"pvTotal"`
	MontantPVTotal        float64                    `json:"montantPVTotal"`
	AlertesActives        int                        `json:"alertesActives"`
	CommissariatsActifs   int                        `json:"commissariatsActifs"`
	AgentsActifs          int                        `json:"agentsActifs"`
	TauxPaiementPV        float64                    `json:"tauxPaiementPV"`
	EvolutionControles    []EvolutionEntry           `json:"evolutionControles"`
	TopInfractions        []InfractionCount          `json:"topInfractions"`
	StatistiquesParRegion []StatistiquesRegion       `json:"statistiquesParRegion"`
}

// EvolutionEntry represents a date/count entry
type EvolutionEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// InfractionCount represents infraction statistics
type InfractionCount struct {
	TypeCode    string  `json:"typeCode"`
	TypeLibelle string  `json:"typeLibelle"`
	Count       int     `json:"count"`
	Montant     float64 `json:"montant"`
}

// StatistiquesRegion represents statistics by region
type StatistiquesRegion struct {
	Region        string `json:"region"`
	Commissariats int    `json:"commissariats"`
	Agents        int    `json:"agents"`
	Controles     int    `json:"controles"`
}

// CommissariatResponse represents a commissariat in API responses
type CommissariatResponse struct {
	ID        string    `json:"id"`
	Nom       string    `json:"nom"`
	Code      string    `json:"code"`
	Adresse   string    `json:"adresse"`
	Ville     string    `json:"ville"`
	Region    string    `json:"region"`
	Telephone string    `json:"telephone"`
	Email     string    `json:"email,omitempty"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Actif     bool      `json:"actif"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AgentResponse represents an agent in API responses
type AgentResponse struct {
	ID               string                `json:"id"`
	Matricule        string                `json:"matricule"`
	Nom              string                `json:"nom"`
	Prenom           string                `json:"prenom"`
	Email            string                `json:"email"`
	Role             string                `json:"role"`
	Grade            string                `json:"grade,omitempty"`
	Telephone        string                `json:"telephone,omitempty"`
	StatutService    string                `json:"statutService"`
	Localisation     string                `json:"localisation,omitempty"`
	Activite         string                `json:"activite,omitempty"`
	DerniereActivite *time.Time            `json:"derniereActivite,omitempty"`
	Actif            bool                  `json:"actif"`
	CommissariatID   string                `json:"commissariatId,omitempty"`
	Commissariat     *CommissariatResponse `json:"commissariat,omitempty"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
	// Nouveaux champs pour informations personnelles
	DateNaissance *time.Time `json:"dateNaissance,omitempty"`
	CNI           string     `json:"cni,omitempty"`
	Adresse       string     `json:"adresse,omitempty"`
	DateEntree    *time.Time `json:"dateEntree,omitempty"`
	GpsPrecision  float64    `json:"gpsPrecision"`
	TempsService  string     `json:"tempsService,omitempty"`
	// Nouvelles relations
	Equipe       *EquipeResponse       `json:"equipe,omitempty"`
	Superieur    *SuperieurResponse    `json:"superieur,omitempty"`
	Missions     []MissionResponse     `json:"missions,omitempty"`
	Objectifs    []ObjectifResponse    `json:"objectifs,omitempty"`
	Observations []ObservationResponse `json:"observations,omitempty"`
	Competences  []CompetenceResponse  `json:"competences,omitempty"`
}

// EquipeResponse represents an equipe in API responses
type EquipeResponse struct {
	ID          string `json:"id"`
	Nom         string `json:"nom"`
	Code        string `json:"code"`
	Zone        string `json:"zone,omitempty"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
}

// SuperieurResponse represents the hierarchical superior
type SuperieurResponse struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Grade     string `json:"grade,omitempty"`
	Matricule string `json:"matricule"`
}

// MissionResponse represents a mission in API responses
type MissionResponse struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Titre     string     `json:"titre,omitempty"`
	DateDebut time.Time  `json:"dateDebut"`
	DateFin   *time.Time `json:"dateFin,omitempty"`
	Duree     string     `json:"duree,omitempty"`
	Zone      string     `json:"zone,omitempty"`
	Statut    string     `json:"statut"`
	Rapport   string     `json:"rapport,omitempty"`
}

// ObjectifResponse represents an objectif in API responses
type ObjectifResponse struct {
	ID             string    `json:"id"`
	Titre          string    `json:"titre"`
	Description    string    `json:"description,omitempty"`
	Periode        string    `json:"periode"`
	DateDebut      time.Time `json:"dateDebut"`
	DateFin        time.Time `json:"dateFin"`
	Statut         string    `json:"statut"`
	ValeurCible    int       `json:"valeurCible,omitempty"`
	ValeurActuelle int       `json:"valeurActuelle"`
	Progression    float64   `json:"progression"`
}

// ObservationResponse represents an observation in API responses
type ObservationResponse struct {
	ID           string    `json:"id"`
	Contenu      string    `json:"contenu"`
	Type         string    `json:"type"`
	Categorie    string    `json:"categorie,omitempty"`
	VisibleAgent bool      `json:"visibleAgent"`
	CreatedAt    time.Time `json:"createdAt"`
	AuteurNom    string    `json:"auteurNom,omitempty"`
}

// CompetenceResponse represents a competence in API responses
type CompetenceResponse struct {
	ID             string     `json:"id"`
	Nom            string     `json:"nom"`
	Type           string     `json:"type"`
	Description    string     `json:"description,omitempty"`
	Organisme      string     `json:"organisme,omitempty"`
	DateObtention  *time.Time `json:"dateObtention,omitempty"`
	DateExpiration *time.Time `json:"dateExpiration,omitempty"`
	Active         bool       `json:"active"`
}

// UpdateAgentRequest represents request to update an agent
type UpdateAgentRequest struct {
	Nom            *string `json:"nom,omitempty"`
	Prenom         *string `json:"prenom,omitempty"`
	Email          *string `json:"email,omitempty"`
	Role           *string `json:"role,omitempty"`
	Grade          *string `json:"grade,omitempty"`
	Telephone      *string `json:"telephone,omitempty"`
	StatutService  *string `json:"statutService,omitempty"`
	Localisation   *string `json:"localisation,omitempty"`
	Activite       *string `json:"activite,omitempty"`
	Actif          *bool   `json:"actif,omitempty"`
	CommissariatID *string `json:"commissariatId,omitempty"`
}

// CreateCommissariatRequest represents request to create a commissariat
type CreateCommissariatRequest struct {
	Nom       string   `json:"nom" validate:"required"`
	Code      string   `json:"code" validate:"required"`
	Adresse   string   `json:"adresse" validate:"required"`
	Ville     string   `json:"ville" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	Telephone string   `json:"telephone" validate:"required"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// UpdateCommissariatRequest represents request to update a commissariat
type UpdateCommissariatRequest struct {
	Nom       *string  `json:"nom,omitempty"`
	Adresse   *string  `json:"adresse,omitempty"`
	Ville     *string  `json:"ville,omitempty"`
	Region    *string  `json:"region,omitempty"`
	Telephone *string  `json:"telephone,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Actif     *bool    `json:"actif,omitempty"`
}

// CreateAgentRequest represents request to create an agent
type CreateAgentRequest struct {
	Matricule      string  `json:"matricule" validate:"required"`
	Nom            string  `json:"nom" validate:"required"`
	Prenom         string  `json:"prenom" validate:"required"`
	Email          string  `json:"email" validate:"required,email"`
	Password       string  `json:"password" validate:"required,min=6"`
	Role           string  `json:"role" validate:"required"`
	Grade          *string `json:"grade,omitempty"`
	Telephone      *string `json:"telephone,omitempty"`
	CommissariatID *string `json:"commissariatId,omitempty"`
}

// AgentStatistiquesResponse represents agent statistics
type AgentStatistiquesResponse struct {
	AgentID          string         `json:"agentId"`
	TotalControles   int            `json:"totalControles"`
	TotalInfractions int            `json:"totalInfractions"`
	TotalPV          int            `json:"totalPV"`
	MontantTotalPV   float64        `json:"montantTotalPV"`
	TauxInfraction   float64        `json:"tauxInfraction"`
	ControlesParJour float64        `json:"controlesParJour"`
	ControlesParMois map[string]int `json:"controlesParMois"`
}

// ==================== DASHBOARD AGENTS TYPES ====================

// AgentDashboardRequest represents the request for agent dashboard
type AgentDashboardRequest struct {
	Periode   string `query:"periode"`   // jour, semaine, mois, annee, tout, personnalise
	DateDebut string `query:"dateDebut"` // For personnalise period
	DateFin   string `query:"dateFin"`   // For personnalise period
}

// AgentDashboardResponse represents the full dashboard response for agents
type AgentDashboardResponse struct {
	Stats           AgentDashboardStats           `json:"stats"`
	ActivityData    []ActivityDataEntry           `json:"activityData"`
	PerformanceData []PerformanceDataEntry        `json:"performanceData"`
	PieData         []PieDataEntry                `json:"pieData"`
	Agents          []AgentDetailedResponse       `json:"agents"`
	Commissariats   []CommissariatStatsEntry      `json:"commissariats"`
}

// AgentDashboardStats represents global statistics for agents dashboard
type AgentDashboardStats struct {
	TotalAgents        int     `json:"totalAgents"`
	EnService          int     `json:"enService"`
	EnPause            int     `json:"enPause"`
	HorsService        int     `json:"horsService"`
	ControlesTotal     int     `json:"controlesTotal"`
	InfractionsTotales int     `json:"infractionsTotales"`
	RevenusTotal       float64 `json:"revenusTotal"`
	PerformanceMoyenne float64 `json:"performanceMoyenne"`
	TempsServiceMoyen  string  `json:"tempsServiceMoyen"`
	TauxReussite       float64 `json:"tauxReussite"`
}

// ActivityDataEntry represents activity data for charts
type ActivityDataEntry struct {
	Period      string `json:"period"`
	Controles   int    `json:"controles"`
	Agents      int    `json:"agents"`
	Infractions int    `json:"infractions"`
}

// PerformanceDataEntry represents performance data by commissariat
type PerformanceDataEntry struct {
	Commissariat string  `json:"commissariat"`
	TauxActivite float64 `json:"tauxActivite"`
	Agents       int     `json:"agents"`
}

// PieDataEntry represents data for pie charts
type PieDataEntry struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

// AgentDetailedResponse represents detailed agent information for dashboard
type AgentDetailedResponse struct {
	ID               string  `json:"id"`
	Nom              string  `json:"nom"`
	Grade            string  `json:"grade"`
	Commissariat     string  `json:"commissariat"`
	Status           string  `json:"status"`
	Localisation     string  `json:"localisation"`
	Activite         string  `json:"activite"`
	Controles        int     `json:"controles"`
	Infractions      int     `json:"infractions"`
	Revenus          float64 `json:"revenus"`
	TauxInfractions  float64 `json:"tauxInfractions"`
	TempsService     string  `json:"tempsService"`
	Gps              int     `json:"gps"`
	DerniereActivite string  `json:"derniereActivite"`
	Performance      string  `json:"performance"`
}

// CommissariatStatsEntry represents commissariat statistics
type CommissariatStatsEntry struct {
	Name         string  `json:"name"`
	Agents       int     `json:"agents"`
	EnService    int     `json:"enService"`
	Controles    int     `json:"controles"`
	TauxActivite float64 `json:"tauxActivite"`
}

// ==================== SESSION MANAGEMENT TYPES ====================

// AgentSessionResponse represents a user session for admin viewing
type AgentSessionResponse struct {
	ID             string    `json:"id"`
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name,omitempty"`
	DeviceType     string    `json:"device_type,omitempty"`
	DeviceOS       string    `json:"device_os,omitempty"`
	AppVersion     string    `json:"app_version,omitempty"`
	LastActivityAt time.Time `json:"last_activity_at"`
	LastIPAddress  string    `json:"last_ip_address,omitempty"`
	SessionStarted time.Time `json:"session_started"`
	IsActive       bool      `json:"is_active"`
}

// RevokeSessionRequest represents a request to revoke a session
type RevokeSessionRequest struct {
	Reason string `json:"reason,omitempty"`
}
