package plainte

import (
	"time"
)

// Request types

// SuspectRequest represents a suspect in the plainte
type SuspectRequest struct {
	Nom         string  `json:"nom" validate:"required"`
	Prenom      string  `json:"prenom" validate:"required"`
	Description *string `json:"description,omitempty"`
	Adresse     *string `json:"adresse,omitempty"`
}

// TemoinRequest represents a witness in the plainte
type TemoinRequest struct {
	Nom       string  `json:"nom" validate:"required"`
	Prenom    string  `json:"prenom" validate:"required"`
	Telephone *string `json:"telephone,omitempty"`
	Adresse   *string `json:"adresse,omitempty"`
}

// CreatePlainteRequest represents the request to create a plainte
type CreatePlainteRequest struct {
	TypePlainte        string           `json:"type_plainte" validate:"required"`
	Description        *string          `json:"description,omitempty"`
	PlaignantNom       string           `json:"plaignant_nom" validate:"required"`
	PlaignantPrenom    string           `json:"plaignant_prenom" validate:"required"`
	PlaignantTelephone *string          `json:"plaignant_telephone,omitempty"`
	PlaignantAdresse   *string          `json:"plaignant_adresse,omitempty"`
	PlaignantEmail     *string          `json:"plaignant_email,omitempty"`
	LieuFaits          *string          `json:"lieu_faits,omitempty"`
	DateFaits          *time.Time       `json:"date_faits,omitempty"`
	Priorite           *string          `json:"priorite,omitempty"`
	Observations       *string          `json:"observations,omitempty"`
	CommissariatID     *string          `json:"commissariat_id,omitempty"`
	AgentAssigneID     *string          `json:"agent_assigne_id,omitempty"`
	Suspects           []SuspectRequest `json:"suspects,omitempty"`
	Temoins            []TemoinRequest  `json:"temoins,omitempty"`
}

// UpdatePlainteRequest represents the request to update a plainte
type UpdatePlainteRequest struct {
	TypePlainte        *string    `json:"type_plainte,omitempty"`
	Description        *string    `json:"description,omitempty"`
	PlaignantNom       *string    `json:"plaignant_nom,omitempty"`
	PlaignantPrenom    *string    `json:"plaignant_prenom,omitempty"`
	PlaignantTelephone *string    `json:"plaignant_telephone,omitempty"`
	PlaignantAdresse   *string    `json:"plaignant_adresse,omitempty"`
	PlaignantEmail     *string    `json:"plaignant_email,omitempty"`
	LieuFaits          *string    `json:"lieu_faits,omitempty"`
	DateFaits          *time.Time `json:"date_faits,omitempty"`
	Priorite           *string    `json:"priorite,omitempty"`
	Statut             *string    `json:"statut,omitempty"`
	EtapeActuelle      *string    `json:"etape_actuelle,omitempty"`
	Observations       *string    `json:"observations,omitempty"`
	DecisionFinale     *string    `json:"decision_finale,omitempty"`
	CommissariatID     *string    `json:"commissariat_id,omitempty"`
	AgentAssigneID     *string    `json:"agent_assigne_id,omitempty"`
}

// ListPlaintesRequest represents the request to list plaintes
type ListPlaintesRequest struct {
	TypePlainte    *string    `json:"type_plainte,omitempty"`
	Statut         *string    `json:"statut,omitempty"`
	Priorite       *string    `json:"priorite,omitempty"`
	EtapeActuelle  *string    `json:"etape_actuelle,omitempty"`
	CommissariatID *string    `json:"commissariat_id,omitempty"`
	AgentAssigneID *string    `json:"agent_assigne_id,omitempty"`
	DateDebut      *time.Time `json:"date_debut,omitempty"`
	DateFin        *time.Time `json:"date_fin,omitempty"`
	Search         *string    `json:"search,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// Response types

// SuspectResponse represents a suspect in API responses
type SuspectResponse struct {
	ID          string  `json:"id"`
	Nom         string  `json:"nom"`
	Prenom      string  `json:"prenom"`
	Description *string `json:"description,omitempty"`
	Adresse     *string `json:"adresse,omitempty"`
}

// TemoinResponse represents a witness in API responses
type TemoinResponse struct {
	ID        string  `json:"id"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
	Adresse   *string `json:"adresse,omitempty"`
}

