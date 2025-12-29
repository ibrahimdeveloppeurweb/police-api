package convocations

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/convocation"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines convocations service interface
type Service interface {
	Create(ctx context.Context, req *CreateConvocationRequest, agentID, commissariatID string) (*ConvocationResponse, error)
	GetByID(ctx context.Context, id string) (*ConvocationResponse, error)
	List(ctx context.Context, filters *FilterConvocationsRequest, role, userID, commissariatID string) (*ListConvocationsResponse, error)
	UpdateStatut(ctx context.Context, id string, req *UpdateStatutConvocationRequest, agentID string) (*ConvocationResponse, error)
	ReporterRdv(ctx context.Context, id string, req *ReporterRdvRequest, agentID string) (*ConvocationResponse, error)
	Notifier(ctx context.Context, id string, req *NotifierRequest, agentID string) (*ConvocationResponse, error)
	AjouterNote(ctx context.Context, id string, req *AjouterNoteRequest, agentID string) (*ConvocationResponse, error)
	GeneratePDF(ctx context.Context, id string) ([]byte, error)
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesConvocationsResponse, error)
	GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardConvocationsResponse, error)
}

// service implements Service interface
type service struct {
	convocationRepo  repository.ConvocationRepository
	commissariatRepo repository.CommissariatRepository
	userRepo         repository.UserRepository
	config           *config.Config
	logger           *zap.Logger
}

// NewService creates a new convocations service
func NewService(
	convocationRepo repository.ConvocationRepository,
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return &service{
		convocationRepo:  convocationRepo,
		commissariatRepo: commissariatRepo,
		userRepo:         userRepo,
		config:           cfg,
		logger:           logger,
	}
}

