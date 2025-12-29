package controle

import (
	"time"
)

// Request types

// InitialCheckOption represents a check option to create with the control/inspection
type InitialCheckOption struct {
	CheckItemID   string  `json:"check_item_id" validate:"required"`
	Resultat      string  `json:"resultat" validate:"required,oneof=PASS FAIL WARNING NOT_CHECKED"`
	Notes         *string `json:"notes,omitempty"`
	MontantAmende *int    `json:"montant_amende,omitempty"`
}

// CreateControleRequest represents the request to create a controle
type CreateControleRequest struct {
	// Date et localisation
	DateControle time.Time `json:"date_controle" validate:"required"`
	LieuControle string    `json:"lieu_controle" validate:"required"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	// Info contrôle
	TypeControle string  `json:"type_controle,omitempty"` // DOCUMENT, SECURITE, GENERAL, MIXTE
	Statut       string  `json:"statut,omitempty"`        // EN_COURS, TERMINE, CONFORME, NON_CONFORME
	Observations *string `json:"observations,omitempty"`
	// Agent
	AgentID        string  `json:"agent_id" validate:"required"`
	CommissariatID *string `json:"commissariat_id,omitempty"`
	// Données véhicule embarquées (dénormalisées)
	VehiculeImmatriculation string  `json:"vehicule_immatriculation" validate:"required"`
	VehiculeMarque          string  `json:"vehicule_marque" validate:"required"`
	VehiculeModele          string  `json:"vehicule_modele" validate:"required"`
	VehiculeType            string  `json:"vehicule_type" validate:"required,oneof=VOITURE SUV CAMION CAMIONNETTE MOTO BUS AUTRE"`
	VehiculeAnnee           *int    `json:"vehicule_annee,omitempty"`
	VehiculeCouleur         *string `json:"vehicule_couleur,omitempty"`
	VehiculeNumeroChassis   *string `json:"vehicule_numero_chassis,omitempty"`
	// Données conducteur embarquées (dénormalisées)
	ConducteurNumeroPermis string  `json:"conducteur_numero_permis" validate:"required"`
	ConducteurNom          string  `json:"conducteur_nom" validate:"required"`
	ConducteurPrenom       string  `json:"conducteur_prenom" validate:"required"`
	ConducteurTelephone    *string `json:"conducteur_telephone,omitempty"`
	ConducteurAdresse      *string `json:"conducteur_adresse,omitempty"`
	// Liens optionnels vers entités normalisées
	VehiculeID   *string `json:"vehicule_id,omitempty"`
	ConducteurID *string `json:"conducteur_id,omitempty"`
	// Initial check options (vérifications) - créées atomiquement avec le contrôle
	InitialOptions []InitialCheckOption `json:"initial_options,omitempty"`
	// Compteurs calculés (envoyés par le frontend)
	TotalVerifications  *int `json:"total_verifications,omitempty"`
	VerificationsOk     *int `json:"verifications_ok,omitempty"`
	VerificationsEchec  *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes *int `json:"montant_total_amendes,omitempty"`
}

// UpdateControleRequest represents the request to update a controle
type UpdateControleRequest struct {
	DateControle *time.Time `json:"date_controle,omitempty"`
	LieuControle *string    `json:"lieu_controle,omitempty"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	TypeControle *string    `json:"type_controle,omitempty"`
	Statut       *string    `json:"statut,omitempty"`
	Observations *string    `json:"observations,omitempty"`
	// Compteurs (depuis CheckOptions)
	TotalVerifications   *int `json:"total_verifications,omitempty"`
	VerificationsOk      *int `json:"verifications_ok,omitempty"`
	VerificationsEchec   *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes  *int `json:"montant_total_amendes,omitempty"`
}

// ListControlesRequest represents the request to list controles
type ListControlesRequest struct {
	AgentID                 *string    `json:"agent_id,omitempty"`
	VehiculeID              *string    `json:"vehicule_id,omitempty"`
	ConducteurID            *string    `json:"conducteur_id,omitempty"`
	CommissariatID          *string    `json:"commissariat_id,omitempty"`
	TypeControle            *string    `json:"type_controle,omitempty"`
	Statut                  *string    `json:"statut,omitempty"`
	LieuControle            *string    `json:"lieu_controle,omitempty"`
	VehiculeImmatriculation *string    `json:"vehicule_immatriculation,omitempty"`
	DateDebut               *time.Time `json:"date_debut,omitempty"`
	DateFin                 *time.Time `json:"date_fin,omitempty"`
	IsArchived              *bool      `json:"is_archived,omitempty"`
	Limit                   int        `json:"limit,omitempty"`
	Offset                  int        `json:"offset,omitempty"`
}

