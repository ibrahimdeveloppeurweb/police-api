package recours

import (
	"time"
)

// CreateRecoursRequest represents request to create a recours
type CreateRecoursRequest struct {
	ProcesVerbalID     string     `json:"proces_verbal_id" validate:"required"`
	TypeRecours        string     `json:"type_recours" validate:"required,oneof=GRACIEUX CONTENTIEUX HIERARCHIQUE"`
	Motif              string     `json:"motif" validate:"required"`
	Argumentaire       string     `json:"argumentaire" validate:"required"`
	AutoriteCompetente *string    `json:"autorite_competente,omitempty"`
	DateLimiteRecours  *time.Time `json:"date_limite_recours,omitempty"`
	Observations       *string    `json:"observations,omitempty"`
}

// UpdateRecoursRequest represents request to update a recours
type UpdateRecoursRequest struct {
	TypeRecours        *string `json:"type_recours,omitempty" validate:"omitempty,oneof=GRACIEUX CONTENTIEUX HIERARCHIQUE"`
	Motif              *string `json:"motif,omitempty"`
	Argumentaire       *string `json:"argumentaire,omitempty"`
	AutoriteCompetente *string `json:"autorite_competente,omitempty"`
	Observations       *string `json:"observations,omitempty"`
}

// TraiterRecoursRequest represents request to process a recours
type TraiterRecoursRequest struct {
	Decision          string   `json:"decision" validate:"required,oneof=ACCEPTE REFUSE_PARTIEL REFUSE_TOTAL"`
	MotifDecision     string   `json:"motif_decision" validate:"required"`
	ReferenceDecision *string  `json:"reference_decision,omitempty"`
	NouveauMontant    *float64 `json:"nouveau_montant,omitempty"`
	RecoursPossible   *bool    `json:"recours_possible,omitempty"`
}

// AssignerRecoursRequest represents request to assign a recours
type AssignerRecoursRequest struct {
	TraiteParID string `json:"traite_par_id" validate:"required"`
}

// AbandonnerRecoursRequest represents request to abandon a recours
type AbandonnerRecoursRequest struct {
	Motif string `json:"motif" validate:"required"`
}

// ListRecoursRequest represents filter for listing recours
type ListRecoursRequest struct {
	ProcesVerbalID *string    `json:"proces_verbal_id,omitempty"`
	TypeRecours    *string    `json:"type_recours,omitempty"`
	Statut         *string    `json:"statut,omitempty"`
	TraiteParID    *string    `json:"traite_par_id,omitempty"`
	DateDebut      *time.Time `json:"date_debut,omitempty"`
	DateFin        *time.Time `json:"date_fin,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// RecoursResponse represents a recours in responses
type RecoursResponse struct {
	ID                 string              `json:"id"`
	NumeroRecours      string              `json:"numero_recours"`
	DateRecours        time.Time           `json:"date_recours"`
	TypeRecours        string              `json:"type_recours"`
	Motif              string              `json:"motif"`
	Argumentaire       string              `json:"argumentaire"`
	Statut             string              `json:"statut"`
	DateTraitement     *time.Time          `json:"date_traitement,omitempty"`
	Decision           string              `json:"decision,omitempty"`
	MotifDecision      string              `json:"motif_decision,omitempty"`
	AutoriteCompetente string              `json:"autorite_competente,omitempty"`
	ReferenceDecision  string              `json:"reference_decision,omitempty"`
	NouveauMontant     float64             `json:"nouveau_montant,omitempty"`
	DateLimiteRecours  *time.Time          `json:"date_limite_recours,omitempty"`
	RecoursPossible    bool                `json:"recours_possible"`
	Observations       string              `json:"observations,omitempty"`
	ProcesVerbal       *ProcesVerbalSummary `json:"proces_verbal,omitempty"`
	TraitePar          *UserSummary        `json:"traite_par,omitempty"`
	NombreDocuments    int                 `json:"nombre_documents"`
	DelaiTraitement    *int                `json:"delai_traitement,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// ProcesVerbalSummary represents PV summary for recours response
type ProcesVerbalSummary struct {
	ID           string    `json:"id"`
	NumeroPV     string    `json:"numero_pv"`
	DateEmission time.Time `json:"date_emission"`
	MontantTotal float64   `json:"montant_total"`
	Statut       string    `json:"statut"`
}

// UserSummary represents user summary for recours response
type UserSummary struct {
	ID        string `json:"id"`
	Matricule string `json:"matricule"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
}

// ListRecoursResponse represents response for listing recours
type ListRecoursResponse struct {
	Recours []*RecoursResponse `json:"recours"`
	Total   int                `json:"total"`
}

// RecoursStatisticsResponse represents recours statistics
type RecoursStatisticsResponse struct {
	TotalRecours    int            `json:"total_recours"`
	ParStatut       map[string]int `json:"par_statut"`
	ParType         map[string]int `json:"par_type"`
	ParDecision     map[string]int `json:"par_decision"`
	TauxAcceptation float64        `json:"taux_acceptation"`
	DelaiMoyenJours float64        `json:"delai_moyen_jours"`
}

// EtapeRecoursResponse represents a step in the recours workflow
type EtapeRecoursResponse struct {
	Code        string     `json:"code"`
	Libelle     string     `json:"libelle"`
	Description string     `json:"description"`
	Statut      string     `json:"statut"` // TERMINEE, EN_COURS, A_VENIR
	DateDebut   *time.Time `json:"date_debut,omitempty"`
	DateFin     *time.Time `json:"date_fin,omitempty"`
	Responsable *string    `json:"responsable,omitempty"`
	Ordre       int        `json:"ordre"`
}