// generateNumero génère un numéro unique pour la convocation
func (s *service) generateNumero(ctx context.Context, commissariatID string) (string, error) {
	// Récupérer le commissariat pour obtenir la ville
	commissariat, err := s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return "", fmt.Errorf("commissariat not found")
	}

	year := time.Now().Year()

	// Extraire les 3 premières lettres de la ville en majuscules
	ville := strings.ToUpper(commissariat.Ville)
	villePrefix := ville
	if len(ville) > 3 {
		villePrefix = ville[:3]
	} else if len(ville) < 3 {
		// Si la ville a moins de 3 lettres, compléter avec des X
		villePrefix = ville + strings.Repeat("X", 3-len(ville))
	}

	// Chercher le dernier numéro de convocation pour cette année
	filters := &repository.ConvocationFilters{
		CommissariatID: &commissariatID,
		Limit:          1000,
		Offset:         0,
	}

	convocations, err := s.convocationRepo.List(ctx, filters)

	nextNumber := 1

	if err == nil && len(convocations) > 0 {
		maxNum := 0
		for _, conv := range convocations {
			// Format: CONV-VILLE-COM-YYYY-NNNN
			parts := strings.Split(conv.Numero, "-")
			if len(parts) == 5 {
				if num, err := strconv.Atoi(parts[4]); err == nil && num > maxNum {
					maxNum = num
				}
			}
		}
		if maxNum > 0 {
			nextNumber = maxNum + 1
		}
	}

	maxRetries := 10
	for retry := 0; retry < maxRetries; retry++ {
		numero := fmt.Sprintf("CONV-%s-COM-%d-%04d", villePrefix, year, nextNumber+retry)
		_, err := s.convocationRepo.GetByNumero(ctx, numero)
		if err != nil {
			return numero, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique numero after %d retries", maxRetries)
}

// Create creates a new convocation with ALL 74 fields
func (s *service) Create(ctx context.Context, req *CreateConvocationRequest, agentID, commissariatID string) (*ConvocationResponse, error) {
	// Validation basique des champs obligatoires
	if req.Nom == "" || req.Prenom == "" || req.Telephone1 == "" {
		return nil, fmt.Errorf("nom, prenom et telephone1 sont obligatoires")
	}
	if req.TypeConvocation == "" {
		return nil, fmt.Errorf("typeConvocation est obligatoire")
	}
	if req.StatutPersonne == "" {
		return nil, fmt.Errorf("statutPersonne est obligatoire")
	}
	if req.TypePiece == "" || req.NumeroPiece == "" {
		return nil, fmt.Errorf("typePiece et numeroPiece sont obligatoires")
	}
	if req.LieuRdv == "" {
		return nil, fmt.Errorf("lieuRdv est obligatoire")
	}
	if req.Motif == "" {
		return nil, fmt.Errorf("motif est obligatoire")
	}

	// Générer le numéro unique
	numero, err := s.generateNumero(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate numero: %w", err)
	}

	// Parser la date de création
	dateCreation, err := time.Parse("2006-01-02", req.DateCreation)
	if err != nil {
		return nil, fmt.Errorf("invalid date format for dateCreation: %w", err)
	}

	// Parser la date de RDV si fournie
	var dateRdv *time.Time
	if req.DateRdv != nil && *req.DateRdv != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateRdv)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for dateRdv: %w", err)
		}
		dateRdv = &parsed
	}

	// Créer le builder avec TOUS les champs
	createBuilder := s.convocationRepo.Client().Convocation.Create().
		SetNumero(numero).
		SetCommissariatID(uuid.MustParse(commissariatID)).
		SetAgentID(uuid.MustParse(agentID))

	// SECTION 1: INFORMATIONS GÉNÉRALES
	if req.Reference != "" {
		createBuilder.SetReference(req.Reference)
	}
	createBuilder.SetTypeConvocation(req.TypeConvocation)
	if req.SousType != nil {
		createBuilder.SetSousType(*req.SousType)
	}
	createBuilder.SetUrgence(convocation.Urgence(req.Urgence))
	createBuilder.SetPriorite(convocation.Priorite(req.Priorite))
	createBuilder.SetConfidentialite(convocation.Confidentialite(req.Confidentialite))

	// SECTION 2: AFFAIRE LIÉE
	if req.AffaireID != nil {
		createBuilder.SetAffaireID(*req.AffaireID)
	}
	if req.AffaireType != nil {
		createBuilder.SetAffaireType(*req.AffaireType)
	}
	if req.AffaireNumero != nil {
		createBuilder.SetAffaireNumero(*req.AffaireNumero)
		createBuilder.SetAffaireLiee(*req.AffaireNumero) // Alias
	}
	if req.AffaireTitre != nil {
		createBuilder.SetAffaireTitre(*req.AffaireTitre)
	}
	if req.SectionJudiciaire != nil {
		createBuilder.SetSectionJudiciaire(*req.SectionJudiciaire)
	}
	if req.Infraction != nil {
		createBuilder.SetInfraction(*req.Infraction)
	}
	if req.QualificationLegale != nil {
		createBuilder.SetQualificationLegale(*req.QualificationLegale)
	}

	// SECTION 3: PERSONNE CONVOQUÉE - Identité
	createBuilder.SetStatutPersonne(req.StatutPersonne)
	createBuilder.SetQualiteConvoque(req.StatutPersonne) // Alias
	createBuilder.SetConvoqueNom(req.Nom)
	createBuilder.SetConvoquePrenom(req.Prenom)
	if req.DateNaissance != nil {
		createBuilder.SetDateNaissance(*req.DateNaissance)
	}
	if req.LieuNaissance != nil {
		createBuilder.SetLieuNaissance(*req.LieuNaissance)
	}
	if req.Nationalite != nil {
		createBuilder.SetNationalite(*req.Nationalite)
	}

	// SECTION 3: Pièce d'identité
	createBuilder.SetTypePiece(req.TypePiece)
	createBuilder.SetNumeroPiece(req.NumeroPiece)
	if req.DateDelivrancePiece != nil {
		createBuilder.SetDateDelivrancePiece(*req.DateDelivrancePiece)
	}
	if req.LieuDelivrancePiece != nil {
		createBuilder.SetLieuDelivrancePiece(*req.LieuDelivrancePiece)
	}
	if req.DateExpirationPiece != nil {
		createBuilder.SetDateExpirationPiece(*req.DateExpirationPiece)
	}

	// SECTION 3: Contact
	createBuilder.SetConvoqueTelephone(req.Telephone1)
	if req.Telephone2 != nil {
		createBuilder.SetConvoqueTelephone2(*req.Telephone2)
	}
	if req.Email != nil {
		createBuilder.SetConvoqueEmail(*req.Email)
	}
	if req.AdresseResidence != nil {
		createBuilder.SetAdresseResidence(*req.AdresseResidence)
		createBuilder.SetConvoqueAdresse(*req.AdresseResidence) // Alias
	}
	if req.AdresseProfessionnelle != nil {
		createBuilder.SetAdresseProfessionnelle(*req.AdresseProfessionnelle)
	}
	if req.DernierLieuConnu != nil {
		createBuilder.SetDernierLieuConnu(*req.DernierLieuConnu)
	}

	// SECTION 3: Informations complémentaires
	if req.Profession != nil {
		createBuilder.SetProfession(*req.Profession)
	}
	if req.SituationFamiliale != nil {
		createBuilder.SetSituationFamiliale(*req.SituationFamiliale)
	}
	if req.NombreEnfants != nil {
		createBuilder.SetNombreEnfants(*req.NombreEnfants)
	}
	if req.Sexe != nil {
		createBuilder.SetSexe(*req.Sexe)
	}
	if req.Taille != nil {
		createBuilder.SetTaille(*req.Taille)
	}
	if req.Poids != nil {
		createBuilder.SetPoids(*req.Poids)
	}
	if req.SignesParticuliers != nil {
		createBuilder.SetSignesParticuliers(*req.SignesParticuliers)
	}
	createBuilder.SetPhotoIdentite(req.PhotoIdentite)
	createBuilder.SetEmpreintes(req.Empreintes)

	// SECTION 4: RENDEZ-VOUS
	createBuilder.SetDateCreation(dateCreation)
	if req.HeureConvocation != nil {
		createBuilder.SetHeureConvocation(*req.HeureConvocation)
	}
	if dateRdv != nil {
		createBuilder.SetDateRdv(*dateRdv)
	}
	if req.HeureRdv != nil {
		createBuilder.SetHeureRdv(*req.HeureRdv)
	}
	if req.DureeEstimee != nil {
		createBuilder.SetDureeEstimee(*req.DureeEstimee)
	}
	createBuilder.SetTypeAudience(req.TypeAudience)

	// SECTION 4: Lieu
	createBuilder.SetLieuRdv(req.LieuRdv)
	if req.Bureau != nil {
		createBuilder.SetBureau(*req.Bureau)
	}
	if req.SalleAudience != nil {
		createBuilder.SetSalleAudience(*req.SalleAudience)
	}
	if req.PointRencontre != nil {
		createBuilder.SetPointRencontre(*req.PointRencontre)
	}
	if req.AccesSpecifique != nil {
		createBuilder.SetAccesSpecifique(*req.AccesSpecifique)
	}

	// SECTION 5: PERSONNES PRÉSENTES
	createBuilder.SetConvocateurNom(req.ConvocateurNom)
	createBuilder.SetConvocateurPrenom(req.ConvocateurPrenom)
	if req.ConvocateurMatricule != nil {
		createBuilder.SetConvocateurMatricule(*req.ConvocateurMatricule)
	}
	if req.ConvocateurFonction != nil {
		createBuilder.SetConvocateurFonction(*req.ConvocateurFonction)
	}
	if req.AgentsPresents != nil {
		createBuilder.SetAgentsPresents(*req.AgentsPresents)
	}
	createBuilder.SetRepresentantParquet(req.RepresentantParquet)
	if req.NomParquetier != nil {
		createBuilder.SetNomParquetier(*req.NomParquetier)
	}
	createBuilder.SetExpertPresent(req.ExpertPresent)
	if req.TypeExpert != nil {
		createBuilder.SetTypeExpert(*req.TypeExpert)
	}
	createBuilder.SetInterpreteNecessaire(req.InterpreteNecessaire)
	if req.LangueInterpretation != nil {
		createBuilder.SetLangueInterpretation(*req.LangueInterpretation)
	}
	createBuilder.SetAvocatPresent(req.AvocatPresent)
	if req.NomAvocat != nil {
		createBuilder.SetNomAvocat(*req.NomAvocat)
	}
	if req.BarreauAvocat != nil {
		createBuilder.SetBarreauAvocat(*req.BarreauAvocat)
	}

	// SECTION 6: MOTIF ET OBJET
	createBuilder.SetMotif(req.Motif)
	if req.ObjetPrecis != nil {
		createBuilder.SetObjetPrecis(*req.ObjetPrecis)
	}
	if req.QuestionsPreparatoires != nil {
		createBuilder.SetQuestionsPreparatoires(*req.QuestionsPreparatoires)
	}
	if req.PiecesAApporter != nil {
		createBuilder.SetPiecesAApporter(*req.PiecesAApporter)
	}
	if req.DocumentsDemandes != nil {
		createBuilder.SetDocumentsDemandes(*req.DocumentsDemandes)
	}

	// SECTION 9: OBSERVATIONS
	if req.Observations != nil {
		createBuilder.SetObservations(*req.Observations)
	}

	// SECTION 10: ÉTAT ET TRAÇABILITÉ
	if req.Statut != nil && *req.Statut != "" {
		createBuilder.SetStatut(convocation.Statut(*req.Statut))
	} else {
		createBuilder.SetStatut(convocation.StatutCRÉATION) // Statut par défaut
	}
	createBuilder.SetModeEnvoi(req.ModeEnvoi)

	// Construire les données complètes en JSON
	donneesCompletes := map[string]interface{}{
		"reference":            req.Reference,
		"typeConvocation":      req.TypeConvocation,
		"urgence":              req.Urgence,
		"priorite":             req.Priorite,
		"confidentialite":      req.Confidentialite,
		"statutPersonne":       req.StatutPersonne,
		"nom":                  req.Nom,
		"prenom":               req.Prenom,
		"typePiece":            req.TypePiece,
		"numeroPiece":          req.NumeroPiece,
		"telephone1":           req.Telephone1,
		"typeAudience":         req.TypeAudience,
		"lieuRdv":              req.LieuRdv,
		"motif":                req.Motif,
		"convocateurNom":       req.ConvocateurNom,
		"convocateurPrenom":    req.ConvocateurPrenom,
		"photoIdentite":        req.PhotoIdentite,
		"empreintes":           req.Empreintes,
		"representantParquet":  req.RepresentantParquet,
		"expertPresent":        req.ExpertPresent,
		"interpreteNecessaire": req.InterpreteNecessaire,
		"avocatPresent":        req.AvocatPresent,
		"modeEnvoi":            req.ModeEnvoi,
	}

	// Ajouter le statut si présent
	if req.Statut != nil && *req.Statut != "" {
		donneesCompletes["statut"] = *req.Statut
	} else {
		donneesCompletes["statut"] = "CRÉATION"
	}

	// Ajouter tous les champs optionnels
	addOptionalField := func(key string, value *string) {
		if value != nil {
			donneesCompletes[key] = *value
		}
	}

	addOptionalField("sousType", req.SousType)
	addOptionalField("affaireId", req.AffaireID)
	addOptionalField("affaireType", req.AffaireType)
	addOptionalField("affaireNumero", req.AffaireNumero)
	addOptionalField("affaireTitre", req.AffaireTitre)
	addOptionalField("sectionJudiciaire", req.SectionJudiciaire)
	addOptionalField("infraction", req.Infraction)
	addOptionalField("qualificationLegale", req.QualificationLegale)
	addOptionalField("dateNaissance", req.DateNaissance)
	addOptionalField("lieuNaissance", req.LieuNaissance)
	addOptionalField("nationalite", req.Nationalite)
	addOptionalField("profession", req.Profession)
	addOptionalField("situationFamiliale", req.SituationFamiliale)
	addOptionalField("nombreEnfants", req.NombreEnfants)
	addOptionalField("dateDelivrancePiece", req.DateDelivrancePiece)
	addOptionalField("lieuDelivrancePiece", req.LieuDelivrancePiece)
	addOptionalField("dateExpirationPiece", req.DateExpirationPiece)
	addOptionalField("telephone2", req.Telephone2)
	addOptionalField("email", req.Email)
	addOptionalField("adresseResidence", req.AdresseResidence)
	addOptionalField("adresseProfessionnelle", req.AdresseProfessionnelle)
	addOptionalField("dernierLieuConnu", req.DernierLieuConnu)
	addOptionalField("sexe", req.Sexe)
	addOptionalField("taille", req.Taille)
	addOptionalField("poids", req.Poids)
	addOptionalField("signesParticuliers", req.SignesParticuliers)
	addOptionalField("heureConvocation", req.HeureConvocation)
	addOptionalField("dateRdv", req.DateRdv)
	addOptionalField("heureRdv", req.HeureRdv)
	addOptionalField("bureau", req.Bureau)
	addOptionalField("salleAudience", req.SalleAudience)
	addOptionalField("pointRencontre", req.PointRencontre)
	addOptionalField("accesSpecifique", req.AccesSpecifique)
	addOptionalField("convocateurMatricule", req.ConvocateurMatricule)
	addOptionalField("convocateurFonction", req.ConvocateurFonction)
	addOptionalField("agentsPresents", req.AgentsPresents)
	addOptionalField("nomParquetier", req.NomParquetier)
	addOptionalField("typeExpert", req.TypeExpert)
	addOptionalField("langueInterpretation", req.LangueInterpretation)
	addOptionalField("nomAvocat", req.NomAvocat)
	addOptionalField("barreauAvocat", req.BarreauAvocat)
	addOptionalField("objetPrecis", req.ObjetPrecis)
	addOptionalField("questionsPreparatoires", req.QuestionsPreparatoires)
	addOptionalField("piecesAApporter", req.PiecesAApporter)
	addOptionalField("documentsDemandes", req.DocumentsDemandes)
	addOptionalField("observations", req.Observations)

	if req.DureeEstimee != nil {
		donneesCompletes["dureeEstimee"] = *req.DureeEstimee
	}
	createBuilder.SetDonneesCompletes(donneesCompletes)

	// Créer l'historique initial
	historiqueInitial := []map[string]interface{}{
		{
			"date":    time.Now().Format("02/01/2006 15:04"),
			"dateISO": time.Now().Format(time.RFC3339),
			"action":  "Création",
			"agent":   fmt.Sprintf("%s %s", req.ConvocateurPrenom, req.ConvocateurNom),
			"details": fmt.Sprintf("Convocation créée pour %s %s", req.Prenom, req.Nom),
		},
	}
	createBuilder.SetHistorique(historiqueInitial)

	// Créer la convocation
	created, err := createBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("failed to create convocation", zap.Error(err))
		return nil, fmt.Errorf("failed to create convocation: %w", err)
	}

	s.logger.Info("convocation created successfully",
		zap.String("id", created.ID.String()),
		zap.String("numero", created.Numero),
		zap.String("nom", req.Nom),
	)

	// Récupérer la convocation avec les relations
	return s.GetByID(ctx, created.ID.String())
}

