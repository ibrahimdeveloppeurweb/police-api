package alertes

import "time"

// TypeAlerte représente les types d'alertes
type TypeAlerte string

const (
	TypeAlerteVehiculeVole     TypeAlerte = "VEHICULE_VOLE"
	TypeAlerteSuspectRecherche TypeAlerte = "SUSPECT_RECHERCHE"
	TypeAlerteUrgenceSecurite  TypeAlerte = "URGENCE_SECURITE"
	TypeAlerteAlerteGenerale   TypeAlerte = "ALERTE_GENERALE"
	TypeAlerteMaintenanceSysteme TypeAlerte = "MAINTENANCE_SYSTEME"
	TypeAlerteAccident         TypeAlerte = "ACCIDENT"
	TypeAlerteIncendie         TypeAlerte = "INCENDIE"
	TypeAlerteAggression       TypeAlerte = "AGGRESSION"
	TypeAlerteAmber            TypeAlerte = "AMBER"
	TypeAlerteAutre            TypeAlerte = "AUTRE"
)

// NiveauAlerte représente le niveau de gravité
type NiveauAlerte string

const (
	NiveauAlerteFaible   NiveauAlerte = "FAIBLE"
	NiveauAlerteMoyen    NiveauAlerte = "MOYEN"
	NiveauAlerteEleve    NiveauAlerte = "ELEVE"
	NiveauAlerteCritique NiveauAlerte = "CRITIQUE"
)

// StatutAlerte représente le statut de l'alerte
type StatutAlerte string

const (
	StatutAlerteActive   StatutAlerte = "ACTIVE"
	StatutAlerteResolue  StatutAlerte = "RESOLUE"
	StatutAlerteArchivee StatutAlerte = "ARCHIVEE"
)

// StatutIntervention représente le statut de l'intervention
type StatutIntervention string

const (
	StatutInterventionEnAttente  StatutIntervention = "EN_ATTENTE"
	StatutInterventionEnCours    StatutIntervention = "EN_COURS"
	StatutInterventionTerminee   StatutIntervention = "TERMINEE"
	StatutInterventionAnnulee    StatutIntervention = "ANNULEE"
)

// PersonneConcernee représente une personne concernée par l'alerte
type PersonneConcernee struct {
	Nom         string  `json:"nom" validate:"required"`
	Telephone   *string `json:"telephone,omitempty"`
	Relation    *string `json:"relation,omitempty"`
	Description *string `json:"description,omitempty"`
}

// VehiculeAlerte représente un véhicule concerné par l'alerte
type VehiculeAlerte struct {
	Immatriculation string  `json:"immatriculation" validate:"required"`
	Marque          *string `json:"marque,omitempty"`
	Modele          *string `json:"modele,omitempty"`
	Couleur         *string `json:"couleur,omitempty"`
	Annee           *string `json:"annee,omitempty"`
}

// Suspect représente un suspect recherché
type Suspect struct {
	Nom          string  `json:"nom" validate:"required"`
	Description  string  `json:"description" validate:"required"`
	Age          *string `json:"age,omitempty"`
	Adresse      *string `json:"adresse,omitempty"`
	Motif        *string `json:"motif,omitempty"`
	DateMandat   *string `json:"dateMandat,omitempty"`
	Juridiction  *string `json:"juridiction,omitempty"`
	Contacts     *string `json:"contacts,omitempty"`
	Signalement  *string `json:"signalement,omitempty"`
}

// MembreEquipe représente un membre de l'équipe d'intervention
type MembreEquipe struct {
	ID        string  `json:"id" validate:"required"`
	Nom       string  `json:"nom" validate:"required"`
	Matricule string  `json:"matricule" validate:"required"`
	Role      *string `json:"role,omitempty"`
}

// Intervention représente une intervention sur le terrain
type Intervention struct {
	Statut       StatutIntervention `json:"statut" validate:"required"`
	Equipe       []MembreEquipe     `json:"equipe" validate:"required"`
	HeureDepart  *string            `json:"heureDepart,omitempty"`
	HeureArrivee *string            `json:"heureArrivee,omitempty"`
	HeureFin     *string            `json:"heureFin,omitempty"`
	Moyens       []string           `json:"moyens" validate:"required"`
	TempsReponse *string            `json:"tempsReponse,omitempty"`
}

