package inspection

import (
	"time"
)

// Request types

// InitialCheckOption represents a check option to create with the inspection
type InitialCheckOption struct {
	CheckItemID   string  `json:"check_item_id" validate:"required"`
	Resultat      string  `json:"resultat" validate:"required,oneof=PASS FAIL WARNING NOT_CHECKED"`
	Notes         *string `json:"notes,omitempty"`
	MontantAmende *int    `json:"montant_amende,omitempty"`
}

// CreateInspectionRequest represents the request to create an inspection
type CreateInspectionRequest struct {
	DateInspection time.Time `json:"date_inspection" validate:"required"`
	InspecteurID   string    `json:"inspecteur_id" validate:"required"`
	// Données véhicule embarquées (dénormalisées)
	VehiculeImmatriculation string  `json:"vehicule_immatriculation" validate:"required"`
	VehiculeMarque          string  `json:"vehicule_marque" validate:"required"`
	VehiculeModele          string  `json:"vehicule_modele" validate:"required"`
	VehiculeType            string  `json:"vehicule_type" validate:"required,oneof=VOITURE MOTO CAMION BUS CAMIONNETTE TRACTEUR AUTRE"`
	VehiculeAnnee           *int    `json:"vehicule_annee,omitempty"`
	VehiculeCouleur         *string `json:"vehicule_couleur,omitempty"`
	VehiculeNumeroChassis   *string `json:"vehicule_numero_chassis,omitempty"`
	// Données conducteur embarquées (dénormalisées)
	ConducteurNumeroPermis string  `json:"conducteur_numero_permis" validate:"required"`
	ConducteurPrenom       string  `json:"conducteur_prenom" validate:"required"`
	ConducteurNom          string  `json:"conducteur_nom" validate:"required"`
	ConducteurTelephone    *string `json:"conducteur_telephone,omitempty"`
	ConducteurAdresse      *string `json:"conducteur_adresse,omitempty"`
	ConducteurTypePiece    *string `json:"conducteur_type_piece,omitempty"` // CNI, PASSEPORT, CARTE_SEJOUR
	ConducteurNumeroPiece  *string `json:"conducteur_numero_piece,omitempty"`
	// Données assurance
	AssuranceCompagnie      *string    `json:"assurance_compagnie,omitempty"`
	AssuranceNumeroPolice   *string    `json:"assurance_numero_police,omitempty"`
	AssuranceDateExpiration *time.Time `json:"assurance_date_expiration,omitempty"`
	AssuranceStatut         *string    `json:"assurance_statut,omitempty"` // ACTIVE, EXPIREE, SUSPENDUE, ANNULEE, INCONNU
	// Localisation
	LieuInspection *string  `json:"lieu_inspection,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	// Liens optionnels vers entités normalisées
	VehiculeID     *string `json:"vehicule_id,omitempty"`
	CommissariatID *string `json:"commissariat_id,omitempty"`
	// Notes
	Observations *string `json:"observations,omitempty"`
	// Initial check options (vérifications) - créées atomiquement avec l'inspection
	InitialOptions []InitialCheckOption `json:"initial_options,omitempty"`
	// Compteurs calculés (envoyés par le frontend)
	TotalVerifications     *int `json:"total_verifications,omitempty"`
	VerificationsOk        *int `json:"verifications_ok,omitempty"`
	VerificationsAttention *int `json:"verifications_attention,omitempty"`
	VerificationsEchec     *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes    *int `json:"montant_total_amendes,omitempty"`
}

// UpdateInspectionRequest represents the request to update an inspection
type UpdateInspectionRequest struct {
	DateInspection *time.Time `json:"date_inspection,omitempty"`
	Statut         *string    `json:"statut,omitempty"` // EN_ATTENTE, EN_COURS, TERMINE, CONFORME, NON_CONFORME
	Observations   *string    `json:"observations,omitempty"`
	// Localisation
	LieuInspection *string  `json:"lieu_inspection,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	// Mise à jour assurance
	AssuranceCompagnie      *string    `json:"assurance_compagnie,omitempty"`
	AssuranceNumeroPolice   *string    `json:"assurance_numero_police,omitempty"`
	AssuranceDateExpiration *time.Time `json:"assurance_date_expiration,omitempty"`
	AssuranceStatut         *string    `json:"assurance_statut,omitempty"`
	// Compteurs (depuis CheckOptions)
	TotalVerifications     *int `json:"total_verifications,omitempty"`
	VerificationsOk        *int `json:"verifications_ok,omitempty"`
	VerificationsAttention *int `json:"verifications_attention,omitempty"`
	VerificationsEchec     *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes    *int `json:"montant_total_amendes,omitempty"`
}

// ListInspectionsRequest represents the request to list inspections
type ListInspectionsRequest struct {
	VehiculeID              *string    `json:"vehicule_id,omitempty"`
	InspecteurID            *string    `json:"inspecteur_id,omitempty"`
	CommissariatID          *string    `json:"commissariat_id,omitempty"`
	Statut                  *string    `json:"statut,omitempty"`
	AssuranceStatut         *string    `json:"assurance_statut,omitempty"`
	VehiculeImmatriculation *string    `json:"vehicule_immatriculation,omitempty"`
	DateDebut               *time.Time `json:"date_debut,omitempty"`
	DateFin                 *time.Time `json:"date_fin,omitempty"`
	Search                  *string    `json:"search,omitempty"`
	Limit                   int        `json:"limit,omitempty"`
	Offset                  int        `json:"offset,omitempty"`
}

