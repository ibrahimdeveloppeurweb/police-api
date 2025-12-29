package document

import (
	"time"
)

// UploadDocumentRequest represents request to upload a document
type UploadDocumentRequest struct {
	TypeDocument   string  `json:"type_document" validate:"required,oneof=PHOTO PERMIS CARTE_GRISE ASSURANCE CONSTAT PV AUTRE"`
	Description    *string `json:"description,omitempty"`
	Public         bool    `json:"public"`
	ControleID     *string `json:"controle_id,omitempty"`
	InfractionID   *string `json:"infraction_id,omitempty"`
	ProcesVerbalID *string `json:"proces_verbal_id,omitempty"`
	RecoursID      *string `json:"recours_id,omitempty"`
}

// UpdateDocumentRequest represents request to update a document
type UpdateDocumentRequest struct {
	NomOriginal  *string `json:"nom_original,omitempty"`
	TypeDocument *string `json:"type_document,omitempty" validate:"omitempty,oneof=PHOTO PERMIS CARTE_GRISE ASSURANCE CONSTAT PV AUTRE"`
	Description  *string `json:"description,omitempty"`
	Public       *bool   `json:"public,omitempty"`
}

// ListDocumentsRequest represents filter for listing documents
type ListDocumentsRequest struct {
	TypeDocument   *string    `json:"type_document,omitempty"`
	Public         *bool      `json:"public,omitempty"`
	UploadedByID   *string    `json:"uploaded_by_id,omitempty"`
	ControleID     *string    `json:"controle_id,omitempty"`
	InfractionID   *string    `json:"infraction_id,omitempty"`
	ProcesVerbalID *string    `json:"proces_verbal_id,omitempty"`
	RecoursID      *string    `json:"recours_id,omitempty"`
	DateDebut      *time.Time `json:"date_debut,omitempty"`
	DateFin        *time.Time `json:"date_fin,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// DocumentResponse represents a document in responses
type DocumentResponse struct {
	ID             string           `json:"id"`
	NomFichier     string           `json:"nom_fichier"`
	NomOriginal    string           `json:"nom_original"`
	TypeMime       string           `json:"type_mime"`
	Taille         int64            `json:"taille"`
	TailleFormatee string           `json:"taille_formatee"`
	TypeDocument   string           `json:"type_document"`
	Description    string           `json:"description,omitempty"`
	Public         bool             `json:"public"`
	URL            string           `json:"url,omitempty"`
	UploadedBy     *UploaderSummary `json:"uploaded_by,omitempty"`
	ControleID     *string          `json:"controle_id,omitempty"`
	InfractionID   *string          `json:"infraction_id,omitempty"`
	ProcesVerbalID *string          `json:"proces_verbal_id,omitempty"`
	RecoursID      *string          `json:"recours_id,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// UploaderSummary represents uploader summary for document response
type UploaderSummary struct {
	ID        string `json:"id"`
	Matricule string `json:"matricule"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
}

// ListDocumentsResponse represents response for listing documents
type ListDocumentsResponse struct {
	Documents []*DocumentResponse `json:"documents"`
	Total     int                 `json:"total"`
}

// DocumentStatisticsResponse represents document statistics
type DocumentStatisticsResponse struct {
	TotalDocuments    int                `json:"total_documents"`
	TailleTotale      int64              `json:"taille_totale"`
	TailleTotaleText  string             `json:"taille_totale_text"`
	ParType           map[string]int     `json:"par_type"`
	ParMois           map[string]int     `json:"par_mois"`
	DocumentsPublics  int                `json:"documents_publics"`
	DocumentsPrives   int                `json:"documents_prives"`
}