// Evaluation représente l'évaluation sur place
type Evaluation struct {
	SituationReelle string   `json:"situationReelle" validate:"required"`
	Victimes        *int     `json:"victimes,omitempty"`
	Degats          *string  `json:"degats,omitempty"`
	MesuresPrises   []string `json:"mesuresPrises" validate:"required"`
	Renforts        bool     `json:"renforts"`
	RenfortsDetails *string  `json:"renfortsDetails,omitempty"`
}

// Actions représente les actions menées
type Actions struct {
	Immediate  []string `json:"immediate"`
	Preventive []string `json:"preventive"`
	Suivi      []string `json:"suivi"`
}

// Rapport représente le rapport final
type Rapport struct {
	Resume           string   `json:"resume" validate:"required"`
	Conclusions      []string `json:"conclusions" validate:"required"`
	Recommandations  []string `json:"recommandations" validate:"required"`
	SuiteADonner     *string  `json:"suiteADonner,omitempty"`
}

// Temoin représente un témoin
type Temoin struct {
	Nom         *string `json:"nom,omitempty"`
	Telephone   *string `json:"telephone,omitempty"`
	Declaration string  `json:"declaration" validate:"required"`
	Anonyme     *bool   `json:"anonyme,omitempty"`
}

// Document représente un document lié à l'alerte
type Document struct {
	Type        string  `json:"type" validate:"required"`
	Numero      *string `json:"numero,omitempty"`
	Date        *string `json:"date,omitempty"`
	Description string  `json:"description" validate:"required"`
	URL         *string `json:"url,omitempty"`
}

// Suivi représente un suivi de l'alerte
type Suivi struct {
	Date    string  `json:"date" validate:"required"`
	Heure   string  `json:"heure" validate:"required"`
	Agent   string  `json:"agent" validate:"required"`
	AgentID *string `json:"agentId,omitempty"`
	Action  string  `json:"action" validate:"required"`
	Statut  string  `json:"statut" validate:"required"`
}

// DiffusionDestinataires représente les destinataires de la diffusion
type DiffusionDestinataires struct {
	DiffusionGenerale *bool     `json:"diffusionGenerale,omitempty"`
	CommissariatsIds  []string  `json:"commissariatsIds,omitempty"`
	AgentsIds         []string  `json:"agentsIds,omitempty"`
}

// AssignationCommissariat représente l'assignation pour un commissariat
type AssignationCommissariat struct {
	AssigneeGenerale    *bool     `json:"assigneeGenerale,omitempty"`
	AgentsIds           []string  `json:"agentsIds,omitempty"`
	DateAssignation     *string   `json:"dateAssignation,omitempty"`
	AgentAssignateurID  *string   `json:"agentAssignateurId,omitempty"`
}

// CreateAlerteRequest représente la requête de création d'alerte
type CreateAlerteRequest struct {
	Type                 TypeAlerte         `json:"type" validate:"required"`
	Titre                string             `json:"titre" validate:"required"`
	Description          string             `json:"description" validate:"required"`
	Contexte             *string            `json:"contexte,omitempty"`
	Niveau               *NiveauAlerte      `json:"niveau,omitempty"`
	Lieu                 *string            `json:"lieu,omitempty"`
	Latitude             *float64           `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude            *float64           `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	PrecisionLocalisation *string           `json:"precisionLocalisation,omitempty"`
	Risques              []string           `json:"risques,omitempty"`
	PersonneConcernee    *PersonneConcernee `json:"personneConcernee,omitempty"`
	Vehicule             *VehiculeAlerte    `json:"vehicule,omitempty"`
	Suspect              *Suspect           `json:"suspect,omitempty"`
	CommissariatID       string             `json:"commissariatId" validate:"required,uuid"`
	DateAlerte           *time.Time         `json:"dateAlerte,omitempty"`
	Observations         *string            `json:"observations,omitempty"`
}

