package infraction

import (
	"time"
)

// Request types

// CreateInfractionRequest represents the request to create an infraction
type CreateInfractionRequest struct {
	DateInfraction      time.Time `json:"date_infraction" validate:"required"`
	LieuInfraction      string    `json:"lieu_infraction" validate:"required"`
	Circonstances       *string   `json:"circonstances,omitempty"`
	VitesseRetenue      *float64  `json:"vitesse_retenue,omitempty"`
	VitesseLimitee      *float64  `json:"vitesse_limitee,omitempty"`
	AppareilMesure      *string   `json:"appareil_mesure,omitempty"`
	Statut              string    `json:"statut,omitempty"`
	Observations        *string   `json:"observations,omitempty"`
	FlagrantDelit       bool      `json:"flagrant_delit,omitempty"`
	Accident            bool      `json:"accident,omitempty"`
	ControleID          string    `json:"controle_id" validate:"required"`
	TypeInfractionID    string    `json:"type_infraction_id" validate:"required"`
	VehiculeID          string    `json:"vehicule_id" validate:"required"`
	ConducteurID        string    `json:"conducteur_id" validate:"required"`
}

// UpdateInfractionRequest represents the request to update an infraction
type UpdateInfractionRequest struct {
	NumeroPV            *string   `json:"numero_pv,omitempty"`
	DateInfraction      *time.Time `json:"date_infraction,omitempty"`
	LieuInfraction      *string   `json:"lieu_infraction,omitempty"`
	Circonstances       *string   `json:"circonstances,omitempty"`
	VitesseRetenue      *float64  `json:"vitesse_retenue,omitempty"`
	VitesseLimitee      *float64  `json:"vitesse_limitee,omitempty"`
	AppareilMesure      *string   `json:"appareil_mesure,omitempty"`
	Statut              *string   `json:"statut,omitempty"`
	Observations        *string   `json:"observations,omitempty"`
	FlagrantDelit       *bool     `json:"flagrant_delit,omitempty"`
	Accident            *bool     `json:"accident,omitempty"`
	TypeInfractionID    *string   `json:"type_infraction_id,omitempty"`
}

// ListInfractionsRequest represents the request to list infractions
type ListInfractionsRequest struct {
	ControleID       *string    `json:"controle_id,omitempty"`
	VehiculeID       *string    `json:"vehicule_id,omitempty"`
	ConducteurID     *string    `json:"conducteur_id,omitempty"`
	TypeInfractionID *string    `json:"type_infraction_id,omitempty"`
	Statut           *string    `json:"statut,omitempty"`
	LieuInfraction   *string    `json:"lieu_infraction,omitempty"`
	DateDebut        *time.Time `json:"date_debut,omitempty"`
	DateFin          *time.Time `json:"date_fin,omitempty"`
	FlagrantDelit    *bool      `json:"flagrant_delit,omitempty"`
	Accident         *bool      `json:"accident,omitempty"`
	Limit            int        `json:"limit,omitempty"`
	Offset           int        `json:"offset,omitempty"`
}

// GeneratePVRequest represents the request to generate a PV
type GeneratePVRequest struct {
	InfractionID string `json:"infraction_id" validate:"required"`
}

// Response types