// PlainteResponse represents a plainte in API responses
type PlainteResponse struct {
	ID                 string               `json:"id"`
	Numero             string               `json:"numero"`
	TypePlainte        string               `json:"type_plainte"`
	Description        string               `json:"description,omitempty"`
	PlaignantNom       string               `json:"plaignant_nom"`
	PlaignantPrenom    string               `json:"plaignant_prenom"`
	PlaignantTelephone string               `json:"plaignant_telephone,omitempty"`
	PlaignantAdresse   string               `json:"plaignant_adresse,omitempty"`
	PlaignantEmail     string               `json:"plaignant_email,omitempty"`
	DateDepot          time.Time            `json:"date_depot"`
	DateResolution     *time.Time           `json:"date_resolution,omitempty"`
	EtapeActuelle      string               `json:"etape_actuelle"`
	Priorite           string               `json:"priorite"`
	Statut             string               `json:"statut"`
	DelaiSLA           string               `json:"delai_sla,omitempty"`
	SLADepasse         bool                 `json:"sla_depasse"`
	LieuFaits          string               `json:"lieu_faits,omitempty"`
	DateFaits          *time.Time           `json:"date_faits,omitempty"`
	Observations       string               `json:"observations,omitempty"`
	DecisionFinale     string               `json:"decision_finale,omitempty"`
	Commissariat       *CommissariatSummary `json:"commissariat,omitempty"`
	AgentAssigne       *AgentSummary        `json:"agent_assigne,omitempty"`
	Suspects           []SuspectResponse    `json:"suspects,omitempty"`
	Temoins            []TemoinResponse     `json:"temoins,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// CommissariatSummary represents commissariat info in responses
type CommissariatSummary struct {
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// AgentSummary represents agent info in responses
type AgentSummary struct {
	ID        string `json:"id"`
	Matricule string `json:"matricule"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
}

// ListPlaintesResponse represents the response for listing plaintes
type ListPlaintesResponse struct {
	Plaintes []*PlainteResponse `json:"plaintes"`
	Total    int                `json:"total"`
}

// ChangerEtapeRequest represents request to change plainte workflow step
type ChangerEtapeRequest struct {
	Etape        string  `json:"etape" validate:"required,oneof=DEPOT ENQUETE RESOLUTION CLOTURE"`
	Observations *string `json:"observations,omitempty"`
}

// ChangerStatutRequest represents request to change plainte status
type ChangerStatutRequest struct {
	Statut         string  `json:"statut" validate:"required,oneof=EN_COURS RESOLU CLASSE TRANSFERE"`
	DecisionFinale *string `json:"decision_finale,omitempty"`
}

// AssignerAgentRequest represents request to assign agent to plainte
type AssignerAgentRequest struct {
	AgentID string `json:"agent_id" validate:"required"`
}

// StatisticsRequest represents request for statistics
type StatisticsRequest struct {
	CommissariatID *string    `json:"commissariat_id,omitempty"`
	DateDebut      *time.Time `json:"date_debut,omitempty"`
	DateFin        *time.Time `json:"date_fin,omitempty"`
}

// PlainteStatisticsResponse represents statistics for plaintes
type PlainteStatisticsResponse struct {
	Total           int            `json:"total"`
	EnCours         int            `json:"en_cours"`
	Resolues        int            `json:"resolues"`
	Classees        int            `json:"classees"`
	Transferees     int            `json:"transferees"`
	ParType         map[string]int `json:"par_type"`
	ParPriorite     map[string]int `json:"par_priorite"`
	ParEtape        map[string]int `json:"par_etape"`
	SLADepasse      int            `json:"sla_depasse"`
	DelaiMoyenJours float64        `json:"delai_moyen_jours"`
}

