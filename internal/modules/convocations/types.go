package convocations

import (
	"time"
)

// StatutConvocation représente le statut de la convocation
type StatutConvocation string

const (
	StatutConvocationCreation   StatutConvocation = "CRÉATION"
	StatutConvocationEnvoyee    StatutConvocation = "ENVOYÉ"
	StatutConvocationHonoree    StatutConvocation = "HONORÉ"
	StatutConvocationEnAttente  StatutConvocation = "EN ATTENTE"
	StatutConvocationConfirme   StatutConvocation = "CONFIRMÉ"
	StatutConvocationNonHonoree StatutConvocation = "NON HONORÉ"
	StatutConvocationAnnule     StatutConvocation = "ANNULÉ"
)

// Urgence représente le niveau d'urgence de la convocation
type Urgence string

const (
	UrgenceNormale    Urgence = "NORMALE"
	UrgenceUrgent     Urgence = "URGENT"
	UrgenceTresUrgent Urgence = "TRES_URGENT"
)

// Priorite représente la priorité de la convocation
type Priorite string

const (
	PrioriteBasse    Priorite = "BASSE"
	PrioriteMoyenne  Priorite = "MOYENNE"
	PrioriteHaute    Priorite = "HAUTE"
	PrioriteCritique Priorite = "CRITIQUE"
)

// Confidentialite représente le niveau de confidentialité
type Confidentialite string

const (
	ConfidentialiteStandard         Confidentialite = "STANDARD"
	ConfidentialiteConfidentiel     Confidentialite = "CONFIDENTIEL"
	ConfidentialiteTresConfidentiel Confidentialite = "TRES_CONFIDENTIEL"
	ConfidentialiteSecretDefense    Confidentialite = "SECRET_DEFENSE"
)