// InfractionResponse represents an infraction in API responses
type InfractionResponse struct {
	ID                  string                    `json:"id"`
	NumeroPV            string                    `json:"numero_pv,omitempty"`
	DateInfraction      time.Time                 `json:"date_infraction"`
	LieuInfraction      string                    `json:"lieu_infraction"`
	Circonstances       string                    `json:"circonstances,omitempty"`
	VitesseRetenue      *float64                  `json:"vitesse_retenue,omitempty"`
	VitesseLimitee      *float64                  `json:"vitesse_limitee,omitempty"`
	AppareilMesure      string                    `json:"appareil_mesure,omitempty"`
	MontantAmende       float64                   `json:"montant_amende"`
	PointsRetires       int                       `json:"points_retires"`
	Statut              string                    `json:"statut"`
	Observations        string                    `json:"observations,omitempty"`
	FlagrantDelit       bool                      `json:"flagrant_delit"`
	Accident            bool                      `json:"accident"`
	Controle            *ControleSummary          `json:"controle,omitempty"`
	TypeInfraction      *TypeInfractionSummary    `json:"type_infraction,omitempty"`
	Vehicule            *VehiculeSummary          `json:"vehicule,omitempty"`
	Conducteur          *ConducteurSummary        `json:"conducteur,omitempty"`
	ProcesVerbal        *ProcesVerbalSummary      `json:"proces_verbal,omitempty"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

// ControleSummary represents controle information in infraction responses
type ControleSummary struct {
	ID           string    `json:"id"`
	DateControle time.Time `json:"date_controle"`
	LieuControle string    `json:"lieu_controle"`
	TypeControle string    `json:"type_controle"`
	AgentNom     string    `json:"agent_nom,omitempty"`
}

// TypeInfractionSummary represents type infraction information
type TypeInfractionSummary struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Libelle     string  `json:"libelle"`
	Description string  `json:"description,omitempty"`
	Amende      float64 `json:"amende"`
	Points      int     `json:"points"`
	Categorie   string  `json:"categorie"`
}

// VehiculeSummary represents vehicule information in infraction responses
type VehiculeSummary struct {
	ID              string `json:"id"`
	Immatriculation string `json:"immatriculation"`
	Marque          string `json:"marque"`
	Modele          string `json:"modele"`
	TypeVehicule    string `json:"type_vehicule"`
}

// ConducteurSummary represents conducteur information in infraction responses
type ConducteurSummary struct {
	ID            string `json:"id"`
	Nom           string `json:"nom"`
	Prenom        string `json:"prenom"`
	NumeroPermis  string `json:"numero_permis,omitempty"`
	PointsPermis  int    `json:"points_permis"`
}

// ProcesVerbalSummary represents proces verbal information
type ProcesVerbalSummary struct {
	ID            string    `json:"id"`
	NumeroPV      string    `json:"numero_pv"`
	DateEmission  time.Time `json:"date_emission"`
	MontantTotal  float64   `json:"montant_total"`
	Statut        string    `json:"statut"`
}

// ListInfractionsResponse represents the response for listing infractions
type ListInfractionsResponse struct {
	Infractions []*InfractionResponse `json:"infractions"`
	Total       int                   `json:"total"`
}

// InfractionStatisticsResponse represents statistics for infractions
type InfractionStatisticsResponse struct {
	Total              int                     `json:"total"`
	ParStatut          map[string]int          `json:"par_statut"`
	ParType            map[string]int          `json:"par_type"`
	ParMois            map[string]int          `json:"par_mois"`
	MontantTotal       float64                 `json:"montant_total"`
	PointsTotal        int                     `json:"points_total"`
	FlagrantDelitTotal int                     `json:"flagrant_delit_total"`
	AccidentTotal      int                     `json:"accident_total"`
	TopInfractions     []TypeInfractionStats   `json:"top_infractions"`
	PeriodeDebut       *time.Time              `json:"periode_debut,omitempty"`
	PeriodeFin         *time.Time              `json:"periode_fin,omitempty"`
}

// TypeInfractionStats represents statistics by infraction type
type TypeInfractionStats struct {
	TypeCode     string  `json:"type_code"`
	TypeLibelle  string  `json:"type_libelle"`
	Count        int     `json:"count"`
	MontantTotal float64 `json:"montant_total"`
}

// PVGenerationResponse represents the response for PV generation
type PVGenerationResponse struct {
	ProcesVerbalID string    `json:"proces_verbal_id"`
	NumeroPV       string    `json:"numero_pv"`
	DateEmission   time.Time `json:"date_emission"`
	MontantTotal   float64   `json:"montant_total"`
	Success        bool      `json:"success"`
	Message        string    `json:"message"`
}

// InfractionValidationResponse represents the response for infraction validation
type InfractionValidationResponse struct {
	InfractionID    string    `json:"infraction_id"`
	NumeroPV        string    `json:"numero_pv"`
	StatutPrecedent string    `json:"statut_precedent"`
	NouveauStatut   string    `json:"nouveau_statut"`
	DateValidation  time.Time `json:"date_validation"`
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
}

// InfractionsByTypeResponse represents infractions grouped by type
type InfractionsByTypeResponse struct {
	TypeInfraction *TypeInfractionSummary `json:"type_infraction"`
	Infractions    []*InfractionResponse  `json:"infractions"`
	Count          int                    `json:"count"`
	MontantTotal   float64                `json:"montant_total"`
	PointsTotal    int                    `json:"points_total"`
}

// CategorieResponse represents an infraction category
type CategorieResponse struct {
	Code        string `json:"code"`
	Libelle     string `json:"libelle"`
	Description string `json:"description,omitempty"`
	NbTypes     int    `json:"nbTypes"`
}

// InfractionArchiveResponse represents the response for archiving an infraction
type InfractionArchiveResponse struct {
	InfractionID    string    `json:"infraction_id"`
	StatutPrecedent string    `json:"statut_precedent"`
	NouveauStatut   string    `json:"nouveau_statut"`
	DateArchivage   time.Time `json:"date_archivage"`
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
}

// PaymentRequest represents the request to record a payment
type PaymentRequest struct {
	ModePaiement string  `json:"mode_paiement" validate:"required"`
	Montant      float64 `json:"montant" validate:"required"`
	Reference    string  `json:"reference,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// PaymentResponse represents the response for payment registration
type PaymentResponse struct {
	InfractionID    string    `json:"infraction_id"`
	NumeroPV        string    `json:"numero_pv"`
	MontantPaye     float64   `json:"montant_paye"`
	ModePaiement    string    `json:"mode_paiement"`
	Reference       string    `json:"reference,omitempty"`
	StatutPrecedent string    `json:"statut_precedent"`
	NouveauStatut   string    `json:"nouveau_statut"`
	DatePaiement    time.Time `json:"date_paiement"`
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
}

// ===================== DASHBOARD TYPES =====================

// DashboardRequest represents the request for dashboard data
type DashboardRequest struct {
	Periode   string     `json:"periode,omitempty"`    // jour, semaine, mois, annee, tout
	DateDebut *time.Time `json:"date_debut,omitempty"`
	DateFin   *time.Time `json:"date_fin,omitempty"`
}

// DashboardStats represents the main statistics for the dashboard
type DashboardStats struct {
	TotalInfractions int     `json:"totalInfractions"`
	Revenus          string  `json:"revenus"`
	TotalTypes       int     `json:"totalTypes"`
	MontantMoyen     string  `json:"montantMoyen"`
	TauxContestation string  `json:"tauxContestation"`
	TauxPaiement     string  `json:"tauxPaiement"`
	Infractions24h   int     `json:"infractions24h"`
	Evolution        float64 `json:"evolution"`
}

// ActivityDataEntry represents activity data for charts
type ActivityDataEntry struct {
	Period       string `json:"period"`
	Total        int    `json:"total"`
	Documents    int    `json:"documents"`
	Securite     int    `json:"securite"`
	Comportement int    `json:"comportement"`
	Technique    int    `json:"technique"`
}

// PieDataEntry represents data for pie charts
type PieDataEntry struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

// EvolutionDataEntry represents evolution data by category
type EvolutionDataEntry struct {
	Category  string  `json:"category"`
	Evolution float64 `json:"evolution"`
}

// CategoryDataEntry represents category data for display
type CategoryDataEntry struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Count       int      `json:"count"`
	BgColor     string   `json:"bgColor"`
	IconColor   string   `json:"iconColor"`
	Infractions []string `json:"infractions"`
	Evolution   float64  `json:"evolution"`
}

// TopInfractionEntry represents a top infraction entry
type TopInfractionEntry struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Category   string  `json:"category"`
}

// DashboardResponse represents the complete dashboard response
type DashboardResponse struct {
	Stats         DashboardStats       `json:"stats"`
	ActivityData  []ActivityDataEntry  `json:"activityData"`
	PieData       []PieDataEntry       `json:"pieData"`
	EvolutionData []EvolutionDataEntry `json:"evolutionData"`
	Categories    []CategoryDataEntry  `json:"categories"`
	TopInfractions []TopInfractionEntry `json:"topInfractions"`
}