// AlerteResponse represents an alert for a plainte
type AlerteResponse struct {
	ID            string `json:"id"`
	PlainteID     string `json:"plainte_id"`
	PlainteNumero string `json:"plainte_numero"`
	TypeAlerte    string `json:"type_alerte"`
	Message       string `json:"message"`
	Niveau        string `json:"niveau"`
	JoursRetard   *int   `json:"jours_retard,omitempty"`
}

// TopAgentResponse represents a top performing agent
type TopAgentResponse struct {
	ID               string  `json:"id"`
	Nom              string  `json:"nom"`
	Prenom           string  `json:"prenom"`
	Matricule        string  `json:"matricule"`
	PlaintesTraitees int     `json:"plaintes_traitees"`
	PlaintesResolues int     `json:"plaintes_resolues"`
	Score            float64 `json:"score"`
	DelaiMoyen       float64 `json:"delai_moyen"`
}

// PreuveResponse represents a preuve for a plainte
type PreuveResponse struct {
	ID                string    `json:"id"`
	NumeroPiece       string    `json:"numero_piece"`
	Type              string    `json:"type"`
	Description       string    `json:"description"`
	LieuConservation  *string   `json:"lieu_conservation,omitempty"`
	DateCollecte      time.Time `json:"date_collecte"`
	CollectePar       *string   `json:"collecte_par,omitempty"`
	Photos            []string  `json:"photos,omitempty"`
	HashVerification  *string   `json:"hash_verification,omitempty"`
	ExpertiseDemandee bool      `json:"expertise_demandee"`
	ExpertiseType     *string   `json:"expertise_type,omitempty"`
	ExpertiseResultat *string   `json:"expertise_resultat,omitempty"`
	Statut            string    `json:"statut"`
	CreatedAt         time.Time `json:"created_at"`
}

// AddPreuveRequest represents request to add a preuve
type AddPreuveRequest struct {
	NumeroPiece       string    `json:"numero_piece" validate:"required"`
	Type              string    `json:"type" validate:"required,oneof=MATERIELLE NUMERIQUE TESTIMONIALE DOCUMENTAIRE"`
	Description       string    `json:"description" validate:"required"`
	LieuConservation  *string   `json:"lieu_conservation,omitempty"`
	TypeCollecte      string    `json:"statut" validate:"required,oneof=COLLECTEE EN_ANALYSE ANALYSEE RETOURNEE"`
	DateCollecte      time.Time `json:"date_collecte"`
	CollectePar       *string   `json:"collecte_par,omitempty"`
	ExpertiseDemandee bool      `json:"expertise_demandee"`
	ExpertiseType     *string   `json:"expertise_type,omitempty"`
}