// CreateConvocationRequest représente la requête de création d'une convocation
type CreateConvocationRequest struct {
	// SECTION 1: INFORMATIONS GÉNÉRALES
	Reference       string          `json:"reference"`
	TypeConvocation string          `json:"typeConvocation" validate:"required"`
	SousType        *string         `json:"sousType,omitempty"`
	Urgence         Urgence         `json:"urgence" validate:"required"`
	Priorite        Priorite        `json:"priorite" validate:"required"`
	Confidentialite Confidentialite `json:"confidentialite" validate:"required"`

	// SECTION 2: AFFAIRE LIÉE
	AffaireID           *string `json:"affaireId,omitempty"`
	AffaireType         *string `json:"affaireType,omitempty"`
	AffaireNumero       *string `json:"affaireNumero,omitempty"`
	AffaireTitre        *string `json:"affaireTitre,omitempty"`
	SectionJudiciaire   *string `json:"sectionJudiciaire,omitempty"`
	Infraction          *string `json:"infraction,omitempty"`
	QualificationLegale *string `json:"qualificationLegale,omitempty"`

	// SECTION 3: PERSONNE CONVOQUÉE
	StatutPersonne     string  `json:"statutPersonne" validate:"required"`
	Nom                string  `json:"nom" validate:"required"`
	Prenom             string  `json:"prenom" validate:"required"`
	DateNaissance      *string `json:"dateNaissance,omitempty"`
	LieuNaissance      *string `json:"lieuNaissance,omitempty"`
	Nationalite        *string `json:"nationalite,omitempty"`
	Profession         *string `json:"profession,omitempty"`
	SituationFamiliale *string `json:"situationFamiliale,omitempty"`
	NombreEnfants      *string `json:"nombreEnfants,omitempty"`

	// Documents d'identité
	TypePiece           string  `json:"typePiece" validate:"required"`
	NumeroPiece         string  `json:"numeroPiece" validate:"required"`
	DateDelivrancePiece *string `json:"dateDelivrancePiece,omitempty"`
	LieuDelivrancePiece *string `json:"lieuDelivrancePiece,omitempty"`
	DateExpirationPiece *string `json:"dateExpirationPiece,omitempty"`

	// Contact
	Telephone1             string  `json:"telephone1" validate:"required"`
	Telephone2             *string `json:"telephone2,omitempty"`
	Email                  *string `json:"email,omitempty"`
	AdresseResidence       *string `json:"adresseResidence,omitempty"`
	AdresseProfessionnelle *string `json:"adresseProfessionnelle,omitempty"`
	DernierLieuConnu       *string `json:"dernierLieuConnu,omitempty"`

	// Caractéristiques physiques
	Sexe               *string `json:"sexe,omitempty"`
	Taille             *string `json:"taille,omitempty"`
	Poids              *string `json:"poids,omitempty"`
	SignesParticuliers *string `json:"signesParticuliers,omitempty"`
	PhotoIdentite      bool    `json:"photoIdentite"`
	Empreintes         bool    `json:"empreintes"`

	// SECTION 4: RENDEZ-VOUS
	DateCreation     string  `json:"dateCreation" validate:"required"`
	HeureConvocation *string `json:"heureConvocation,omitempty"`
	DateRdv          *string `json:"dateRdv,omitempty"`
	HeureRdv         *string `json:"heureRdv,omitempty"`
	DureeEstimee     *int    `json:"dureeEstimee,omitempty"`
	TypeAudience     string  `json:"typeAudience" validate:"required"`

	// Lieu
	LieuRdv         string  `json:"lieuRdv" validate:"required"`
	Bureau          *string `json:"bureau,omitempty"`
	SalleAudience   *string `json:"salleAudience,omitempty"`
	PointRencontre  *string `json:"pointRencontre,omitempty"`
	AccesSpecifique *string `json:"accesSpecifique,omitempty"`

	// SECTION 5: PERSONNES PRÉSENTES
	ConvocateurNom       string  `json:"convocateurNom" validate:"required"`
	ConvocateurPrenom    string  `json:"convocateurPrenom" validate:"required"`
	ConvocateurMatricule *string `json:"convocateurMatricule,omitempty"`
	ConvocateurFonction  *string `json:"convocateurFonction,omitempty"`

	AgentsPresents       *string `json:"agentsPresents,omitempty"`
	RepresentantParquet  bool    `json:"representantParquet"`
	NomParquetier        *string `json:"nomParquetier,omitempty"`
	ExpertPresent        bool    `json:"expertPresent"`
	TypeExpert           *string `json:"typeExpert,omitempty"`
	InterpreteNecessaire bool    `json:"interpreteNecessaire"`
	LangueInterpretation *string `json:"langueInterpretation,omitempty"`
	AvocatPresent        bool    `json:"avocatPresent"`
	NomAvocat            *string `json:"nomAvocat,omitempty"`
	BarreauAvocat        *string `json:"barreauAvocat,omitempty"`

	// SECTION 6: MOTIF ET OBJET
	Motif                  string  `json:"motif" validate:"required"`
	ObjetPrecis            *string `json:"objetPrecis,omitempty"`
	QuestionsPreparatoires *string `json:"questionsPreparatoires,omitempty"`
	PiecesAApporter        *string `json:"piecesAApporter,omitempty"`
	DocumentsDemandes      *string `json:"documentsDemandes,omitempty"`

	// SECTION 9: OBSERVATIONS
	Observations *string `json:"observations,omitempty"`

	// SECTION 10: ÉTAT ET TRAÇABILITÉ
	Statut    *StatutConvocation `json:"statut,omitempty"`
	ModeEnvoi string             `json:"modeEnvoi" validate:"required"`

	// Contexte
	AgentID        *string `json:"agentId,omitempty"`
	CommissariatID *string `json:"commissariatId,omitempty"`
	CreatedBy      *string `json:"createdBy,omitempty"`
	UpdatedBy      *string `json:"updatedBy,omitempty"`
}