// GetByID retrieves a convocation by ID
func (s *service) GetByID(ctx context.Context, id string) (*ConvocationResponse, error) {
	convocID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid convocation ID: %w", err)
	}

	conv, err := s.convocationRepo.GetByID(ctx, convocID.String())
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	return s.toResponse(conv), nil
}

// List retrieves convocations with filters
func (s *service) List(ctx context.Context, filters *FilterConvocationsRequest, role, userID, commissariatID string) (*ListConvocationsResponse, error) {
	repoFilters := &repository.ConvocationFilters{
		Statut:          filters.Statut,
		TypeConvocation: filters.TypeConvocation,
		DateDebut:       filters.DateDebut,
		DateFin:         filters.DateFin,
		Search:          filters.Search,
		Page:            filters.Page,
		Limit:           filters.Limit,
	}

	if role != "ADMIN" {
		repoFilters.CommissariatID = &commissariatID
	} else if filters.CommissariatID != nil {
		repoFilters.CommissariatID = filters.CommissariatID
	}

	convocations, err := s.convocationRepo.List(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list convocations: %w", err)
	}

	total, err := s.convocationRepo.Count(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count convocations: %w", err)
	}

	responses := make([]ConvocationResponse, len(convocations))
	for i, conv := range convocations {
		responses[i] = *s.toResponse(conv)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 10
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &ListConvocationsResponse{
		Convocations: responses,
		Pagination: PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateStatut updates convocation status
func (s *service) UpdateStatut(ctx context.Context, id string, req *UpdateStatutConvocationRequest, agentID string) (*ConvocationResponse, error) {
	convocID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid convocation ID: %w", err)
	}

	conv, err := s.convocationRepo.GetByID(ctx, convocID.String())
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	// Récupérer les informations de l'agent
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		s.logger.Warn("Failed to get agent info", zap.Error(err))
	}

	agentName := agentID
	if agent != nil {
		agentName = fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
	}

	updateBuilder := s.convocationRepo.Client().Convocation.
		UpdateOneID(convocID).
		SetStatut(convocation.Statut(req.Statut))

	if req.DateEnvoi != nil {
		updateBuilder.SetDateEnvoi(*req.DateEnvoi)
	}
	if req.DateHonoration != nil {
		updateBuilder.SetDateHonoration(*req.DateHonoration)
	}
	if req.ResultatAudition != nil {
		updateBuilder.SetResultatAudition(*req.ResultatAudition)
	}
	if req.Observations != nil {
		updateBuilder.SetObservations(*req.Observations)
	}

	var historique []map[string]interface{}
	if len(conv.Historique) > 0 {
		historique = conv.Historique
	}

	nouvelleEntree := map[string]interface{}{
		"date":    time.Now().Format("02/01/2006 15:04"),
		"dateISO": time.Now().Format(time.RFC3339),
		"action":  fmt.Sprintf("Changement de statut en %s", req.Statut),
		"agent":   agentName,
	}

	// Ajouter le commentaire dans les détails s'il est fourni
	if req.Commentaire != nil && *req.Commentaire != "" {
		nouvelleEntree["details"] = *req.Commentaire
	} else if req.Observations != nil {
		nouvelleEntree["details"] = *req.Observations
	}

	historique = append(historique, nouvelleEntree)
	updateBuilder.SetHistorique(historique)

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update convocation: %w", err)
	}

	return s.GetByID(ctx, id)
}

