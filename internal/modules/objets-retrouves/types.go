package objetsretrouves

import (
	"time"
)

// StatutObjetRetrouve représente le statut de l'objet retrouvé
type StatutObjetRetrouve string

const (
	StatutObjetRetrouveDisponible StatutObjetRetrouve = "DISPONIBLE"
	StatutObjetRetrouveRestitue   StatutObjetRetrouve = "RESTITUÉ"
	StatutObjetRetrouveNonReclame StatutObjetRetrouve = "NON_RÉCLAMÉ"
)

// InventoryItem représente un objet dans l'inventaire d'un contenant
type InventoryItem struct {
	ID             int     `json:"id,omitempty"`
	Category       string  `json:"category" validate:"required"`
	Icon           string  `json:"icon,omitempty"`
	Name           string  `json:"name" validate:"required"`
	Color          string  `json:"color" validate:"required"`
	Brand          *string `json:"brand,omitempty"`
	Serial         *string `json:"serial,omitempty"`
	Description    *string `json:"description,omitempty"`
	IdentityType   *string `json:"identityType,omitempty"`
	IdentityNumber *string `json:"identityNumber,omitempty"`
	IdentityName   *string `json:"identityName,omitempty"`
	CardType       *string `json:"cardType,omitempty"`
	CardBank       *string `json:"cardBank,omitempty"`
	CardLast4      *string `json:"cardLast4,omitempty"`
}

// ContainerDetails représente les détails d'un contenant
type ContainerDetails struct {
	Type              string          `json:"type" validate:"required"`
	Couleur           *string         `json:"couleur,omitempty"`
	Marque            *string         `json:"marque,omitempty"`
	Taille            *string         `json:"taille,omitempty"`
	SignesDistinctifs *string         `json:"signesDistinctifs,omitempty"`
	Inventory         []InventoryItem `json:"inventory,omitempty"`
}

// CreateObjetRetrouveRequest représente la requête de création d'un objet retrouvé
type CreateObjetRetrouveRequest struct {
	TypeObjet          string                 `json:"typeObjet" validate:"required"`
	Description        string                 `json:"description" validate:"required"`
	ValeurEstimee      *string                `json:"valeurEstimee,omitempty"`
	Couleur            *string                `json:"couleur,omitempty"`
	DetailsSpecifiques map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer      *bool             `json:"isContainer,omitempty"`
	ContainerDetails *ContainerDetails `json:"containerDetails,omitempty"`
	
	Deposant         DeposantRequest   `json:"deposant" validate:"required"`
	LieuTrouvaille   string            `json:"lieuTrouvaille" validate:"required"`
	AdresseLieu      *string           `json:"adresseLieu,omitempty"`
	DateTrouvaille   string            `json:"dateTrouvaille" validate:"required"`
	HeureTrouvaille  *string           `json:"heureTrouvaille,omitempty"`
	Observations     *string           `json:"observations,omitempty"`
}

// DeposantRequest représente les informations du déposant
type DeposantRequest struct {
	Nom       string  `json:"nom" validate:"required"`
	Prenom    string  `json:"prenom" validate:"required"`
	Telephone string  `json:"telephone" validate:"required"`
	Email     *string `json:"email,omitempty"`
	Adresse   *string `json:"adresse,omitempty"`
	CNI       *string `json:"cni,omitempty"`
}

// UpdateObjetRetrouveRequest représente la requête de mise à jour d'un objet retrouvé
type UpdateObjetRetrouveRequest struct {
	TypeObjet          *string                `json:"typeObjet,omitempty"`
	Description        *string                `json:"description,omitempty"`
	ValeurEstimee      *string                `json:"valeurEstimee,omitempty"`
	Couleur            *string                `json:"couleur,omitempty"`
	DetailsSpecifiques map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer      *bool             `json:"isContainer,omitempty"`
	ContainerDetails *ContainerDetails `json:"containerDetails,omitempty"`
	
	Deposant         *DeposantRequest  `json:"deposant,omitempty"`
	LieuTrouvaille   *string           `json:"lieuTrouvaille,omitempty"`
	AdresseLieu      *string           `json:"adresseLieu,omitempty"`
	DateTrouvaille   *string           `json:"dateTrouvaille,omitempty"`
	HeureTrouvaille  *string           `json:"heureTrouvaille,omitempty"`
	Observations     *string           `json:"observations,omitempty"`
}