// ActeEnqueteResponse represents an acte d'enquête
type ActeEnqueteResponse struct {
	ID                 string    `json:"id"`
	Type               string    `json:"type"`
	Date               time.Time `json:"date"`
	Heure              *string   `json:"heure,omitempty"`
	Duree              *string   `json:"duree,omitempty"`
	Lieu               *string   `json:"lieu,omitempty"`
	OfficierCharge     string    `json:"officier_charge"`
	Description        string    `json:"description"`
	PVNumero           *string   `json:"pv_numero,omitempty"`
	MandatNumero       *string   `json:"mandat_numero,omitempty"`
	PersonnesPresentes []string  `json:"personnes_presentes,omitempty"`
	ObjetsSaisis       []string  `json:"objets_saisis,omitempty"`
	Conclusions        *string   `json:"conclusions,omitempty"`
	DocumentsJoints    []string  `json:"documents_joints,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// AddActeEnqueteRequest represents request to add an acte d'enquête
type AddActeEnqueteRequest struct {
	Type           string    `json:"type" validate:"required,oneof=AUDITION PERQUISITION EXPERTISE GARDE_A_VUE CONFRONTATION RECONSTITUTION"`
	Date           time.Time `json:"date" validate:"required"`
	Heure          *string   `json:"heure,omitempty"`
	Duree          *string   `json:"duree,omitempty"`
	Lieu           *string   `json:"lieu,omitempty"`
	OfficierCharge string    `json:"officier_charge" validate:"required"`
	Description    string    `json:"description" validate:"required"`
	PVNumero       *string   `json:"pv_numero,omitempty"`
	MandatNumero   *string   `json:"mandat_numero,omitempty"`
}

// TimelineEventResponse represents a timeline event
type TimelineEventResponse struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Heure       *string   `json:"heure,omitempty"`
	Type        string    `json:"type"`
	Titre       string    `json:"titre"`
	Description string    `json:"description"`
	Acteur      *string   `json:"acteur,omitempty"`
	Statut      *string   `json:"statut,omitempty"`
	Documents   []string  `json:"documents,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AddTimelineEventRequest represents request to add a timeline event
type AddTimelineEventRequest struct {
	Date        time.Time `json:"date" validate:"required"`
	Heure       *string   `json:"heure,omitempty"`
	Type        string    `json:"type" validate:"required,oneof=DEPOT AUDITION PERQUISITION EXPERTISE DECISION AUTRE"`
	Titre       string    `json:"titre" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Acteur      *string   `json:"acteur,omitempty"`
	Statut      *string   `json:"statut,omitempty"`
}


// EnqueteResponse represents an enquête
type EnqueteResponse struct {
	ID                   string     `json:"id"`
	Type                 string     `json:"type"`
	OfficierCharge       string     `json:"officier_charge"`
	DateDebut            time.Time  `json:"date_debut"`
	DateFin              *time.Time `json:"date_fin,omitempty"`
	Lieu                 *string    `json:"lieu,omitempty"`
	Description          string     `json:"description"`
	Resultats            *string    `json:"resultats,omitempty"`
	PersonnesInterrogees []string   `json:"personnes_interrogees,omitempty"`
	PreuvesCollectees    []string   `json:"preuves_collectees,omitempty"`
	Conclusions          *string    `json:"conclusions,omitempty"`
	Statut               string     `json:"statut"`
	Documents            []string   `json:"documents,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// AddEnqueteRequest represents request to add an enquête
type AddEnqueteRequest struct {
	Type           string    `json:"type" validate:"required,oneof=AUDITION PERQUISITION EXPERTISE SURVEILLANCE AUTRE"`
	OfficierCharge string    `json:"officier_charge" validate:"required"`
	DateDebut      time.Time `json:"date_debut" validate:"required"`
	Lieu           *string   `json:"lieu,omitempty"`
	Description    string    `json:"description" validate:"required"`
}

// DecisionResponse represents a decision
type DecisionResponse struct {
	ID                string     `json:"id"`
	Type              string     `json:"type"`
	DateDecision      time.Time  `json:"date_decision"`
	Autorite          string     `json:"autorite"`
	Description       string     `json:"description"`
	Motivation        *string    `json:"motivation,omitempty"`
	Dispositions      []string   `json:"dispositions,omitempty"`
	Suites            *string    `json:"suites,omitempty"`
	DocumentReference *string    `json:"document_reference,omitempty"`
	Notifiee          bool       `json:"notifiee"`
	DateNotification  *time.Time `json:"date_notification,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// AddDecisionRequest represents request to add a decision
type AddDecisionRequest struct {
	Type         string    `json:"type" validate:"required,oneof=CLASSEMENT POURSUITE RENVOI ACQUITTEMENT CONDAMNATION NON_LIEU AUTRE"`
	DateDecision time.Time `json:"date_decision" validate:"required"`
	Autorite     string    `json:"autorite" validate:"required"`
	Description  string    `json:"description" validate:"required"`
	Motivation   *string   `json:"motivation,omitempty"`
}

// HistoriqueResponse represents a historique entry
type HistoriqueResponse struct {
	ID             string    `json:"id"`
	TypeChangement string    `json:"type_changement"`
	ChampModifie   string    `json:"champ_modifie"`
	AncienneValeur *string   `json:"ancienne_valeur,omitempty"`
	NouvelleValeur string    `json:"nouvelle_valeur"`
	Commentaire    *string   `json:"commentaire,omitempty"`
	AuteurNom      *string   `json:"auteur_nom,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