// GetStatistiques retrieves convocations statistics
func (s *service) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesConvocationsResponse, error) {
	// Calculer les plages de dates
	debut, fin := s.calculateDateRange(dateDebut, dateFin, periode)

	// Filtres pour la période actuelle
	filters := &repository.ConvocationFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}

	// Récupérer les convocations de la période actuelle
	convocations, err := s.convocationRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get convocations: %w", err)
	}

	// Calculer les statistiques
	total := int64(len(convocations))
	envoyes := int64(0)
	honores := int64(0)
	enAttente := int64(0)

	for _, conv := range convocations {
		switch conv.Statut {
		case "ENVOYÉ":
			envoyes++
		case "HONORÉ":
			honores++
		case "EN_ATTENTE":
			enAttente++
		}
	}

	// Calculer le taux de convocations honorées
	pourcentageHonores := 0.0
	if total > 0 {
		pourcentageHonores = (float64(honores) / float64(total)) * 100
	}

	// Calculer les évolutions par rapport à la période précédente
	debutPrecedent, finPrecedent := s.calculatePreviousPeriod(debut, fin, periode)
	filtersPrecedent := &repository.ConvocationFilters{
		CommissariatID: commissariatID,
		DateDebut:      debutPrecedent,
		DateFin:        finPrecedent,
	}

	convocationsPrecedentes, err := s.convocationRepo.List(ctx, filtersPrecedent)
	totalPrecedent := int64(0)
	envoyesPrecedent := int64(0)
	honoresPrecedent := int64(0)

	if err == nil {
		totalPrecedent = int64(len(convocationsPrecedentes))
		for _, conv := range convocationsPrecedentes {
			switch conv.Statut {
			case "ENVOYÉ":
				envoyesPrecedent++
			case "HONORÉ":
				honoresPrecedent++
			}
		}
	}

	// Calculer les évolutions en pourcentage
	evolutionConvocations := s.calculateEvolution(total, totalPrecedent)
	evolutionEnvoyes := s.calculateEvolution(envoyes, envoyesPrecedent)
	evolutionHonores := s.calculateEvolution(honores, honoresPrecedent)

	// Convocations du jour
	aujourdhui := time.Now()
	debutJour := time.Date(aujourdhui.Year(), aujourdhui.Month(), aujourdhui.Day(), 0, 0, 0, 0, time.UTC)
	finJour := time.Date(aujourdhui.Year(), aujourdhui.Month(), aujourdhui.Day(), 23, 59, 59, 0, time.UTC)

	filtersJour := &repository.ConvocationFilters{
		CommissariatID: commissariatID,
		DateDebut:      &debutJour,
		DateFin:        &finJour,
	}

	convocationsJour, err := s.convocationRepo.List(ctx, filtersJour)
	convocationsJourCount := int64(0)
	if err == nil {
		convocationsJourCount = int64(len(convocationsJour))
	}

	return &StatistiquesConvocationsResponse{
		TotalConvocations:     total,
		ConvocationsJour:      convocationsJourCount,
		Envoyes:               envoyes,
		Honores:               honores,
		EnAttente:             enAttente,
		PourcentageHonores:    pourcentageHonores,
		EvolutionConvocations: evolutionConvocations,
		EvolutionEnvoyes:      evolutionEnvoyes,
		EvolutionHonores:      evolutionHonores,
	}, nil
}