// Response types

// InspectionResponse represents an inspection in API responses
type InspectionResponse struct {
	ID             string    `json:"id"`
	Numero         string    `json:"numero"`
	DateInspection time.Time `json:"date_inspection"`
	Statut         string    `json:"statut"`
	Observations   string    `json:"observations,omitempty"`
	// Compteurs
	TotalVerifications     int `json:"total_verifications"`
	VerificationsOk        int `json:"verifications_ok"`
	VerificationsAttention int `json:"verifications_attention"`
	VerificationsEchec     int `json:"verifications_echec"`
	MontantTotalAmendes    int `json:"montant_total_amendes"`
	// Données véhicule embarquées
	VehiculeImmatriculation string `json:"vehicule_immatriculation"`
	VehiculeMarque          string `json:"vehicule_marque"`
	VehiculeModele          string `json:"vehicule_modele"`
	VehiculeAnnee           int    `json:"vehicule_annee,omitempty"`
	VehiculeCouleur         string `json:"vehicule_couleur,omitempty"`
	VehiculeNumeroChassis   string `json:"vehicule_numero_chassis,omitempty"`
	VehiculeType            string `json:"vehicule_type"`
	// Données conducteur embarquées
	ConducteurNumeroPermis string `json:"conducteur_numero_permis"`
	ConducteurPrenom       string `json:"conducteur_prenom"`
	ConducteurNom          string `json:"conducteur_nom"`
	ConducteurTelephone    string `json:"conducteur_telephone,omitempty"`
	ConducteurAdresse      string `json:"conducteur_adresse,omitempty"`
	ConducteurTypePiece    string `json:"conducteur_type_piece,omitempty"`
	ConducteurNumeroPiece  string `json:"conducteur_numero_piece,omitempty"`
	// Données assurance
	AssuranceCompagnie      string     `json:"assurance_compagnie,omitempty"`
	AssuranceNumeroPolice   string     `json:"assurance_numero_police,omitempty"`
	AssuranceDateExpiration *time.Time `json:"assurance_date_expiration,omitempty"`
	AssuranceStatut         string     `json:"assurance_statut"`
	// Localisation
	LieuInspection string   `json:"lieu_inspection,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	// Relations (optionnelles, pour données normalisées liées)
	Vehicule     *VehiculeSummary     `json:"vehicule,omitempty"`
	Inspecteur   *AgentSummary        `json:"inspecteur,omitempty"`
	Commissariat *CommissariatSummary `json:"commissariat,omitempty"`
	ProcesVerbal *PVSummary           `json:"proces_verbal,omitempty"`
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VehiculeSummary represents vehicule info in responses
type VehiculeSummary struct {
	ID              string `json:"id"`
	Immatriculation string `json:"immatriculation"`
	Marque          string `json:"marque"`
	Modele          string `json:"modele"`
	ProprietaireNom string `json:"proprietaire_nom,omitempty"`
}

// AgentSummary represents agent info in responses
type AgentSummary struct {
	ID        string `json:"id"`
	Matricule string `json:"matricule"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
}

// CommissariatSummary represents commissariat info in responses
type CommissariatSummary struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// PVSummary represents PV info in responses
type PVSummary struct {
	ID           string    `json:"id"`
	NumeroPV     string    `json:"numero_pv"`
	DateEmission time.Time `json:"date_emission"`
	MontantTotal float64   `json:"montant_total"`
	Statut       string    `json:"statut"`
}

// ListInspectionsResponse represents the response for listing inspections
type ListInspectionsResponse struct {
	Inspections []*InspectionResponse `json:"inspections"`
	Total       int                   `json:"total"`
}

// ChangerStatutRequest represents request to change inspection status
type ChangerStatutRequest struct {
	Statut       string  `json:"statut" validate:"required,oneof=EN_ATTENTE EN_COURS TERMINE CONFORME NON_CONFORME"`
	Observations *string `json:"observations,omitempty"`
	// Compteurs
	TotalVerifications     *int `json:"total_verifications,omitempty"`
	VerificationsOk        *int `json:"verifications_ok,omitempty"`
	VerificationsAttention *int `json:"verifications_attention,omitempty"`
	VerificationsEchec     *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes    *int `json:"montant_total_amendes,omitempty"`
}

// InspectionStatisticsResponse represents statistics for inspections
type InspectionStatisticsResponse struct {
	Total               int            `json:"total"`
	EnAttente           int            `json:"en_attente"`
	EnCours             int            `json:"en_cours"`
	Termine             int            `json:"termine"`
	Conforme            int            `json:"conforme"`
	NonConforme         int            `json:"non_conforme"`
	AssuranceInvalide   int            `json:"assurance_invalide"`
	ParStatut           map[string]int `json:"par_statut"`
	TauxConformite      float64        `json:"taux_conformite"`
	MontantTotalAmendes int            `json:"montant_total_amendes"`
}