// ConvocationResponse représente une convocation COMPLÈTE avec TOUS les 74 champs
type ConvocationResponse struct {
	// Identifiants
	ID     string `json:"id"`
	Numero string `json:"numero"`

	// SECTION 1: INFORMATIONS GÉNÉRALES
	Reference       *string          `json:"reference,omitempty"`
	TypeConvocation string           `json:"typeConvocation"`
	SousType        *string          `json:"sousType,omitempty"`
	Urgence         *string          `json:"urgence,omitempty"`
	Priorite        *string          `json:"priorite,omitempty"`
	Confidentialite *string          `json:"confidentialite,omitempty"`

	// SECTION 2: AFFAIRE LIÉE
	AffaireID           *string `json:"affaireId,omitempty"`
	AffaireType         *string `json:"affaireType,omitempty"`
	AffaireNumero       *string `json:"affaireNumero,omitempty"`
	AffaireTitre        *string `json:"affaireTitre,omitempty"`
	SectionJudiciaire   *string `json:"sectionJudiciaire,omitempty"`
	Infraction          *string `json:"infraction,omitempty"`
	QualificationLegale *string `json:"qualificationLegale,omitempty"`

	// SECTION 3: PERSONNE CONVOQUÉE - Identité
	StatutPersonne     string  `json:"statutPersonne"`
	ConvoqueNom        string  `json:"convoqueNom"`
	ConvoquePrenom     string  `json:"convoquePrenom"`
	DateNaissance      *string `json:"dateNaissance,omitempty"`
	LieuNaissance      *string `json:"lieuNaissance,omitempty"`
	Nationalite        *string `json:"nationalite,omitempty"`
	Profession         *string `json:"profession,omitempty"`
	SituationFamiliale *string `json:"situationFamiliale,omitempty"`
	NombreEnfants      *string `json:"nombreEnfants,omitempty"`

	// SECTION 3: Pièce d'identité
	TypePiece           string  `json:"typePiece"`
	NumeroPiece         string  `json:"numeroPiece"`
	DateDelivrancePiece *string `json:"dateDelivrancePiece,omitempty"`
	LieuDelivrancePiece *string `json:"lieuDelivrancePiece,omitempty"`
	DateExpirationPiece *string `json:"dateExpirationPiece,omitempty"`

	// SECTION 3: Contact
	ConvoqueTelephone      string  `json:"convoqueTelephone"`
	ConvoqueTelephone2     *string `json:"convoqueTelephone2,omitempty"`
	ConvoqueEmail          *string `json:"convoqueEmail,omitempty"`
	AdresseResidence       *string `json:"adresseResidence,omitempty"`
	AdresseProfessionnelle *string `json:"adresseProfessionnelle,omitempty"`
	DernierLieuConnu       *string `json:"dernierLieuConnu,omitempty"`

	// SECTION 3: Caractéristiques physiques
	Sexe               *string `json:"sexe,omitempty"`
	Taille             *string `json:"taille,omitempty"`
	Poids              *string `json:"poids,omitempty"`
	SignesParticuliers *string `json:"signesParticuliers,omitempty"`
	PhotoIdentite      bool    `json:"photoIdentite"`
	Empreintes         bool    `json:"empreintes"`

	// SECTION 4: RENDEZ-VOUS
	DateCreation     time.Time  `json:"dateCreation"`
	HeureConvocation *string    `json:"heureConvocation,omitempty"`
	DateRdv          *time.Time `json:"dateRdv,omitempty"`
	HeureRdv         *string    `json:"heureRdv,omitempty"`
	DureeEstimee     *int       `json:"dureeEstimee,omitempty"`
	TypeAudience     *string    `json:"typeAudience,omitempty"`

	// SECTION 4: Lieu
	LieuRdv         string  `json:"lieuRdv"`
	Bureau          *string `json:"bureau,omitempty"`
	SalleAudience   *string `json:"salleAudience,omitempty"`
	PointRencontre  *string `json:"pointRencontre,omitempty"`
	AccesSpecifique *string `json:"accesSpecifique,omitempty"`

	// SECTION 5: PERSONNES PRÉSENTES
	ConvocateurNom       string  `json:"convocateurNom"`
	ConvocateurPrenom    string  `json:"convocateurPrenom"`
	ConvocateurMatricule *string `json:"convocateurMatricule,omitempty"`
	ConvocateurFonction  *string `json:"convocateurFonction,omitempty"`

	AgentsPresents       *string `json:"agentsPresents,omitempty"`
	RepresentantParquet  bool    `json:"representantParquet"`
	NomParquetier        *string `json:"nomParquetier,omitempty"`
	ExpertPresent        bool    `json:"expertPresent"`
	TypeExpert           *string `json:"typeExpert,omitempty"`
	InterpreteNecessaire bool    `json:"interpreteNecessaire"`
	LangueInterpretation *string `json:"langueInterpretation,omitempty"`
	AvocatPresent        bool    `json:"avocatPresent"`
	NomAvocat            *string `json:"nomAvocat,omitempty"`
	BarreauAvocat        *string `json:"barreauAvocat,omitempty"`

	// SECTION 6: MOTIF ET OBJET
	Motif                  string  `json:"motif"`
	ObjetPrecis            *string `json:"objetPrecis,omitempty"`
	QuestionsPreparatoires *string `json:"questionsPreparatoires,omitempty"`
	PiecesAApporter        *string `json:"piecesAApporter,omitempty"`
	DocumentsDemandes      *string `json:"documentsDemandes,omitempty"`

	// SECTION 9: OBSERVATIONS
	Observations *string `json:"observations,omitempty"`

	// SECTION 10: ÉTAT ET TRAÇABILITÉ
	DateEnvoi        *time.Time        `json:"dateEnvoi,omitempty"`
	DateHonoration   *time.Time        `json:"dateHonoration,omitempty"`
	Statut           StatutConvocation `json:"statut"`
	ResultatAudition *string           `json:"resultatAudition,omitempty"`
	ModeEnvoi        string            `json:"modeEnvoi"`

	// Relations
	Agent        *AgentSummary        `json:"agent,omitempty"`
	Commissariat *CommissariatSummary `json:"commissariat,omitempty"`
	Historique   []HistoriqueEntry    `json:"historique,omitempty"`

	// Aliases pour compatibilité
	QualiteConvoque string  `json:"qualiteConvoque"`
	ConvoqueAdresse *string `json:"convoqueAdresse,omitempty"`
	AffaireLiee     *string `json:"affaireLiee,omitempty"`

	// Métadonnées
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
	ID   string `json:"id"`
	Nom  string `json:"nom"`
	Code string `json:"code"`
}

