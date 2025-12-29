package pv

import (
	"time"
)

// CreatePVRequest represents request to create a PV
type CreatePVRequest struct {
	InfractionIDs      []string   `json:"infraction_ids" validate:"required,min=1"`
	MontantTotal       float64    `json:"montant_total" validate:"required,gt=0"`
	DateLimitePaiement *time.Time `json:"date_limite_paiement,omitempty"`
	Observations       *string    `json:"observations,omitempty"`
	ControleID         *string    `json:"controle_id,omitempty"`
	InspectionID       *string    `json:"inspection_id,omitempty"`
}

// UpdatePVRequest represents request to update a PV
type UpdatePVRequest struct {
	MontantTotal       *float64   `json:"montant_total,omitempty" validate:"omitempty,gt=0"`
	DateLimitePaiement *time.Time `json:"date_limite_paiement,omitempty"`
	Observations       *string    `json:"observations,omitempty"`
}

// PayerPVRequest represents request to pay a PV
type PayerPVRequest struct {
	MontantPaye       float64 `json:"montant_paye" validate:"required,gt=0"`
	MoyenPaiement     string  `json:"moyen_paiement" validate:"required,oneof=CB CHEQUE ESPECES VIREMENT MOBILE_MONEY"`
	ReferencePaiement *string `json:"reference_paiement,omitempty"`
}

// ContesterPVRequest represents request to contest a PV
type ContesterPVRequest struct {
	MotifContestation string  `json:"motif_contestation" validate:"required"`
	TribunalCompetent *string `json:"tribunal_competent,omitempty"`
}

// DecisionContestationRequest represents decision on a contestation
type DecisionContestationRequest struct {
	Decision      string   `json:"decision" validate:"required,oneof=ACCEPTE REFUSE_PARTIEL REFUSE_TOTAL"`
	Motif         string   `json:"motif" validate:"required"`
	NouveauMontant *float64 `json:"nouveau_montant,omitempty"`
}

// MajorerPVRequest represents request to add penalty to a PV
type MajorerPVRequest struct {
	MontantMajore  float64   `json:"montant_majore" validate:"required,gt=0"`
	DateMajoration time.Time `json:"date_majoration" validate:"required"`
}

// AnnulerPVRequest represents request to cancel a PV
type AnnulerPVRequest struct {
	Motif string `json:"motif" validate:"required"`
}

// ListPVRequest represents filter for listing PVs
type ListPVRequest struct {
	InfractionID *string    `json:"infraction_id,omitempty"`
	Statut       *string    `json:"statut,omitempty"`
	DateDebut    *time.Time `json:"date_debut,omitempty"`
	DateFin      *time.Time `json:"date_fin,omitempty"`
	MontantMin   *float64   `json:"montant_min,omitempty"`
	MontantMax   *float64   `json:"montant_max,omitempty"`
	Expired      *bool      `json:"expired,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// PVResponse represents a PV in responses
type PVResponse struct {
	ID                   string              `json:"id"`
	NumeroPV             string              `json:"numero_pv"`
	DateEmission         time.Time           `json:"date_emission"`
	MontantTotal         float64             `json:"montant_total"`
	MontantMajore        float64             `json:"montant_majore,omitempty"`
	DateLimitePaiement   *time.Time          `json:"date_limite_paiement,omitempty"`
	DateMajoration       *time.Time          `json:"date_majoration,omitempty"`
	Statut               string              `json:"statut"`
	DatePaiement         *time.Time          `json:"date_paiement,omitempty"`
	MontantPaye          float64             `json:"montant_paye,omitempty"`
	MoyenPaiement        string              `json:"moyen_paiement,omitempty"`
	ReferencePaiement    string              `json:"reference_paiement,omitempty"`
	DateContestation     *time.Time          `json:"date_contestation,omitempty"`
	MotifContestation    string              `json:"motif_contestation,omitempty"`
	DecisionContestation string              `json:"decision_contestation,omitempty"`
	TribunalCompetent    string              `json:"tribunal_competent,omitempty"`
	Observations         string              `json:"observations,omitempty"`
	Infractions          []*InfractionSummary `json:"infractions,omitempty"`
	NombrePaiements      int                 `json:"nombre_paiements"`
	NombreRecours        int                 `json:"nombre_recours"`
	EstExpire            bool                `json:"est_expire"`
	MontantRestant       float64             `json:"montant_restant"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// InfractionSummary represents infraction summary for PV response
type InfractionSummary struct {
	ID             string    `json:"id"`
	NumeroPV       string    `json:"numero_pv"`
	DateInfraction time.Time `json:"date_infraction"`
	TypeInfraction string    `json:"type_infraction"`
	LieuInfraction string    `json:"lieu_infraction"`
	MontantAmende  float64   `json:"montant_amende"`
}

// ListPVResponse represents response for listing PVs
type ListPVResponse struct {
	PVs   []*PVResponse `json:"pvs"`
	Total int           `json:"total"`
}

// PVStatisticsResponse represents PV statistics
type PVStatisticsResponse struct {
	TotalPV          int                `json:"total_pv"`
	MontantTotal     float64            `json:"montant_total"`
	MontantPaye      float64            `json:"montant_paye"`
	MontantImpaye    float64            `json:"montant_impaye"`
	TauxRecouvrement float64            `json:"taux_recouvrement"`
	PVExpires        int                `json:"pv_expires"`
	ParStatut        map[string]int     `json:"par_statut"`
	ParMois          map[string]float64 `json:"par_mois"`
}

// RappelResponse represents response for sending a reminder
type RappelResponse struct {
	PVID          string    `json:"pv_id"`
	NumeroPV      string    `json:"numero_pv"`
	DateRappel    time.Time `json:"date_rappel"`
	NumeroRappel  int       `json:"numero_rappel"`
	MontantDu     float64   `json:"montant_du"`
	DateLimite    time.Time `json:"date_limite"`
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
}