// GetDashboard retrieves dashboard data
func (s *service) GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardConvocationsResponse, error) {
	// Calculer les plages de dates
	debut, fin := s.calculateDateRange(dateDebut, dateFin, periode)

	// Filtres pour la période actuelle
	filters := &repository.ConvocationFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}

	// Récupérer les convocations de la période actuelle
	convocations, err := s.convocationRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get convocations: %w", err)
	}

	// Calculer les statistiques
	total := int64(len(convocations))
	envoyees := int64(0)
	honorees := int64(0)
	enAttente := int64(0)

	for _, conv := range convocations {
		switch conv.Statut {
		case "ENVOYÉ":
			envoyees++
		case "HONORÉ":
			honorees++
		case "EN_ATTENTE":
			enAttente++
		}
	}

	// Calculer le taux honoré
	tauxHonore := 0.0
	if total > 0 {
		tauxHonore = (float64(honorees) / float64(total)) * 100
	}

	// Calculer les évolutions
	debutPrecedent, finPrecedent := s.calculatePreviousPeriod(debut, fin, periode)
	filtersPrecedent := &repository.ConvocationFilters{
		CommissariatID: commissariatID,
		DateDebut:      debutPrecedent,
		DateFin:        finPrecedent,
	}

	convocationsPrecedentes, _ := s.convocationRepo.List(ctx, filtersPrecedent)
	totalPrecedent := int64(len(convocationsPrecedentes))
	envoyeesPrecedent := int64(0)
	honoreesPrecedent := int64(0)
	enAttentePrecedent := int64(0)

	for _, conv := range convocationsPrecedentes {
		switch conv.Statut {
		case "ENVOYÉ":
			envoyeesPrecedent++
		case "HONORÉ":
			honoreesPrecedent++
		case "EN_ATTENTE":
			enAttentePrecedent++
		}
	}

	// Générer les données d'activité
	activityData := s.generateActivityData(ctx, convocations, periode)

	// Générer les données du graphique en camembert
	pieData := []PieDataEntry{
		{Name: "Envoyées", Value: int(envoyees), Color: "#3b82f6"},
		{Name: "Honorées", Value: int(honorees), Color: "#10b981"},
		{Name: "En attente", Value: int(enAttente), Color: "#f59e0b"},
	}

	// Générer les types les plus fréquents
	topTypes := s.calculateTopTypes(convocations)

	return &DashboardConvocationsResponse{
		Stats: DashboardStats{
			TotalConvocations:     total,
			Envoyees:              envoyees,
			Honorees:              honorees,
			EnAttente:             enAttente,
			DelaiMoyenJours:       0.0,
			TauxHonore:            tauxHonore,
			AgentsActifsCount:     0,
			TotalAgents:           0,
			Nouvelles:             0,
			EvolutionConvocations: s.calculateEvolution(total, totalPrecedent),
			EvolutionEnvoyees:     s.calculateEvolution(envoyees, envoyeesPrecedent),
			EvolutionHonorees:     s.calculateEvolution(honorees, honoreesPrecedent),
			EvolutionEnAttente:    s.calculateEvolution(enAttente, enAttentePrecedent),
			EvolutionDelai:        "+0",
			EvolutionTauxHonore:   "+0",
			EvolutionNouvelles:    "+0",
		},
		ActivityData: activityData,
		PieData:      pieData,
		TopTypes:     topTypes,
	}, nil
}