// Response types

// ControleResponse represents a controle in API responses
type ControleResponse struct {
	ID        string `json:"id"`
	Reference string `json:"reference,omitempty"`
	// Date et localisation
	DateControle time.Time `json:"date_controle"`
	LieuControle string    `json:"lieu_controle"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	// Info contrôle
	TypeControle string `json:"type_controle"`
	Statut       string `json:"statut"`
	Observations string `json:"observations,omitempty"`
	// Compteurs
	TotalVerifications  int `json:"total_verifications"`
	VerificationsOk     int `json:"verifications_ok"`
	VerificationsEchec  int `json:"verifications_echec"`
	MontantTotalAmendes int `json:"montant_total_amendes"`
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
	ConducteurNom          string `json:"conducteur_nom"`
	ConducteurPrenom       string `json:"conducteur_prenom"`
	ConducteurTelephone    string `json:"conducteur_telephone,omitempty"`
	ConducteurAdresse      string `json:"conducteur_adresse,omitempty"`
	// Relations (optionnelles)
	Agent             *AgentSummary        `json:"agent,omitempty"`
	Vehicule          *VehiculeSummary     `json:"vehicule,omitempty"`
	Conducteur        *ConducteurSummary   `json:"conducteur,omitempty"`
	Commissariat      *CommissariatSummary `json:"commissariat,omitempty"`
	Infractions       []*InfractionSummary `json:"infractions,omitempty"`
	NombreInfractions int                  `json:"nombre_infractions"`
	// Documents et éléments contrôlés
	DocumentsVerifies  []*DocumentVerifie `json:"documents_verifies,omitempty"`
	ElementsControles  []*ElementControle `json:"elements_controles,omitempty"`
	// PV et amende
	PV     *PVSummary     `json:"pv,omitempty"`
	Amende *AmendeSummary `json:"amende,omitempty"`
	// Recommandations, photos et suivi
	Recommandations []string        `json:"recommandations,omitempty"`
	Photos          []*PhotoControle `json:"photos,omitempty"`
	DateSuivi       *time.Time      `json:"date_suivi,omitempty"`
	Duree           string          `json:"duree,omitempty"` // Ex: "25 minutes"
	// Archivage
	IsArchived bool       `json:"is_archived"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgentSummary represents agent information in controle responses
type AgentSummary struct {
	ID           string `json:"id"`
	Matricule    string `json:"matricule"`
	Nom          string `json:"nom"`
	Prenom       string `json:"prenom"`
	Role         string `json:"role"`
	Grade        string `json:"grade,omitempty"`
	Telephone    string `json:"telephone,omitempty"`
	Email        string `json:"email,omitempty"`
	Commissariat string `json:"commissariat,omitempty"`
}

// VehiculeSummary represents vehicule information in controle responses
type VehiculeSummary struct {
	ID              string `json:"id"`
	Immatriculation string `json:"immatriculation"`
	Marque          string `json:"marque"`
	Modele          string `json:"modele"`
	TypeVehicule    string `json:"type_vehicule"`
	Annee           int    `json:"annee,omitempty"`
	Couleur         string `json:"couleur,omitempty"`
	NumeroSerie     string `json:"numero_serie,omitempty"`
}

// ConducteurSummary represents conducteur information in controle responses
type ConducteurSummary struct {
	ID             string  `json:"id"`
	Nom            string  `json:"nom"`
	Prenom         string  `json:"prenom"`
	NumeroPermis   string  `json:"numero_permis,omitempty"`
	PointsPermis   int     `json:"points_permis"`
	PermisValide   bool    `json:"permis_valide"`
	ValiditePermis *string `json:"validite_permis,omitempty"`
	CNI            string  `json:"cni,omitempty"`
	Telephone      string  `json:"telephone,omitempty"`
	Email          string  `json:"email,omitempty"`
	Adresse        string  `json:"adresse,omitempty"`
}

// CommissariatSummary represents commissariat information
type CommissariatSummary struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// InfractionSummary represents infraction information in controle responses
type InfractionSummary struct {
	ID             string    `json:"id"`
	NumeroPV       string    `json:"numero_pv,omitempty"`
	DateInfraction time.Time `json:"date_infraction"`
	TypeInfraction string    `json:"type_infraction"`
	MontantAmende  float64   `json:"montant_amende"`
	PointsRetires  int       `json:"points_retires"`
	Statut         string    `json:"statut"`
}

// DocumentVerifie represents a verified document in controle
type DocumentVerifie struct {
	Type     string  `json:"type"`     // CARTE_GRISE, ASSURANCE, CONTROLE_TECHNIQUE, PERMIS_CONDUIRE
	Statut   string  `json:"statut"`   // OK, NOK, N/A
	Details  string  `json:"details"`
	Validite *string `json:"validite,omitempty"`
	Photo    *string `json:"photo,omitempty"` // URL de la photo preuve (depuis CheckOption.evidence_file)
}

// ElementControle represents a controlled element in controle
type ElementControle struct {
	Type    string  `json:"type"`    // ECLAIRAGE, FREINAGE, PNEUMATIQUES, CEINTURES, EXTINCTEUR, TRIANGLE
	Statut  string  `json:"statut"`  // OK, NOK, N/A
	Details string  `json:"details"`
	Photo   *string `json:"photo,omitempty"` // URL de la photo preuve (depuis CheckOption.evidence_file)
}

// PVSummary represents PV information in controle responses
type PVSummary struct {
	ID           string    `json:"id"`
	Numero       string    `json:"numero"`
	DateEmission time.Time `json:"date_emission"`
	Infractions  []string  `json:"infractions"`
	Gravite      string    `json:"gravite"` // CLASSE_1, CLASSE_2, CLASSE_3, CLASSE_4
}

// AmendeSummary represents amende information in controle responses
type AmendeSummary struct {
	ID      string  `json:"id"`
	Numero  string  `json:"numero"`
	Montant float64 `json:"montant"`
	Statut  string  `json:"statut"` // EN_ATTENTE, PAYE
}

// PhotoControle represents a photo taken during controle
type PhotoControle struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	URL         string    `json:"url"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListControlesResponse represents the response for listing controles
type ListControlesResponse struct {
	Controles []*ControleResponse `json:"controles"`
	Total     int                 `json:"total"`
}

// ControleStatisticsResponse represents statistics for controles
type ControleStatisticsResponse struct {
	AgentID             *string        `json:"agent_id,omitempty"`
	Total               int            `json:"total"`
	EnCours             int            `json:"en_cours"`
	Termine             int            `json:"termine"`
	Conforme            int            `json:"conforme"`
	NonConforme         int            `json:"non_conforme"`
	ParType             map[string]int `json:"par_type"`
	ParJour             map[string]int `json:"par_jour"`
	InfractionsAvec     int            `json:"infractions_avec"`
	InfractionsSans     int            `json:"infractions_sans"`
	MontantTotalAmendes int            `json:"montant_total_amendes"`
	PeriodeDebut        *time.Time     `json:"periode_debut,omitempty"`
	PeriodeFin          *time.Time     `json:"periode_fin,omitempty"`
}

// ChangerStatutRequest represents request to change controle status
type ChangerStatutRequest struct {
	Statut       string  `json:"statut" validate:"required,oneof=EN_COURS TERMINE CONFORME NON_CONFORME"`
	Observations *string `json:"observations,omitempty"`
	// Compteurs
	TotalVerifications  *int `json:"total_verifications,omitempty"`
	VerificationsOk     *int `json:"verifications_ok,omitempty"`
	VerificationsEchec  *int `json:"verifications_echec,omitempty"`
	MontantTotalAmendes *int `json:"montant_total_amendes,omitempty"`
}

// StatisticsFilters represents filters for statistics endpoint
type StatisticsFilters struct {
	AgentID   *string    `json:"agent_id,omitempty"`
	DateDebut *time.Time `json:"date_debut,omitempty"`
	DateFin   *time.Time `json:"date_fin,omitempty"`
}

// GeneratePVRequest represents request to generate PV from controle
type GeneratePVRequest struct {
	Infractions []string `json:"infractions" validate:"required,min=1"`
}

// GeneratePVResponse represents the response after generating PV
type GeneratePVResponse struct {
	ID                 string    `json:"id"`
	NumeroPV           string    `json:"numero_pv"`
	DateEmission       time.Time `json:"date_emission"`
	MontantTotal       float64   `json:"montant_total"`
	DateLimitePaiement time.Time `json:"date_limite_paiement"`
	Statut             string    `json:"statut"`
	ControleID         string    `json:"controle_id"`
	NbInfractions      int       `json:"nb_infractions"`
}