// UpdateAlerteRequest représente la requête de mise à jour d'alerte
type UpdateAlerteRequest struct {
	Type                 *TypeAlerte         `json:"type,omitempty"`
	Titre                *string             `json:"titre,omitempty"`
	Description          *string             `json:"description,omitempty"`
	Contexte             *string             `json:"contexte,omitempty"`
	Niveau               *NiveauAlerte       `json:"niveau,omitempty"`
	Statut               *StatutAlerte       `json:"statut,omitempty"`
	Lieu                 *string             `json:"lieu,omitempty"`
	Latitude             *float64            `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude            *float64            `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	PrecisionLocalisation *string            `json:"precisionLocalisation,omitempty"`
	Risques              []string            `json:"risques,omitempty"`
	PersonneConcernee    *PersonneConcernee  `json:"personneConcernee,omitempty"`
	Vehicule             *VehiculeAlerte     `json:"vehicule,omitempty"`
	Suspect              *Suspect            `json:"suspect,omitempty"`
	CommissariatID       *string             `json:"commissariatId,omitempty" validate:"omitempty,uuid"`
	DateAlerte           *time.Time          `json:"dateAlerte,omitempty"`
	Observations         *string             `json:"observations,omitempty"`
	Intervention         *Intervention       `json:"intervention,omitempty"`
	Evaluation           *Evaluation         `json:"evaluation,omitempty"`
	Rapport              *Rapport            `json:"rapport,omitempty"`
	Actions              *Actions            `json:"actions,omitempty"`
	Temoins              []Temoin            `json:"temoins,omitempty"`
	Documents            []Document          `json:"documents,omitempty"`
	Photos               []string            `json:"photos,omitempty"`
	Suivis               []Suivi             `json:"suivis,omitempty"`
	DateResolution       *time.Time          `json:"dateResolution,omitempty"`
	DateCloture          *time.Time          `json:"dateCloture,omitempty"`
	Diffusee             *bool               `json:"diffusee,omitempty"`
	DateDiffusion        *time.Time          `json:"dateDiffusion,omitempty"`
}