// calculateDateRange calcule les dates de début et fin selon la période
func (s *service) calculateDateRange(dateDebut, dateFin, periode *string) (*time.Time, *time.Time) {
	if dateDebut != nil && dateFin != nil {
		debut, err1 := parseDateTime(*dateDebut)
		fin, err2 := parseDateTime(*dateFin)
		if err1 == nil && err2 == nil {
			return &debut, &fin
		}
	}

	if periode == nil || *periode == "" {
		periodeStr := "jour"
		periode = &periodeStr
	}

	now := time.Now().UTC()
	var debut, fin time.Time

	switch *periode {
	case "jour":
		debut = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		fin = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	case "semaine":
		dayOfWeek := int(now.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		debut = now.AddDate(0, 0, -dayOfWeek+1)
		debut = time.Date(debut.Year(), debut.Month(), debut.Day(), 0, 0, 0, 0, time.UTC)
		fin = now
	case "mois":
		debut = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		fin = now
	case "annee":
		debut = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		fin = now
	default:
		debut = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		fin = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	}

	return &debut, &fin
}

// calculatePreviousPeriod calcule la période précédente
func (s *service) calculatePreviousPeriod(debut, fin *time.Time, periode *string) (*time.Time, *time.Time) {
	if debut == nil || fin == nil {
		return nil, nil
	}

	duration := fin.Sub(*debut)
	debutPrecedent := debut.Add(-duration)
	finPrecedent := debut.Add(-time.Second)

	return &debutPrecedent, &finPrecedent
}

// calculateEvolution calcule l'évolution en nombre absolu avec le signe
func (s *service) calculateEvolution(current, previous int64) string {
	difference := current - previous

	if difference > 0 {
		return fmt.Sprintf("+%d", difference)
	} else if difference < 0 {
		return fmt.Sprintf("%d", difference)
	}
	return "+0"
}

// generateActivityData génère les données d'activité par période
func (s *service) generateActivityData(ctx context.Context, convocations []*ent.Convocation, periode *string) []DashboardActivityData {
	if periode == nil {
		periodeStr := "jour"
		periode = &periodeStr
	}

	var activityData []DashboardActivityData

	switch *periode {
	case "jour":
		// Par tranches de 4 heures
		tranches := []struct {
			label      string
			heureDebut int
			heureFin   int
		}{
			{"00h-04h", 0, 4},
			{"04h-08h", 4, 8},
			{"08h-12h", 8, 12},
			{"12h-16h", 12, 16},
			{"16h-20h", 16, 20},
			{"20h-24h", 20, 24},
		}

		for _, tranche := range tranches {
			totalConvocations := 0
			envoyees := 0
			honorees := 0

			for _, conv := range convocations {
				heure := conv.CreatedAt.Hour()
				if heure >= tranche.heureDebut && heure < tranche.heureFin {
					totalConvocations++
					if conv.Statut == "ENVOYÉ" {
						envoyees++
					} else if conv.Statut == "HONORÉ" {
						honorees++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       tranche.label,
				Convocations: totalConvocations,
				Envoyees:     envoyees,
				Honorees:     honorees,
			})
		}

	case "semaine":
		// Par jour de la semaine
		jours := []struct {
			label string
			jour  time.Weekday
		}{
			{"Lun", time.Monday},
			{"Mar", time.Tuesday},
			{"Mer", time.Wednesday},
			{"Jeu", time.Thursday},
			{"Ven", time.Friday},
			{"Sam", time.Saturday},
			{"Dim", time.Sunday},
		}

		for _, j := range jours {
			totalConvocations := 0
			envoyees := 0
			honorees := 0

			for _, conv := range convocations {
				if conv.CreatedAt.Weekday() == j.jour {
					totalConvocations++
					if conv.Statut == "ENVOYÉ" {
						envoyees++
					} else if conv.Statut == "HONORÉ" {
						honorees++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       j.label,
				Convocations: totalConvocations,
				Envoyees:     envoyees,
				Honorees:     honorees,
			})
		}

	case "mois":
		// Par semaine du mois
		for i := 1; i <= 4; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period:       fmt.Sprintf("Sem %d", i),
				Convocations: 0,
				Envoyees:     0,
				Honorees:     0,
			})
		}

	case "annee":
		// Par mois de l'année
		mois := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
		for _, m := range mois {
			activityData = append(activityData, DashboardActivityData{
				Period:       m,
				Convocations: 0,
				Envoyees:     0,
				Honorees:     0,
			})
		}
	}

	return activityData
}

// calculateTopTypes calcule les types les plus fréquents
func (s *service) calculateTopTypes(convocations []*ent.Convocation) []TopTypesEntry {
	typesCount := make(map[string]int)

	for _, conv := range convocations {
		typesCount[conv.TypeConvocation]++
	}

	var topTypes []TopTypesEntry
	for typeConv, count := range typesCount {
		topTypes = append(topTypes, TopTypesEntry{
			Type:  typeConv,
			Count: count,
		})
	}

	// Trier par ordre décroissant
	for i := 0; i < len(topTypes)-1; i++ {
		for j := i + 1; j < len(topTypes); j++ {
			if topTypes[j].Count > topTypes[i].Count {
				topTypes[i], topTypes[j] = topTypes[j], topTypes[i]
			}
		}
	}

	// Garder seulement le top 5
	if len(topTypes) > 5 {
		topTypes = topTypes[:5]
	}

	return topTypes
}