// HistoriqueEntry représente une entrée d'historique
type HistoriqueEntry struct {
	Date    string  `json:"date"`
	DateISO string  `json:"dateISO"`
	Action  string  `json:"action"`
	Agent   string  `json:"agent"`
	Details *string `json:"details,omitempty"`
}

// UpdateStatutConvocationRequest représente la requête de mise à jour du statut
type UpdateStatutConvocationRequest struct {
	Statut           StatutConvocation `json:"statut" validate:"required"`
	Motif            *string           `json:"motif,omitempty"`
	Commentaire      *string           `json:"commentaire,omitempty"`
	DateEnvoi        *time.Time        `json:"dateEnvoi,omitempty"`
	DateHonoration   *time.Time        `json:"dateHonoration,omitempty"`
	ResultatAudition *string           `json:"resultatAudition,omitempty"`
	Observations     *string           `json:"observations,omitempty"`
}

// FilterConvocationsRequest représente les filtres pour la liste des convocations
type FilterConvocationsRequest struct {
	Statut          *string    `json:"statut,omitempty"`
	TypeConvocation *string    `json:"typeConvocation,omitempty"`
	QualiteConvoque *string    `json:"qualiteConvoque,omitempty"`
	CommissariatID  *string    `json:"commissariatId,omitempty"`
	DateDebut       *time.Time `json:"dateDebut,omitempty"`
	DateFin         *time.Time `json:"dateFin,omitempty"`
	Search          *string    `json:"search,omitempty"`
	Page            int        `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit           int        `json:"limit,omitempty" validate:"omitempty,min=1"`
}

// ListConvocationsResponse représente la réponse de liste avec pagination
type ListConvocationsResponse struct {
	Convocations []ConvocationResponse `json:"convocations"`
	Pagination   PaginationInfo        `json:"pagination"`
}

// PaginationInfo représente les informations de pagination
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// StatistiquesConvocationsResponse représente les statistiques des convocations
type StatistiquesConvocationsResponse struct {
	TotalConvocations     int64   `json:"totalConvocations"`
	ConvocationsJour      int64   `json:"convocationsJour"`
	Envoyes               int64   `json:"envoyes"`
	Honores               int64   `json:"honores"`
	EnAttente             int64   `json:"enAttente"`
	PourcentageHonores    float64 `json:"pourcentageHonores"`
	EvolutionConvocations string  `json:"evolutionConvocations"`
	EvolutionEnvoyes      string  `json:"evolutionEnvoyes"`
	EvolutionHonores      string  `json:"evolutionHonores"`
	DelaiMoyen            *string `json:"delaiMoyen,omitempty"`
	AgentsActifs          *int    `json:"agentsActifs,omitempty"`
	Nouvelles             *int    `json:"nouvelles,omitempty"`
}

// DashboardConvocationsResponse représente la réponse complète du dashboard
type DashboardConvocationsResponse struct {
	Stats        DashboardStats          `json:"stats"`
	ActivityData []DashboardActivityData `json:"activityData"`
	PieData      []PieDataEntry          `json:"pieData"`
	TopTypes     []TopTypesEntry         `json:"topTypes"`
}

// DashboardStats représente les statistiques du dashboard
type DashboardStats struct {
	TotalConvocations     int64   `json:"totalConvocations"`
	Envoyees              int64   `json:"envoyees"`
	Honorees              int64   `json:"honorees"`
	EnAttente             int64   `json:"enAttente"`
	DelaiMoyenJours       float64 `json:"delaiMoyenJours"`
	TauxHonore            float64 `json:"tauxHonore"`
	AgentsActifsCount     int64   `json:"agentsActifsCount"`
	TotalAgents           int64   `json:"totalAgents"`
	Nouvelles             int64   `json:"nouvelles"`
	EvolutionConvocations string  `json:"evolutionConvocations"`
	EvolutionEnvoyees     string  `json:"evolutionEnvoyees"`
	EvolutionHonorees     string  `json:"evolutionHonorees"`
	EvolutionEnAttente    string  `json:"evolutionEnAttente"`
	EvolutionDelai        string  `json:"evolutionDelai"`
	EvolutionTauxHonore   string  `json:"evolutionTauxHonore"`
	EvolutionNouvelles    string  `json:"evolutionNouvelles"`
}

// DashboardActivityData représente les données d'activité par période
type DashboardActivityData struct {
	Period       string `json:"period"`
	Convocations int    `json:"convocations"`
	Envoyees     int    `json:"envoyees"`
	Honorees     int    `json:"honorees"`
}

// PieDataEntry représente une entrée de données pour le graphique en camembert
type PieDataEntry struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

// TopTypesEntry représente les types de convocations les plus communs
type TopTypesEntry struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ReporterRdvRequest représente une demande de report de rendez-vous
type ReporterRdvRequest struct {
	NouvelleDate  string `json:"nouvelleDate" validate:"required"`
	NouvelleHeure string `json:"nouvelleHeure" validate:"required"`
	Motif         string `json:"motif" validate:"required"`
}

// NotifierRequest représente une demande de notification
type NotifierRequest struct {
	Moyens  []string `json:"moyens" validate:"required,min=1"`
	Message *string  `json:"message,omitempty"`
}

// AjouterNoteRequest représente une demande d'ajout de note
type AjouterNoteRequest struct {
	Note string `json:"note" validate:"required"`
}
