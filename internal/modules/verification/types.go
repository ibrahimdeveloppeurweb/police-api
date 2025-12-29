package verification

import (
	"time"
)

// CheckItemResponse represents a check item (catalogue) in API responses
type CheckItemResponse struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Nom          string `json:"nom"`
	Categorie    string `json:"categorie"` // DOCUMENT, SAFETY, EQUIPMENT, LIGHTING, VISIBILITY
	Description  string `json:"description,omitempty"`
	Icon         string `json:"icon"`
	Obligatoire  bool   `json:"obligatoire"`
	Actif        bool   `json:"actif"`
	Ordre        int    `json:"ordre"`
	MontantAmende int   `json:"montant_amende"`
	PointsRetrait int   `json:"points_retrait"`
	ApplicableA   string `json:"applicable_a"` // INSPECTION, CONTROL, BOTH
}

// CheckOptionResponse represents a verification result in API responses
type CheckOptionResponse struct {
	ID            string    `json:"id"`
	CheckItemID   string    `json:"check_item_id"`
	CheckItemCode string    `json:"check_item_code"`
	CheckItemNom  string    `json:"check_item_nom"`
	Categorie     string    `json:"categorie"`
	Resultat      string    `json:"resultat"` // PASS, FAIL, WARNING, NOT_CHECKED
	Notes         string    `json:"notes,omitempty"`
	MontantAmende int       `json:"montant_amende"`
	DateVerification time.Time `json:"date_verification"`
	// Info du CheckItem associé
	Icon          string `json:"icon,omitempty"`
	Obligatoire   bool   `json:"obligatoire"`
	PointsRetrait int    `json:"points_retrait"`
}

// ListVerificationsResponse represents the response for listing verifications
type ListVerificationsResponse struct {
	Verifications []*CheckOptionResponse `json:"verifications"`
	Total         int                    `json:"total"`
	// Résumé
	TotalOk       int `json:"total_ok"`
	TotalEchec    int `json:"total_echec"`
	TotalAttention int `json:"total_attention"`
	TotalNonVerifie int `json:"total_non_verifie"`
	MontantTotal  int `json:"montant_total"`
}

// ListCheckItemsResponse represents the response for listing check items (catalogue)
type ListCheckItemsResponse struct {
	Items []*CheckItemResponse `json:"items"`
	Total int                  `json:"total"`
}

// CreateCheckOptionRequest represents request to create/update a verification
type CreateCheckOptionRequest struct {
	CheckItemID   string  `json:"check_item_id" validate:"required"`
	Resultat      string  `json:"resultat" validate:"required,oneof=PASS FAIL WARNING NOT_CHECKED"`
	Notes         *string `json:"notes,omitempty"`
	MontantAmende *int    `json:"montant_amende,omitempty"`
}

// BatchCheckOptionsRequest represents request to save multiple verifications at once
type BatchCheckOptionsRequest struct {
	Verifications []CreateCheckOptionRequest `json:"verifications" validate:"required,min=1"`
}