// FilterAlertesRequest représente les filtres pour la liste des alertes
type FilterAlertesRequest struct {
	Statut         *StatutAlerte  `json:"statut,omitempty"`
	Type           *TypeAlerte    `json:"type,omitempty"`
	Niveau         *NiveauAlerte  `json:"niveau,omitempty"`
	DateDebut      *time.Time     `json:"dateDebut,omitempty"`
	DateFin        *time.Time     `json:"dateFin,omitempty"`
	CommissariatID *string        `json:"commissariatId,omitempty" validate:"omitempty,uuid"`
	Search         *string        `json:"search,omitempty"`
	Diffusee       *bool          `json:"diffusee,omitempty"`
	Page           int            `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit          int            `json:"limit,omitempty" validate:"omitempty,min=1"`
}

// AddSuiviRequest représente l'ajout d'un suivi
type AddSuiviRequest struct {
	Action string `json:"action" validate:"required"`
	Statut string `json:"statut" validate:"required"`
}

// DeployInterventionRequest représente le déploiement d'une intervention
type DeployInterventionRequest struct {
	Equipe []MembreEquipe `json:"equipe" validate:"required"`
	Moyens []string       `json:"moyens" validate:"required"`
}

// UpdateInterventionRequest représente la mise à jour d'une intervention
type UpdateInterventionRequest struct {
	Statut       *StatutIntervention `json:"statut,omitempty"`
	HeureDepart  *string             `json:"heureDepart,omitempty"`
	HeureArrivee *string             `json:"heureArrivee,omitempty"`
	HeureFin     *string             `json:"heureFin,omitempty"`
	Moyens       []string            `json:"moyens,omitempty"`
	TempsReponse *string             `json:"tempsReponse,omitempty"`
}

// AddEvaluationRequest représente l'ajout d'une évaluation
type AddEvaluationRequest struct {
	SituationReelle string   `json:"situationReelle" validate:"required"`
	Victimes        *int     `json:"victimes,omitempty"`
	Degats          *string  `json:"degats,omitempty"`
	MesuresPrises   []string `json:"mesuresPrises" validate:"required"`
	Renforts        bool     `json:"renforts"`
	RenfortsDetails *string  `json:"renfortsDetails,omitempty"`
}

// AddRapportRequest représente l'ajout d'un rapport
type AddRapportRequest struct {
	Resume          string   `json:"resume" validate:"required"`
	Conclusions     []string `json:"conclusions" validate:"required"`
	Recommandations []string `json:"recommandations" validate:"required"`
	SuiteADonner    *string  `json:"suiteADonner,omitempty"`
}

// AddTemoinRequest représente l'ajout d'un témoin
type AddTemoinRequest struct {
	Nom         *string `json:"nom,omitempty"`
	Telephone   *string `json:"telephone,omitempty"`
	Declaration string  `json:"declaration" validate:"required"`
	Anonyme     *bool   `json:"anonyme,omitempty"`
}

// AddDocumentRequest représente l'ajout d'un document
type AddDocumentRequest struct {
	Type        string  `json:"type" validate:"required"`
	Numero      *string `json:"numero,omitempty"`
	Date        *string `json:"date,omitempty"`
	Description string  `json:"description" validate:"required"`
	URL         *string `json:"url,omitempty"`
}

// UpdateActionsRequest représente la mise à jour des actions
type UpdateActionsRequest struct {
	Immediate  []string `json:"immediate,omitempty"`
	Preventive []string `json:"preventive,omitempty"`
	Suivi      []string `json:"suivi,omitempty"`
}

// BroadcastAlerteRequest représente la diffusion d'une alerte
type BroadcastAlerteRequest struct {
	DiffusionGenerale *bool    `json:"diffusionGenerale,omitempty"`
	CommissariatsIds  []string `json:"commissariatsIds,omitempty" validate:"omitempty,dive,uuid"`
	AgentsIds         []string `json:"agentsIds,omitempty" validate:"omitempty,dive,uuid"`
}

// AssignAlerteRequest représente l'assignation d'une alerte
type AssignAlerteRequest struct {
	AssigneeGenerale *bool    `json:"assigneeGenerale,omitempty"`
	AgentsIds        []string `json:"agentsIds,omitempty" validate:"omitempty,dive,uuid"`
}

// GenerateDescriptionRequest représente la génération de description IA
type GenerateDescriptionRequest struct {
	Type                        TypeAlerte             `json:"type" validate:"required"`
	Titre                       string                 `json:"titre" validate:"required"`
	Lieu                        *string                `json:"lieu,omitempty"`
	Risques                     []string               `json:"risques,omitempty"`
	InformationsComplementaires map[string]interface{} `json:"informationsComplementaires,omitempty"`
}

// GenerateDescriptionResponse représente la réponse de génération
type GenerateDescriptionResponse struct {
	Success     bool                        `json:"success"`
	Data        GenerateDescriptionData     `json:"data"`
	Mode        string                      `json:"mode"`
	Message     string                      `json:"message"`
}

// GenerateRapportResponse représente la réponse de génération de rapport
type GenerateRapportResponse struct {
	Success          bool     `json:"success"`
	Resume           string   `json:"resume"`
	Conclusions      []string `json:"conclusions"`
	Recommandations  []string `json:"recommandations"`
	Mode             string   `json:"mode"`
	Message          string   `json:"message"`
}

// GenerateDescriptionData contient les données générées
type GenerateDescriptionData struct {
	Description string `json:"description"`
	Contexte    string `json:"contexte"`
}

// AlerteResponse représente une alerte dans les réponses
type AlerteResponse struct {
	ID                       string                              `json:"id"`
	Numero                   string                              `json:"numero"`
	Type                     TypeAlerte                          `json:"type"`
	Titre                    string                              `json:"titre"`
	Description              string                              `json:"description"`
	Contexte                 *string                             `json:"contexte,omitempty"`
	Niveau                   NiveauAlerte                        `json:"niveau"`
	Statut                   StatutAlerte                        `json:"statut"`
	Lieu                     *string                             `json:"lieu,omitempty"`
	Latitude                 *float64                            `json:"latitude,omitempty"`
	Longitude                *float64                            `json:"longitude,omitempty"`
	PrecisionLocalisation    *string                             `json:"precisionLocalisation,omitempty"`
	Risques                  []string                            `json:"risques"`
	PersonneConcernee        *PersonneConcernee                  `json:"personneConcernee,omitempty"`
	Vehicule                 *VehiculeAlerte                     `json:"vehicule,omitempty"`
	Suspect                  *Suspect                            `json:"suspect,omitempty"`
	AgentRecepteurID         string                              `json:"agentRecepteurId"`
	CommissariatID           string                              `json:"commissariatId"`
	Commissariat             *CommissariatSummary                `json:"commissariat,omitempty"`
	DateAlerte               time.Time                           `json:"dateAlerte"`
	Intervention             *Intervention                       `json:"intervention,omitempty"`
	Evaluation               *Evaluation                         `json:"evaluation,omitempty"`
	Actions                  Actions                             `json:"actions"`
	Rapport                  *Rapport                            `json:"rapport,omitempty"`
	Temoins                  []Temoin                            `json:"temoins"`
	Documents                []Document                          `json:"documents"`
	Photos                   []string                            `json:"photos"`
	Suivis                   []Suivi                             `json:"suivis"`
	Diffusee                 bool                                `json:"diffusee"`
	DateDiffusion            *time.Time                          `json:"dateDiffusion,omitempty"`
	DiffusionDestinataires   *DiffusionDestinataires             `json:"diffusionDestinataires,omitempty"`
	AssignationDestinataires map[string]*AssignationCommissariat `json:"assignationDestinataires,omitempty"`
	DateResolution           *time.Time                          `json:"dateResolution,omitempty"`
	DateCloture              *time.Time                          `json:"dateCloture,omitempty"`
	Observations             *string                             `json:"observations,omitempty"`
	CreatedAt                time.Time                           `json:"createdAt"`
	UpdatedAt                time.Time                           `json:"updatedAt"`
}

// ListAlertesResponse représente la réponse de liste avec pagination
type ListAlertesResponse struct {
	Alertes []AlerteResponse `json:"alertes"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

// StatistiquesAlertesResponse représente les statistiques
type StatistiquesAlertesResponse struct {
	Total              int64                   `json:"total"`
	Actives            int64                   `json:"actives"`
	Resolues           int64                   `json:"resolues"`
	Archivees          int64                   `json:"archivees"`
	ParType            map[TypeAlerte]int64    `json:"parType"`
	ParNiveau          map[NiveauAlerte]int64  `json:"parNiveau"`
	TempsReponseMoyen  *float64                `json:"tempsReponseMoyen,omitempty"`
	TauxResolution     *float64                `json:"tauxResolution,omitempty"`
	EvolutionAlertes   *string                 `json:"evolutionAlertes,omitempty"`     // Ex: "+15% vs hier"
	EvolutionResolution *string                `json:"evolutionResolution,omitempty"` // Ex: "+10% vs semaine dernière"
}

// DashboardStatsValue représente une valeur de statistique avec son évolution
type DashboardStatsValue struct {
	Total     int    `json:"total"`
	Evolution string `json:"evolution"`
}

// DashboardTempsReponse représente le temps de réponse moyen avec son évolution
type DashboardTempsReponse struct {
	Moyen     string `json:"moyen"`
	Evolution string `json:"evolution"`
}

// DashboardStats représente les statistiques principales du dashboard
type DashboardStats struct {
	TotalAlertes DashboardStatsValue   `json:"totalAlertes"`
	Resolues     DashboardStatsValue   `json:"resolues"`
	EnCours      DashboardStatsValue   `json:"enCours"`
	TempsReponse DashboardTempsReponse `json:"tempsReponse"`
}

// DashboardStatsTableItem représente une ligne du tableau de statistiques par type
type DashboardStatsTableItem struct {
	Type     string `json:"type"`
	Nombre   int    `json:"nombre"`
	Resolues int    `json:"resolues"`
	Taux     int    `json:"taux"`
}

// DashboardActivityData représente les données d'activité par période
type DashboardActivityData struct {
	Period   string `json:"period"`
	Alertes  int    `json:"alertes"`
	EnCours  int    `json:"enCours"`
	Resolues int    `json:"resolues"`
}

// DashboardAlertItem représente une alerte dans le tableau du dashboard
type DashboardAlertItem struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	TypeAlerte      string `json:"typeAlerte"`
	Libelle         string `json:"libelle"`
	TypeDiffusion   string `json:"typeDiffusion"`
	DateDiffusion   string `json:"dateDiffusion"`
	Status          string `json:"status"`
	VilleDiffusion  string `json:"villeDiffusion"`
	Priorite        string `json:"priorite"`
}

// DashboardResponse représente la réponse complète du dashboard
type DashboardResponse struct {
	Stats        DashboardStats            `json:"stats"`
	StatsTable   []DashboardStatsTableItem `json:"statsTable"`
	ActivityData []DashboardActivityData   `json:"activityData"`
	Alerts       []DashboardAlertItem      `json:"alerts"`
}

// CommissariatSummary représente un résumé de commissariat
type CommissariatSummary struct {
	ID    string `json:"id"`
	Nom   string `json:"nom"`
	Code  string `json:"code"`
	Ville string `json:"ville"`
}