// parseDateTime est une fonction helper pour parser les dates
func parseDateTime(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// toResponse converts ent.Convocation to ConvocationResponse with ALL 74 fields
func (s *service) toResponse(conv *ent.Convocation) *ConvocationResponse {
	// Helper function
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	response := &ConvocationResponse{
		// Identifiants
		ID:     conv.ID.String(),
		Numero: conv.Numero,

		// SECTION 1: INFORMATIONS GÉNÉRALES
		Reference:       conv.Reference,
		TypeConvocation: conv.TypeConvocation,
		SousType:        conv.SousType,
		Urgence:         strPtr(string(conv.Urgence)),
		Priorite:        strPtr(string(conv.Priorite)),
		Confidentialite: strPtr(string(conv.Confidentialite)),

		// SECTION 2: AFFAIRE LIÉE
		AffaireID:           conv.AffaireID,
		AffaireType:         conv.AffaireType,
		AffaireNumero:       conv.AffaireNumero,
		AffaireTitre:        conv.AffaireTitre,
		SectionJudiciaire:   conv.SectionJudiciaire,
		Infraction:          conv.Infraction,
		QualificationLegale: conv.QualificationLegale,

		// SECTION 3: PERSONNE CONVOQUÉE - Identité
		StatutPersonne:     conv.StatutPersonne,
		ConvoqueNom:        conv.ConvoqueNom,
		ConvoquePrenom:     conv.ConvoquePrenom,
		DateNaissance:      conv.DateNaissance,
		LieuNaissance:      conv.LieuNaissance,
		Nationalite:        conv.Nationalite,
		Profession:         conv.Profession,
		SituationFamiliale: conv.SituationFamiliale,
		NombreEnfants:      conv.NombreEnfants,

		// SECTION 3: Pièce d'identité
		TypePiece:           conv.TypePiece,
		NumeroPiece:         conv.NumeroPiece,
		DateDelivrancePiece: conv.DateDelivrancePiece,
		LieuDelivrancePiece: conv.LieuDelivrancePiece,
		DateExpirationPiece: conv.DateExpirationPiece,

		// SECTION 3: Contact
		ConvoqueTelephone:      conv.ConvoqueTelephone,
		ConvoqueTelephone2:     conv.ConvoqueTelephone2,
		ConvoqueEmail:          conv.ConvoqueEmail,
		AdresseResidence:       conv.AdresseResidence,
		AdresseProfessionnelle: conv.AdresseProfessionnelle,
		DernierLieuConnu:       conv.DernierLieuConnu,

		// SECTION 3: Caractéristiques physiques
		Sexe:               conv.Sexe,
		Taille:             conv.Taille,
		Poids:              conv.Poids,
		SignesParticuliers: conv.SignesParticuliers,
		PhotoIdentite:      conv.PhotoIdentite,
		Empreintes:         conv.Empreintes,

		// SECTION 4: RENDEZ-VOUS
		DateCreation:     conv.DateCreation,
		HeureConvocation: conv.HeureConvocation,
		DateRdv:          conv.DateRdv,
		HeureRdv:         conv.HeureRdv,
		DureeEstimee:     conv.DureeEstimee,
		TypeAudience:     strPtr(conv.TypeAudience),

		// SECTION 4: Lieu
		LieuRdv:         conv.LieuRdv,
		Bureau:          conv.Bureau,
		SalleAudience:   conv.SalleAudience,
		PointRencontre:  conv.PointRencontre,
		AccesSpecifique: conv.AccesSpecifique,

		// SECTION 5: PERSONNES PRÉSENTES
		ConvocateurNom:       conv.ConvocateurNom,
		ConvocateurPrenom:    conv.ConvocateurPrenom,
		ConvocateurMatricule: conv.ConvocateurMatricule,
		ConvocateurFonction:  conv.ConvocateurFonction,
		AgentsPresents:       conv.AgentsPresents,
		RepresentantParquet:  conv.RepresentantParquet,
		NomParquetier:        conv.NomParquetier,
		ExpertPresent:        conv.ExpertPresent,
		TypeExpert:           conv.TypeExpert,
		InterpreteNecessaire: conv.InterpreteNecessaire,
		LangueInterpretation: conv.LangueInterpretation,
		AvocatPresent:        conv.AvocatPresent,
		NomAvocat:            conv.NomAvocat,
		BarreauAvocat:        conv.BarreauAvocat,

		// SECTION 6: MOTIF ET OBJET
		Motif:                  conv.Motif,
		ObjetPrecis:            conv.ObjetPrecis,
		QuestionsPreparatoires: conv.QuestionsPreparatoires,
		PiecesAApporter:        conv.PiecesAApporter,
		DocumentsDemandes:      conv.DocumentsDemandes,

		// SECTION 9: OBSERVATIONS
		Observations: conv.Observations,

		// SECTION 10: ÉTAT ET TRAÇABILITÉ
		DateEnvoi:        conv.DateEnvoi,
		DateHonoration:   conv.DateHonoration,
		Statut:           StatutConvocation(conv.Statut),
		ResultatAudition: conv.ResultatAudition,
		ModeEnvoi:        conv.ModeEnvoi,

		// Aliases pour compatibilité
		QualiteConvoque: conv.StatutPersonne,
		ConvoqueAdresse: conv.ConvoqueAdresse,
		AffaireLiee:     conv.AffaireLiee,

		// Métadonnées
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}

	// Relations
	if conv.Edges.Agent != nil {
		response.Agent = &AgentSummary{
			ID:        conv.Edges.Agent.ID.String(),
			Nom:       conv.Edges.Agent.Nom,
			Prenom:    conv.Edges.Agent.Prenom,
			Matricule: conv.Edges.Agent.Matricule,
		}
	}

	if conv.Edges.Commissariat != nil {
		response.Commissariat = &CommissariatSummary{
			ID:   conv.Edges.Commissariat.ID.String(),
			Nom:  conv.Edges.Commissariat.Nom,
			Code: conv.Edges.Commissariat.Code,
		}
	}

	// Historique
	if len(conv.Historique) > 0 {
		historique := make([]HistoriqueEntry, len(conv.Historique))
		for i, h := range conv.Historique {
			details := getStringFromMap(h, "details")
			var detailsPtr *string
			if details != "" {
				detailsPtr = &details
			}
			historique[i] = HistoriqueEntry{
				Date:    getStringFromMap(h, "date"),
				DateISO: getStringFromMap(h, "dateISO"),
				Action:  getStringFromMap(h, "action"),
				Agent:   getStringFromMap(h, "agent"),
				Details: detailsPtr,
			}
		}
		response.Historique = historique
	}

	return response
}

// getStringFromMap safely extracts a string value from a map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ReporterRdv reports a convocation to a new date
func (s *service) ReporterRdv(ctx context.Context, id string, req *ReporterRdvRequest, agentID string) (*ConvocationResponse, error) {
	convocID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid convocation ID: %w", err)
	}

	conv, err := s.convocationRepo.GetByID(ctx, convocID.String())
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	// Récupérer les informations de l'agent
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		s.logger.Warn("Failed to get agent info", zap.Error(err))
	}

	agentName := agentID
	if agent != nil {
		agentName = fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
	}

	// Sauvegarder l'ancienne date et heure pour l'historique
	ancienneDateStr := "Non défini"
	ancienneHeureStr := "—"
	if conv.DateRdv != nil && !conv.DateRdv.IsZero() {
		ancienneDateStr = conv.DateRdv.Format("02/01/2006")
	}
	if conv.HeureRdv != nil && *conv.HeureRdv != "" {
		ancienneHeureStr = *conv.HeureRdv
	}

	// Parser la nouvelle date
	nouvelleDate, err := time.Parse("2006-01-02", req.NouvelleDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Mettre à jour la convocation
	updateBuilder := s.convocationRepo.Client().Convocation.
		UpdateOneID(convocID).
		SetDateRdv(nouvelleDate).
		SetHeureRdv(req.NouvelleHeure)

	// Ajouter à l'historique
	var historique []map[string]interface{}
	if len(conv.Historique) > 0 {
		historique = conv.Historique
	}

	// Formater la nouvelle date pour l'affichage
	nouvelleDateFormatee := nouvelleDate.Format("02/01/2006")

	nouvelleEntree := map[string]interface{}{
		"date":    time.Now().Format("02/01/2006 15:04"),
		"dateISO": time.Now().Format(time.RFC3339),
		"action":  "Rendez-vous reporté",
		"agent":   agentName,
		"details": fmt.Sprintf("RDV reporté du %s à %s au %s à %s. Motif: %s",
			ancienneDateStr, ancienneHeureStr, nouvelleDateFormatee, req.NouvelleHeure, req.Motif),
	}

	historique = append(historique, nouvelleEntree)
	updateBuilder.SetHistorique(historique)

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update convocation: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Notifier sends notifications to the convoqué
func (s *service) Notifier(ctx context.Context, id string, req *NotifierRequest, agentID string) (*ConvocationResponse, error) {
	convocID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid convocation ID: %w", err)
	}

	conv, err := s.convocationRepo.GetByID(ctx, convocID.String())
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	// Récupérer les informations de l'agent
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		s.logger.Warn("Failed to get agent info", zap.Error(err))
	}

	agentName := agentID
	if agent != nil {
		agentName = fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
	}

	// Mettre à jour le statut si ce n'est pas déjà envoyé
	updateBuilder := s.convocationRepo.Client().Convocation.
		UpdateOneID(convocID)

	if conv.Statut == "CRÉATION" {
		updateBuilder.SetStatut(convocation.StatutENVOYÉ)
		DateEnvoi := time.Now()
		updateBuilder.SetDateEnvoi(DateEnvoi)
	}

	// Ajouter à l'historique
	var historique []map[string]interface{}
	if len(conv.Historique) > 0 {
		historique = conv.Historique
	}

	moyensStr := strings.Join(req.Moyens, ", ")
	messageDetails := fmt.Sprintf("Notification envoyée par: %s", moyensStr)
	if req.Message != nil && *req.Message != "" {
		messageDetails += fmt.Sprintf(". Message: %s", *req.Message)
	}

	nouvelleEntree := map[string]interface{}{
		"date":    time.Now().Format("02/01/2006 15:04"),
		"dateISO": time.Now().Format(time.RFC3339),
		"action":  "Notification envoyée",
		"agent":   agentName,
		"details": messageDetails,
	}

	historique = append(historique, nouvelleEntree)
	updateBuilder.SetHistorique(historique)

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update convocation: %w", err)
	}

	// TODO: Implémenter l'envoi réel des notifications (SMS, Email, etc.)
	s.logger.Info("Notification sent",
		zap.String("convocation_id", id),
		zap.Strings("moyens", req.Moyens),
	)

	return s.GetByID(ctx, id)
}

