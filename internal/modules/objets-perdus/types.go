package objetsperdus

import (
	"time"
)

// StatutObjetPerdu représente le statut de l'objet perdu
type StatutObjetPerdu string

const (
	StatutObjetPerduEnRecherche StatutObjetPerdu = "EN_RECHERCHE"
	StatutObjetPerduRetrouve    StatutObjetPerdu = "RETROUVÉ"
	StatutObjetPerduCloture     StatutObjetPerdu = "CLÔTURÉ"
)

// InventoryItem représente un objet dans l'inventaire d'un contenant
type InventoryItem struct {
	ID                int     `json:"id,omitempty"`
	Category          string  `json:"category" validate:"required"`
	Icon              string  `json:"icon,omitempty"`
	Name              string  `json:"name" validate:"required"`
	Color             string  `json:"color" validate:"required"`
	Brand             *string `json:"brand,omitempty"`
	Serial            *string `json:"serial,omitempty"`
	Description       *string `json:"description,omitempty"`
	IdentityType      *string `json:"identityType,omitempty"`
	IdentityNumber    *string `json:"identityNumber,omitempty"`
	IdentityName      *string `json:"identityName,omitempty"`
	CardType          *string `json:"cardType,omitempty"`
	CardBank          *string `json:"cardBank,omitempty"`
	CardLast4         *string `json:"cardLast4,omitempty"`
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

// CreateObjetPerduRequest représente la requête de création d'un objet perdu
type CreateObjetPerduRequest struct {
	TypeObjet          string                 `json:"typeObjet" validate:"required"`
	Description        string                 `json:"description" validate:"required"`
	ValeurEstimee      *string                `json:"valeurEstimee,omitempty"`
	Couleur            *string                `json:"couleur,omitempty"`
	DetailsSpecifiques map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer        *bool             `json:"isContainer,omitempty"`
	ContainerDetails   *ContainerDetails `json:"containerDetails,omitempty"`
	
	Declarant          DeclarantRequest  `json:"declarant" validate:"required"`
	LieuPerte          string            `json:"lieuPerte" validate:"required"`
	AdresseLieu        *string           `json:"adresseLieu,omitempty"`
	DatePerte          string            `json:"datePerte" validate:"required"`
	HeurePerte         *string           `json:"heurePerte,omitempty"`
	Observations       *string           `json:"observations,omitempty"`
}

// DeclarantRequest représente les informations du déclarant
type DeclarantRequest struct {
	Nom       string  `json:"nom" validate:"required"`
	Prenom    string  `json:"prenom" validate:"required"`
	Telephone string  `json:"telephone" validate:"required"`
	Email     *string `json:"email,omitempty"`
	Adresse   *string `json:"adresse,omitempty"`
	CNI       *string `json:"cni,omitempty"`
}

// UpdateObjetPerduRequest représente la requête de mise à jour d'un objet perdu
type UpdateObjetPerduRequest struct {
	TypeObjet          *string                `json:"typeObjet,omitempty"`
	Description        *string                `json:"description,omitempty"`
	ValeurEstimee      *string                `json:"valeurEstimee,omitempty"`
	Couleur            *string                `json:"couleur,omitempty"`
	DetailsSpecifiques map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer        *bool             `json:"isContainer,omitempty"`
	ContainerDetails   *ContainerDetails `json:"containerDetails,omitempty"`
	
	Declarant          *DeclarantRequest `json:"declarant,omitempty"`
	LieuPerte          *string           `json:"lieuPerte,omitempty"`
	AdresseLieu        *string           `json:"adresseLieu,omitempty"`
	DatePerte          *string           `json:"datePerte,omitempty"`
	HeurePerte         *string           `json:"heurePerte,omitempty"`
	Observations       *string           `json:"observations,omitempty"`
}

// FilterObjetsPerdusRequest représente les filtres pour la liste des objets perdus
type FilterObjetsPerdusRequest struct {
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
	Statut      string     `json:"statut" validate:"required"`
	DateRetrouve *time.Time `json:"dateRetrouve,omitempty"`
}

// ObjetPerduResponse représente un objet perdu dans les réponses
type ObjetPerduResponse struct {
	ID                      string                 `json:"id"`
	Numero                  string                 `json:"numero"`
	TypeObjet               string                 `json:"typeObjet"`
	Description             string                 `json:"description"`
	ValeurEstimee           *string                `json:"valeurEstimee,omitempty"`
	Couleur                 *string                `json:"couleur,omitempty"`
	DetailsSpecifiques      map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	
	// Nouveaux champs pour le mode contenant
	IsContainer             bool                   `json:"isContainer"`
	ContainerDetails        *ContainerDetails      `json:"containerDetails"`
	
	Declarant               map[string]interface{} `json:"declarant"`
	LieuPerte               string                 `json:"lieuPerte"`
	AdresseLieu             *string                `json:"adresseLieu,omitempty"`
	DatePerte               time.Time              `json:"datePerte"`
	DatePerteFormatee       string                 `json:"datePerteFormatee,omitempty"`
	HeurePerte              *string                `json:"heurePerte,omitempty"`
	Statut                  StatutObjetPerdu       `json:"statut"`
	DateDeclaration         time.Time              `json:"dateDeclaration"`
	DateDeclarationFormatee string                 `json:"dateDeclarationFormatee,omitempty"`
	DateRetrouve            *time.Time             `json:"dateRetrouve,omitempty"`
	DateRetrouveFormatee    *string                `json:"dateRetrouveFormatee,omitempty"`
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

// ListObjetsPerdusResponse représente la réponse de liste avec pagination
type ListObjetsPerdusResponse struct {
	Objets []ObjetPerduResponse `json:"objets"`
	Total  int64                `json:"total"`
	Page   int                  `json:"page"`
	Limit  int                  `json:"limit"`
}

// StatistiquesObjetsPerdusResponse représente les statistiques des objets perdus
type StatistiquesObjetsPerdusResponse struct {
	Total                 int64   `json:"total"`
	EnRecherche           int64   `json:"enRecherche"`
	Retrouves             int64   `json:"retrouves"`
	Clotures              int64   `json:"clotures"`
	TauxRetrouve          float64 `json:"tauxRetrouve"`
	EvolutionTotal        string  `json:"evolutionTotal"`
	EvolutionEnRecherche  string  `json:"evolutionEnRecherche"`
	EvolutionRetrouves    string  `json:"evolutionRetrouves"`
	EvolutionClotures     string  `json:"evolutionClotures"`
	EvolutionTauxRetrouve string  `json:"evolutionTauxRetrouve"`
}

// DashboardResponse représente la réponse complète du dashboard
type DashboardResponse struct {
	Stats      DashboardStatsValue  `json:"stats"`
	TopTypes   []TopTypes	 `json:"topTypes"`
//	StatusDistribution []StatusDistribution `json:"statusDistribution"`
	ActivityData []DashboardActivityData   `json:"activityData"`
	
}

// DashboardStatsValue représente une valeur de statistique avec son évolution
type DashboardStatsValue struct {
	Total                 int64   `json:"total"`
	EnRecherche           int64   `json:"enRecherche"`
	Retrouves             int64   `json:"retrouves"`
	Clotures              int64   `json:"clotures"`
	TauxRetrouve          float64 `json:"tauxRetrouve"`
	EvolutionTotal        string  `json:"evolutionTotal"`
	EvolutionEnRecherche  string  `json:"evolutionEnRecherche"`
	EvolutionRetrouves    string  `json:"evolutionRetrouves"`
	EvolutionClotures     string  `json:"evolutionClotures"`
	EvolutionTauxRetrouve string  `json:"evolutionTauxRetrouve"`
}

// DashboardTempsReponse représente le temps de réponse moyen avec son évolution
type DashboardTempsReponse struct {
	Moyen     string `json:"moyen"`
	Evolution string `json:"evolution"`
}



// DashboardActivityData représente les données d'activité par période
type DashboardActivityData struct {
	Period   string `json:"period"`
	ObjetsPerdus  int    `json:"objetsPerdus"`
	Recherche  int    `json:"recherche"`
	Retrouves int    `json:"retrouves"`
	Clotures int    `json:"clotures"`
}


// TopTypes représente les types d'objets perdus les plus communs
type TopTypes struct {
	Type     string `json:"type"`
	Count    int    `json:"count"`
}


// StatusDistribution représente la distribution des statuts des objets perdus
type StatusDistribution struct {
	Name     string `json:"name"`
	Value    int    `json:"value"`
	Color    string `json:"color"`
}

// CheckMatchesRequest représente la requête de vérification de correspondances
type CheckMatchesRequest struct {
	TypeObjet   string                 `json:"typeObjet" validate:"required"`
	Identifiers map[string]interface{} `json:"identifiers" validate:"required"`
}

// CheckMatchesResponse représente la réponse avec les objets retrouvés correspondants
type CheckMatchesResponse struct {
	Matches []MatchedObjetRetrouve `json:"matches"`
	Count   int                    `json:"count"`
}

// MatchedObjetRetrouve représente un objet retrouvé correspondant avec son score
type MatchedObjetRetrouve struct {
	ID                      string                 `json:"id"`
	Numero                  string                 `json:"numero"`
	TypeObjet               string                 `json:"typeObjet"`
	Description             string                 `json:"description"`
	ValeurEstimee           *string                `json:"valeurEstimee,omitempty"`
	Couleur                 *string                `json:"couleur,omitempty"`
	DetailsSpecifiques      map[string]interface{} `json:"detailsSpecifiques,omitempty"`
	IsContainer             bool                   `json:"isContainer"`
	ContainerDetails        map[string]interface{} `json:"containerDetails,omitempty"`
	LieuTrouvaille          string                 `json:"lieuTrouvaille"`
	DateTrouvaille          string                 `json:"dateTrouvaille"`
	DateTrouvailleFormatee  string                 `json:"dateTrouvailleFormatee,omitempty"`
	Statut                  string                 `json:"statut"`
	Deposant                map[string]interface{} `json:"deposant"`
	Commissariat            *CommissariatSummary   `json:"commissariat,omitempty"`
	MatchScore              int                    `json:"matchScore"`
	MatchedField            string                 `json:"matchedField"`
	MatchedIn               string                 `json:"matchedIn"` // "direct" ou "inventory"
	InventoryItem           map[string]interface{} `json:"inventoryItem,omitempty"`
}