// FilterObjetsRetrouvesRequest représente les filtres pour la liste des objets retrouvés
type FilterObjetsRetrouvesRequest struct {
	Statut         *string    `json:"statut,omitempty"`
	TypeObjet      *string    `json:"typeObjet,omitempty"`
	CommissariatID *string    `json:"commissariatId,omitempty"`
	IsContainer    *bool      `json:"isContainer,omitempty"`
	DateDebut      *time.Time `json:"dateDebut,omitempty"`
	DateFin        *time.Time `json:"dateFin,omitempty"`
	Search         *string    `json:"search,omitempty"`
	Page           int        `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit          int        `json:"limit,omitempty" validate:"omitempty,min=1"`
}

// UpdateStatutRequest représente la requête de mise à jour du statut
type UpdateStatutRequest struct {
	Statut          string                 `json:"statut" validate:"required"`
	DateRestitution *time.Time             `json:"dateRestitution,omitempty"`
	Proprietaire    *ProprietaireRequest   `json:"proprietaire,omitempty"`
}

// ProprietaireRequest représente les informations du propriétaire
type ProprietaireRequest struct {
	Nom       string  `json:"nom" validate:"required"`
	Prenom    string  `json:"prenom" validate:"required"`
	Telephone string  `json:"telephone" validate:"required"`
	Email     *string `json:"email,omitempty"`
	Adresse   *string `json:"adresse,omitempty"`
	CNI       *string `json:"cni,omitempty"`
}

// ObjetRetrouveResponse représente un objet retrouvé dans les réponses
type ObjetRetrouveResponse struct {
	ID                      string                 `json:"id"`
	Numero                  string                 `json:"numero"`
	TypeObjet               string                 `json:"typeObjet"`
	Description             string                 `json:"description"`
	ValeurEstimee           *string                `json:"valeurEstimee,omitempty"`
	Couleur                 *string                `json:"couleur,omitempty"`
	DetailsSpecifiques      map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer             bool                   `json:"isContainer"`
	ContainerDetails        *ContainerDetails      `json:"containerDetails,omitempty"`
	
	Deposant                map[string]interface{} `json:"deposant"`
	LieuTrouvaille          string                 `json:"lieuTrouvaille"`
	AdresseLieu             *string                `json:"adresseLieu,omitempty"`
	DateTrouvaille          time.Time              `json:"dateTrouvaille"`
	DateTrouvailleFormatee  string                 `json:"dateTrouvailleFormatee,omitempty"`
	HeureTrouvaille         *string                `json:"heureTrouvaille,omitempty"`
	Statut                  StatutObjetRetrouve    `json:"statut"`
	DateDepot               time.Time              `json:"dateDepot"`
	DateDepotFormatee       string                 `json:"dateDepotFormatee,omitempty"`
	DateRestitution         *time.Time             `json:"dateRestitution,omitempty"`
	DateRestitutionFormatee *string                `json:"dateRestitutionFormatee,omitempty"`
	Proprietaire            map[string]interface{} `json:"proprietaire,omitempty"`
	Observations            *string                `json:"observations,omitempty"`
	Agent                   *AgentSummary          `json:"agent,omitempty"`
	Commissariat            *CommissariatSummary   `json:"commissariat,omitempty"`
	Historique              []HistoriqueEntry      `json:"historique,omitempty"`
	CreatedAt               time.Time              `json:"createdAt"`
	UpdatedAt               time.Time              `json:"updatedAt"`
}

// AgentSummary représente un résumé d'agent
type AgentSummary struct {
	ID        string `json:"id"`
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Matricule string `json:"matricule"`
}

// CommissariatSummary représente un résumé de commissariat
type CommissariatSummary struct {
	ID    string `json:"id"`
	Nom   string `json:"nom"`
	Code  string `json:"code"`
	Ville string `json:"ville"`
}

// HistoriqueEntry représente une entrée d'historique
type HistoriqueEntry struct {
	Date    string `json:"date"`
	DateISO string `json:"dateISO"`
	Action  string `json:"action"`
	Agent   string `json:"agent"`
	Details *string `json:"details,omitempty"`
}

// ListObjetsRetrouvesResponse représente la réponse de liste avec pagination
type ListObjetsRetrouvesResponse struct {
	Objets []ObjetRetrouveResponse `json:"objets"`
	Total  int64                   `json:"total"`
	Page   int                     `json:"page"`
	Limit  int                     `json:"limit"`
}

// StatistiquesObjetsRetrouvesResponse représente les statistiques des objets retrouvés
type StatistiquesObjetsRetrouvesResponse struct {
	Total                int64   `json:"total"`
	Disponibles          int64   `json:"disponibles"`
	Restitues            int64   `json:"restitues"`
	NonReclames          int64   `json:"nonReclames"`
	TauxRestitution      float64 `json:"tauxRestitution"`
	EvolutionTotal       string  `json:"evolutionTotal"`
	EvolutionDisponibles string  `json:"evolutionDisponibles"`
	EvolutionRestitues   string  `json:"evolutionRestitues"`
	EvolutionNonReclames string  `json:"evolutionNonReclames"`
	EvolutionTauxRestitution string `json:"evolutionTauxRestitution"`
}

// DashboardResponse représente la réponse complète du dashboard
type DashboardResponse struct {
	Stats      DashboardStatsValue  `json:"stats"`
	TopTypes   []TopTypes	 `json:"topTypes"`
	ActivityData []DashboardActivityData   `json:"activityData"`
	
}

// DashboardStatsValue représente une valeur de statistique avec son évolution

type DashboardStatsValue struct {
	Total                 int64   `json:"total"`
	Disponibles           int64   `json:"disponibles"`
	Restitues             int64   `json:"restitues"`
	NonReclames              int64   `json:"nonReclames"`
	TauxRestitution          float64 `json:"tauxRestitution"`
	EvolutionTotal        string  `json:"evolutionTotal"`
	EvolutionDisponibles  string  `json:"evolutionDisponibles"`
	EvolutionRestitues    string  `json:"evolutionRestitues"`
	EvolutionNonReclames     string  `json:"evolutionNonReclames"`
	
}

// DashboardActivityData représente les données d'activité par période
type DashboardActivityData struct {
	Period   string `json:"period"`
	ObjetsRetrouves  int    `json:"objetsPerdus"`
	Disponibles  int    `json:"disponibles"`
	Restitues int    `json:"restitues"`
	NonReclames int    `json:"nonReclames"`
}

// TopTypes représente les types d'objets perdus les plus communs
type TopTypes struct {
	Type     string `json:"type"`
	Count    int    `json:"count"`
}