// AjouterNote adds a note to a convocation
func (s *service) AjouterNote(ctx context.Context, id string, req *AjouterNoteRequest, agentID string) (*ConvocationResponse, error) {
	convocID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid convocation ID: %w", err)
	}

	conv, err := s.convocationRepo.GetByID(ctx, convocID.String())
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	// Récupérer les informations de l'agent
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		s.logger.Warn("Failed to get agent info", zap.Error(err))
	}

	agentName := agentID
	if agent != nil {
		agentName = fmt.Sprintf("%s %s", agent.Prenom, agent.Nom)
	}

	// Ajouter la note aux observations
	currentObservations := ""
	if conv.Observations != nil {
		currentObservations = *conv.Observations + "\n\n"
	}

	newObservations := currentObservations + fmt.Sprintf("[%s] %s", time.Now().Format("02/01/2006 15:04"), req.Note)

	// Mettre à jour
	updateBuilder := s.convocationRepo.Client().Convocation.
		UpdateOneID(convocID).
		SetObservations(newObservations)

	// Ajouter à l'historique
	var historique []map[string]interface{}
	if len(conv.Historique) > 0 {
		historique = conv.Historique
	}

	nouvelleEntree := map[string]interface{}{
		"date":    time.Now().Format("02/01/2006 15:04"),
		"dateISO": time.Now().Format(time.RFC3339),
		"action":  "Note ajoutée",
		"agent":   agentName,
		"details": req.Note,
	}

	historique = append(historique, nouvelleEntree)
	updateBuilder.SetHistorique(historique)

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	return s.GetByID(ctx, id)
}

// GeneratePDF generates a PDF document for a convocation
func (s *service) GeneratePDF(ctx context.Context, id string) ([]byte, error) {
	conv, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("convocation not found: %w", err)
	}

	// TODO: Implémenter la génération du PDF avec une bibliothèque comme go-pdf
	// Pour l'instant, retourner un PDF minimal
	s.logger.Info("PDF generation requested", zap.String("convocation_id", id))

	// Créer un contenu PDF minimal
	pdfContent := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /Resources 4 0 R /MediaBox [0 0 612 792] /Contents 5 0 R >>
endobj
4 0 obj
<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>
endobj
5 0 obj
<< /Length 200 >>
stream
BT
/F1 24 Tf
100 700 Td
(CONVOCATION) Tj
0 -30 Td
/F1 12 Tf
(Numero: %s) Tj
0 -20 Td
(Nom: %s %s) Tj
0 -20 Td
(Type: %s) Tj
0 -20 Td
(Statut: %s) Tj
ET
endstream
endobj
xref
0 6
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000214 00000 n
0000000304 00000 n
trailer
<< /Size 6 /Root 1 0 R >>
startxref
554
%%%%EOF`,
		conv.Numero,
		conv.ConvoquePrenom,
		conv.ConvoqueNom,
		conv.TypeConvocation,
		conv.Statut,
	)

	return []byte(pdfContent), nil
}
