package paiement

import (
	"time"
)

// CreatePaiementRequest represents request to create a payment
type CreatePaiementRequest struct {
	ProcesVerbalID    string  `json:"proces_verbal_id" validate:"required"`
	Montant           float64 `json:"montant" validate:"required,gt=0"`
	MoyenPaiement     string  `json:"moyen_paiement" validate:"required,oneof=CB CHEQUE ESPECES VIREMENT MOBILE_MONEY TRESOR_PUBLIC"`
	ReferenceExterne  *string `json:"reference_externe,omitempty"`
	CodeAutorisation  *string `json:"code_autorisation,omitempty"`
	DetailsPaiement   *string `json:"details_paiement,omitempty"`
	// Champs spécifiques au paiement Trésor Public
	NumeroRecuTresor  *string `json:"numero_recu_tresor,omitempty"`
	AgentTresor       *string `json:"agent_tresor,omitempty"`
	BureauTresor      *string `json:"bureau_tresor,omitempty"`
}

// UpdatePaiementRequest represents request to update a payment
type UpdatePaiementRequest struct {
	Statut           *string `json:"statut,omitempty" validate:"omitempty,oneof=EN_COURS VALIDE REFUSE REMBOURSE"`
	ReferenceExterne *string `json:"reference_externe,omitempty"`
	CodeAutorisation *string `json:"code_autorisation,omitempty"`
	MotifRefus       *string `json:"motif_refus,omitempty"`
}

// ValidatePaiementRequest represents request to validate a payment
type ValidatePaiementRequest struct {
	CodeAutorisation string `json:"code_autorisation" validate:"required"`
}

// RefusePaiementRequest represents request to refuse a payment
type RefusePaiementRequest struct {
	MotifRefus string `json:"motif_refus" validate:"required"`
}

// RemboursementRequest represents request for a refund
type RemboursementRequest struct {
	Motif            string  `json:"motif" validate:"required"`
	MontantRembourse float64 `json:"montant_rembourse,omitempty"`
}

// ListPaiementsRequest represents filter for listing payments
type ListPaiementsRequest struct {
	ProcesVerbalID *string    `json:"proces_verbal_id,omitempty"`
	Statut         *string    `json:"statut,omitempty"`
	MoyenPaiement  *string    `json:"moyen_paiement,omitempty"`
	DateDebut      *time.Time `json:"date_debut,omitempty"`
	DateFin        *time.Time `json:"date_fin,omitempty"`
	MontantMin     *float64   `json:"montant_min,omitempty"`
	MontantMax     *float64   `json:"montant_max,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// PaiementResponse represents a payment in responses
type PaiementResponse struct {
	ID                 string                  `json:"id"`
	NumeroTransaction  string                  `json:"numero_transaction"`
	DatePaiement       time.Time               `json:"date_paiement"`
	Montant            float64                 `json:"montant"`
	MoyenPaiement      string                  `json:"moyen_paiement"`
	ReferenceExterne   string                  `json:"reference_externe,omitempty"`
	Statut             string                  `json:"statut"`
	CodeAutorisation   string                  `json:"code_autorisation,omitempty"`
	DetailsPaiement    string                  `json:"details_paiement,omitempty"`
	DateValidation     *time.Time              `json:"date_validation,omitempty"`
	MotifRefus         string                  `json:"motif_refus,omitempty"`
	// Champs Trésor Public
	NumeroRecuTresor   string                  `json:"numero_recu_tresor,omitempty"`
	AgentTresor        string                  `json:"agent_tresor,omitempty"`
	BureauTresor       string                  `json:"bureau_tresor,omitempty"`
	ProcesVerbal       *ProcesVerbalSummary    `json:"proces_verbal,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// ProcesVerbalSummary represents PV summary for payment response
type ProcesVerbalSummary struct {
	ID           string    `json:"id"`
	NumeroPV     string    `json:"numero_pv"`
	DateEmission time.Time `json:"date_emission"`
	MontantTotal float64   `json:"montant_total"`
	Statut       string    `json:"statut"`
}

// ListPaiementsResponse represents response for listing payments
type ListPaiementsResponse struct {
	Paiements []*PaiementResponse `json:"paiements"`
	Total     int                 `json:"total"`
}

// PaiementStatisticsResponse represents payment statistics
type PaiementStatisticsResponse struct {
	TotalPaiements     int                       `json:"total_paiements"`
	MontantTotal       float64                   `json:"montant_total"`
	MontantValide      float64                   `json:"montant_valide"`
	MontantEnCours     float64                   `json:"montant_en_cours"`
	MontantRembourse   float64                   `json:"montant_rembourse"`
	ParStatut          map[string]int            `json:"par_statut"`
	ParMoyenPaiement   map[string]float64        `json:"par_moyen_paiement"`
	EvolutionMensuelle []*MontantMensuel         `json:"evolution_mensuelle"`
}

// MontantMensuel represents monthly amount
type MontantMensuel struct {
	Mois    string  `json:"mois"`
	Montant float64 `json:"montant"`
	Nombre  int     `json:"nombre"`
}

// RecuTresorRequest represents request to generate a treasury receipt
type RecuTresorRequest struct {
	PaiementID   string `json:"paiement_id" validate:"required"`
	AgentTresor  string `json:"agent_tresor" validate:"required"`
	BureauTresor string `json:"bureau_tresor" validate:"required"`
}

// RecuTresorResponse represents a treasury receipt
type RecuTresorResponse struct {
	NumeroRecu       string    `json:"numero_recu"`
	DateEmission     time.Time `json:"date_emission"`
	Montant          float64   `json:"montant"`
	MontantEnLettres string    `json:"montant_en_lettres"`
	// Informations PV
	NumeroPV         string    `json:"numero_pv"`
	DatePV           time.Time `json:"date_pv"`
	// Informations contrevenant
	NomContrevenant  string    `json:"nom_contrevenant"`
	// Informations Trésor
	AgentTresor      string    `json:"agent_tresor"`
	BureauTresor     string    `json:"bureau_tresor"`
	// QR Code pour vérification
	QRCodeData       string    `json:"qr_code_data"`
	// Timestamps
	PaiementID       string    `json:"paiement_id"`
	CreatedAt        time.Time `json:"created_at"`
}
