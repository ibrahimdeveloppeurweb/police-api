package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/checkitem"
	"police-trafic-api-frontend-aligned/ent/checkoption"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/inspection"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// idMap stores the mapping between symbolic IDs and actual UUIDs
var idMap = make(map[string]uuid.UUID)

// getOrCreateID gets an existing UUID from the map or creates a new one
func getOrCreateID(symbolicID string) uuid.UUID {
	if id, exists := idMap[symbolicID]; exists {
		return id
	}
	id := uuid.New()
	idMap[symbolicID] = id
	return id
}

func main() {
	fmt.Println("🌱 Insertion des données de test...")

	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Erreur de chargement de la configuration: %v", err)
	}

	// Construire la chaîne de connexion
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.DBName,
	)

	if cfg.Database.Password != "" {
		dsn += fmt.Sprintf(" password=%s", cfg.Database.Password)
	}

	// Ouvrir la connexion
	drv, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		log.Fatalf("❌ Erreur d'ouverture de la connexion: %v", err)
	}
	defer drv.Close()

	// Créer le client Ent
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	// Seed all data in order (dependencies first)
	seedCommissariats(ctx, client)
	seedEquipes(ctx, client)
	seedUsers(ctx, client)
	seedCompetences(ctx, client)
	seedMissions(ctx, client)
	seedObjectifs(ctx, client)
	seedObservations(ctx, client)
	seedInfractionTypes(ctx, client)
	seedConducteurs(ctx, client)
	seedVehicules(ctx, client)
	seedControles(ctx, client)
	seedInfractions(ctx, client)
	seedCheckItems(ctx, client)
	seedCheckOptions(ctx, client)
	seedInspections(ctx, client)

	fmt.Println("\n🎉 Données de test insérées avec succès!")
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Erreur de hashage du mot de passe: %v", err)
	}
	return string(hash)
}

func seedCommissariats(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🏛️ Création des commissariats...")

	commissariats := []struct {
		ID        string
		Nom       string
		Code      string
		Adresse   string
		Ville     string
		Region    string
		Telephone string
		Email     string
		Latitude  float64
		Longitude float64
	}{
		{"comm-1", "Commissariat du 3ème Arrondissement", "ABJ-003", "Boulevard du Général de Gaulle", "Abidjan", "Abidjan", "+225 20 21 30 03", "comm3@police.ci", 5.3364, -4.0266},
		{"comm-2", "Commissariat du 5ème Arrondissement", "ABJ-005", "Avenue Crosson Duplessis", "Abidjan", "Abidjan", "+225 20 21 50 05", "comm5@police.ci", 5.3195, -4.0154},
		{"comm-3", "Commissariat du 7ème Arrondissement", "ABJ-007", "Zone Industrielle de Vridi", "Abidjan", "Abidjan", "+225 20 21 70 07", "comm7@police.ci", 5.2568, -4.0012},
		{"comm-4", "Commissariat du 10ème Arrondissement", "ABJ-010", "Boulevard Lagunaire", "Abidjan", "Abidjan", "+225 20 22 10 10", "comm10@police.ci", 5.3156, -4.0522},
		{"comm-5", "Commissariat du 15ème Arrondissement", "ABJ-015", "Yopougon Zone 4", "Abidjan", "Abidjan", "+225 20 23 15 15", "comm15@police.ci", 5.3489, -4.0856},
	}

	for _, c := range commissariats {
		_, err := client.Commissariat.Create().
			SetID(getOrCreateID(c.ID)).
			SetNom(c.Nom).
			SetCode(c.Code).
			SetAdresse(c.Adresse).
			SetVille(c.Ville).
			SetRegion(c.Region).
			SetTelephone(c.Telephone).
			SetEmail(c.Email).
			SetLatitude(c.Latitude).
			SetLongitude(c.Longitude).
			SetActif(true).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Commissariat %s existe déjà ou erreur: %v\n", c.Nom, err)
		} else {
			fmt.Printf("✅ Commissariat créé: %s\n", c.Nom)
		}
	}
}

func seedUsers(ctx context.Context, client *ent.Client) {
	fmt.Println("\n👤 Création des utilisateurs de test...")

	// Mot de passe par défaut pour tous les utilisateurs de test: "password123"
	defaultPassword := hashPassword("password123")

	users := []struct {
		ID             string
		Matricule      string
		Nom            string
		Prenom         string
		Email          string
		Role           string
		Grade          string
		Telephone      string
		StatutService  string
		Localisation   string
		Activite       string
		CommissariatID string
		// Nouveaux champs
		DateNaissance  time.Time
		CNI            string
		Adresse        string
		DateEntree     time.Time
		GpsPrecision   float64
		TempsService   string
		EquipeID       string
		SuperieurID    string
	}{
		// Commissaire (admin) - pas de supérieur
		{
			"user-2", "67890", "Martin", "Marie", "m.martin@police.ci", "admin", "Commissaire",
			"+225 07 02 34 56", "EN_SERVICE", "Commissariat Central", "Supervision", "comm-2",
			time.Date(1975, 8, 20, 0, 0, 0, 0, time.UTC), "CI-1975-020820", "Cocody Riviera 3, Villa 45",
			time.Date(1998, 9, 1, 0, 0, 0, 0, time.UTC), 98.5, "8h15", "", "",
		},
		// Superviseur - supérieur: commissaire
		{
			"user-4", "22222", "Diallo", "Fatou", "f.diallo@police.ci", "supervisor", "Lieutenant",
			"+225 07 04 56 78", "EN_PAUSE", "Pause déjeuner", "Repos", "comm-3",
			time.Date(1982, 3, 12, 0, 0, 0, 0, time.UTC), "CI-1982-120382", "Marcory Zone 4, Immeuble B",
			time.Date(2005, 6, 15, 0, 0, 0, 0, time.UTC), 95.0, "4h30", "", "user-2",
		},
		// Agents - supérieur: superviseur ou commissaire
		{
			"user-1", "12345", "Dupont", "Jean", "j.dupont@police.ci", "agent", "Sergent",
			"+225 07 01 23 45", "EN_SERVICE", "Boulevard Principal, Plateau", "Patrouille mobile", "comm-1",
			time.Date(1988, 5, 15, 0, 0, 0, 0, time.UTC), "CI-1988-150588", "Plateau, Rue du Commerce 23",
			time.Date(2010, 3, 1, 0, 0, 0, 0, time.UTC), 92.5, "6h45", "equipe-1", "user-4",
		},
		{
			"user-3", "11111", "Koné", "Amadou", "a.kone@police.ci", "agent", "Adjudant",
			"+225 07 03 45 67", "EN_SERVICE", "Avenue Centrale, Cocody", "Contrôle fixe", "comm-2",
			time.Date(1985, 11, 3, 0, 0, 0, 0, time.UTC), "CI-1985-031185", "Cocody Angré, Résidence Soleil",
			time.Date(2008, 1, 15, 0, 0, 0, 0, time.UTC), 88.0, "7h20", "equipe-2", "user-2",
		},
		{
			"user-5", "33333", "Touré", "Ibrahim", "i.toure@police.ci", "agent", "Brigadier",
			"+225 07 05 67 89", "EN_SERVICE", "Zone Industrielle, Vridi", "Investigation", "comm-3",
			time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC), "CI-1990-220790", "Vridi Canal, Bloc 12",
			time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC), 95.5, "5h30", "equipe-3", "user-4",
		},
		{
			"user-6", "44444", "Diabaté", "Moussa", "m.diabate@police.ci", "agent", "Sergent",
			"+225 07 06 78 90", "EN_SERVICE", "Adjamé Marché", "Patrouille mobile", "comm-1",
			time.Date(1987, 2, 28, 0, 0, 0, 0, time.UTC), "CI-1987-280287", "Adjamé Liberté, Rue 15",
			time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC), 90.0, "7h00", "equipe-1", "user-4",
		},
		{
			"user-7", "55555", "Assane", "Fatou", "f.assane@police.ci", "agent", "Adjudant",
			"+225 07 07 89 01", "EN_SERVICE", "Yopougon Zone 4", "Contrôle routier", "comm-5",
			time.Date(1992, 9, 10, 0, 0, 0, 0, time.UTC), "CI-1992-100992", "Yopougon Selmer, Villa 8",
			time.Date(2018, 2, 1, 0, 0, 0, 0, time.UTC), 97.0, "6h15", "equipe-5", "user-4",
		},
		{
			"user-8", "66666", "Yao", "Kofi", "k.yao@police.ci", "agent", "Gardien",
			"+225 07 08 90 12", "HORS_SERVICE", "Position inconnue", "Non joignable", "comm-4",
			time.Date(1995, 12, 5, 0, 0, 0, 0, time.UTC), "CI-1995-051295", "Marcory Remblais, Apt 302",
			time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC), 0.0, "0h00", "equipe-4", "user-4",
		},
	}

	// Première passe: créer tous les utilisateurs sans les relations supérieur/équipe
	for _, u := range users {
		builder := client.User.Create().
			SetID(getOrCreateID(u.ID)).
			SetMatricule(u.Matricule).
			SetNom(u.Nom).
			SetPrenom(u.Prenom).
			SetEmail(u.Email).
			SetPassword(defaultPassword).
			SetRole(u.Role).
			SetGrade(u.Grade).
			SetTelephone(u.Telephone).
			SetStatutService(u.StatutService).
			SetLocalisation(u.Localisation).
			SetActivite(u.Activite).
			SetDerniereActivite(time.Now().Add(-time.Duration(5) * time.Minute)).
			SetActive(true).
			SetDateNaissance(u.DateNaissance).
			SetCni(u.CNI).
			SetAdresse(u.Adresse).
			SetDateEntree(u.DateEntree).
			SetGpsPrecision(u.GpsPrecision).
			SetTempsService(u.TempsService)

		if u.CommissariatID != "" {
			builder = builder.SetCommissariatID(getOrCreateID(u.CommissariatID))
		}

		_, err := builder.Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Utilisateur %s existe déjà ou erreur: %v\n", u.Matricule, err)
		} else {
			fmt.Printf("✅ Utilisateur créé: %s %s (%s) - %s @ %s\n", u.Prenom, u.Nom, u.Role, u.Grade, u.StatutService)
		}
	}

	// Deuxième passe: mettre à jour les relations supérieur et équipe
	fmt.Println("\n🔗 Association des supérieurs et équipes...")
	for _, u := range users {
		if u.SuperieurID != "" || u.EquipeID != "" {
			update := client.User.UpdateOneID(getOrCreateID(u.ID))
			if u.SuperieurID != "" {
				update = update.SetSuperieurID(getOrCreateID(u.SuperieurID))
			}
			if u.EquipeID != "" {
				update = update.SetEquipeID(getOrCreateID(u.EquipeID))
			}
			_, err := update.Save(ctx)
			if err != nil {
				fmt.Printf("⚠️  Erreur mise à jour relations pour %s: %v\n", u.ID, err)
			}
		}
	}

	// Définir les chefs d'équipe
	fmt.Println("\n👑 Définition des chefs d'équipe...")
	chefsEquipe := map[string]string{
		"equipe-1": "user-1", // Dupont Jean chef de l'équipe Alpha
		"equipe-2": "user-3", // Koné Amadou chef de l'équipe Bravo
		"equipe-3": "user-5", // Touré Ibrahim chef de l'équipe Charlie
		"equipe-4": "user-8", // Yao Kofi chef de l'équipe Delta
		"equipe-5": "user-7", // Assane Fatou chef de l'équipe Echo
	}

	for equipeID, chefID := range chefsEquipe {
		_, err := client.Equipe.UpdateOneID(getOrCreateID(equipeID)).
			SetChefEquipeID(getOrCreateID(chefID)).
			Save(ctx)
		if err != nil {
			fmt.Printf("⚠️  Erreur définition chef équipe %s: %v\n", equipeID, err)
		} else {
			fmt.Printf("✅ Chef d'équipe défini pour %s\n", equipeID)
		}
	}
}

func seedInfractionTypes(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🚫 Création des types d'infractions...")

	infractions := []struct {
		ID          string
		Code        string
		Libelle     string
		Description string
		Amende      float64
		Points      int
		Categorie   string
	}{
		// Vitesse
		{"inf-type-1", "EV001", "Excès de vitesse < 20 km/h", "Dépassement de la vitesse autorisée de moins de 20 km/h", 25000.0, 1, "Vitesse"},
		{"inf-type-2", "EV002", "Excès de vitesse 20-30 km/h", "Dépassement de la vitesse autorisée entre 20 et 30 km/h", 50000.0, 2, "Vitesse"},
		{"inf-type-3", "EV003", "Excès de vitesse 30-40 km/h", "Dépassement de la vitesse autorisée entre 30 et 40 km/h", 75000.0, 3, "Vitesse"},
		{"inf-type-4", "EV004", "Excès de vitesse > 40 km/h", "Dépassement de la vitesse autorisée de plus de 40 km/h", 150000.0, 6, "Vitesse"},

		// Signalisation
		{"inf-type-5", "FR001", "Non-respect feu rouge", "Franchissement d'un feu rouge", 50000.0, 4, "Signalisation"},
		{"inf-type-6", "FR002", "Non-respect stop", "Non-respect d'un panneau stop", 35000.0, 4, "Signalisation"},
		{"inf-type-7", "FR003", "Non-respect sens interdit", "Circulation en sens interdit", 50000.0, 4, "Signalisation"},

		// Sécurité
		{"inf-type-8", "TEL001", "Téléphone au volant", "Usage du téléphone en conduisant", 35000.0, 3, "Sécurité"},
		{"inf-type-9", "CE001", "Non-port ceinture", "Non-port de la ceinture de sécurité", 25000.0, 3, "Sécurité"},
		{"inf-type-10", "CQ001", "Non-port casque", "Non-port du casque pour deux-roues", 25000.0, 3, "Sécurité"},

		// Documents
		{"inf-type-11", "AS001", "Défaut d'assurance", "Conduite sans assurance valide", 100000.0, 0, "Documents"},
		{"inf-type-12", "CT001", "Défaut contrôle technique", "Véhicule sans contrôle technique valide", 50000.0, 0, "Documents"},
		{"inf-type-13", "PC001", "Défaut de permis", "Conduite sans permis valide", 150000.0, 0, "Documents"},

		// Stationnement
		{"inf-type-14", "SP001", "Stationnement interdit", "Stationnement sur un emplacement interdit", 15000.0, 0, "Stationnement"},
		{"inf-type-15", "SP002", "Stationnement gênant", "Stationnement gênant la circulation", 25000.0, 0, "Stationnement"},
	}

	for _, inf := range infractions {
		_, err := client.InfractionType.Create().
			SetID(getOrCreateID(inf.ID)).
			SetCode(inf.Code).
			SetLibelle(inf.Libelle).
			SetDescription(inf.Description).
			SetAmende(inf.Amende).
			SetPoints(inf.Points).
			SetCategorie(inf.Categorie).
			SetActive(true).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Type d'infraction %s existe déjà\n", inf.Code)
		} else {
			fmt.Printf("✅ Type d'infraction créé: %s - %s (%.0f FCFA)\n", inf.Code, inf.Libelle, inf.Amende)
		}
	}
}

func seedConducteurs(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🚗 Création des conducteurs de test...")

	conducteurs := []struct {
		ID            string
		Nom           string
		Prenom        string
		DateNaissance time.Time
		Adresse       string
		Ville         string
		Telephone     string
		Email         string
		NumeroPermis  string
		PointsPermis  int
		Nationalite   string
	}{
		{
			"cond-1", "Kouassi", "Yao", time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			"Rue du Commerce 45", "Abidjan", "+225 07 08 09 10 11", "y.kouassi@email.ci",
			"CI-2015-123456", 12, "CI",
		},
		{
			"cond-2", "Bamba", "Aminata", time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC),
			"Boulevard de la Paix 12", "Bouaké", "+225 05 06 07 08 09", "a.bamba@email.ci",
			"CI-2018-234567", 10, "CI",
		},
		{
			"cond-3", "Ouattara", "Moussa", time.Date(1978, 11, 8, 0, 0, 0, 0, time.UTC),
			"Avenue Houphouët-Boigny 78", "Yamoussoukro", "+225 01 02 03 04 05", "m.ouattara@email.ci",
			"CI-2010-345678", 6, "CI",
		},
		{
			"cond-4", "Koné", "Mariam", time.Date(1995, 1, 30, 0, 0, 0, 0, time.UTC),
			"Rue des Jardins 23", "San Pedro", "+225 07 11 22 33 44", "m.kone@email.ci",
			"CI-2020-456789", 12, "CI",
		},
		{
			"cond-5", "Diabaté", "Sekou", time.Date(1982, 5, 17, 0, 0, 0, 0, time.UTC),
			"Avenue de l'Indépendance 56", "Korhogo", "+225 05 55 66 77 88", "s.diabate@email.ci",
			"CI-2012-567890", 8, "CI",
		},
	}

	for _, c := range conducteurs {
		_, err := client.Conducteur.Create().
			SetID(getOrCreateID(c.ID)).
			SetNom(c.Nom).
			SetPrenom(c.Prenom).
			SetDateNaissance(c.DateNaissance).
			SetAdresse(c.Adresse).
			SetVille(c.Ville).
			SetTelephone(c.Telephone).
			SetEmail(c.Email).
			SetNumeroPermis(c.NumeroPermis).
			SetPointsPermis(c.PointsPermis).
			SetNationalite(c.Nationalite).
			SetPermisDelivreLe(c.DateNaissance.AddDate(18, 0, 0)).
			SetPermisValideJusqu(time.Now().AddDate(5, 0, 0)).
			SetActive(true).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Conducteur %s %s existe déjà\n", c.Prenom, c.Nom)
		} else {
			fmt.Printf("✅ Conducteur créé: %s %s (%s)\n", c.Prenom, c.Nom, c.NumeroPermis)
		}
	}
}

func seedVehicules(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🚙 Création des véhicules de test...")

	vehicules := []struct {
		ID                 string
		Immatriculation    string
		Marque             string
		Modele             string
		Couleur            string
		TypeVehicule       string
		ProprietaireNom    string
		ProprietairePrenom string
		AssuranceCompagnie string
		AssuranceNumero    string
	}{
		{"veh-1", "1234AB01", "Toyota", "Corolla", "Blanc", "VP", "Kouassi", "Yao", "NSIA Assurances", "NSIA-2023-001"},
		{"veh-2", "5678CD01", "Peugeot", "308", "Gris", "VP", "Bamba", "Aminata", "Allianz CI", "ALZ-2023-002"},
		{"veh-3", "9012EF01", "Renault", "Clio", "Noir", "VP", "Ouattara", "Moussa", "SUNU Assurances", "SUNU-2023-003"},
		{"veh-4", "3456GH01", "Mercedes", "Sprinter", "Blanc", "VU", "Koné", "Mariam", "AXA Assurances", "AXA-2023-004"},
		{"veh-5", "7890IJ01", "Honda", "CB500", "Rouge", "MOTO", "Diabaté", "Sekou", "NSIA Assurances", "NSIA-2023-005"},
		{"veh-6", "2345KL01", "Hyundai", "Tucson", "Bleu", "SUV", "Touré", "Fatou", "Allianz CI", "ALZ-2023-006"},
		{"veh-7", "6789MN01", "Kia", "Picanto", "Jaune", "VP", "Sanogo", "Ibrahim", "SUNU Assurances", "SUNU-2023-007"},
		{"veh-8", "0123OP01", "Ford", "Transit", "Blanc", "VU", "Coulibaly", "Adama", "AXA Assurances", "AXA-2023-008"},
	}

	for _, v := range vehicules {
		_, err := client.Vehicule.Create().
			SetID(getOrCreateID(v.ID)).
			SetImmatriculation(v.Immatriculation).
			SetMarque(v.Marque).
			SetModele(v.Modele).
			SetCouleur(v.Couleur).
			SetTypeVehicule(v.TypeVehicule).
			SetProprietaireNom(v.ProprietaireNom).
			SetProprietairePrenom(v.ProprietairePrenom).
			SetAssuranceCompagnie(v.AssuranceCompagnie).
			SetAssuranceNumero(v.AssuranceNumero).
			SetActive(true).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Véhicule %s existe déjà\n", v.Immatriculation)
		} else {
			fmt.Printf("✅ Véhicule créé: %s %s (%s)\n", v.Marque, v.Modele, v.Immatriculation)
		}
	}
}

func seedControles(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🚔 Création des contrôles de test...")

	now := time.Now()

	controles := []struct {
		ID              string
		Reference       string
		Location        string
		ControlType     string
		Statut          string
		AgentID         string
		CommissariatID  string
		Notes           string
		VehiclePlate    string
		VehicleMake     string
		VehicleModel    string
		VehicleType     string
		DriverLicense   string
		DriverFirstName string
		DriverLastName  string
		DateOffset      time.Duration
	}{
		// Contrôles d'aujourd'hui
		{
			"ctrl-1", "CTRL-2025-00001", "Boulevard Principal, Plateau", "GENERAL", "TERMINE", "user-1", "comm-1",
			"Contrôle de routine - Excès de vitesse constaté",
			"AB-1234-CI", "Toyota", "Corolla", "VOITURE", "P123456", "Kouamé", "Aya",
			-2 * time.Hour,
		},
		{
			"ctrl-2", "CTRL-2025-00002", "Avenue Centrale, Cocody", "SECURITE", "TERMINE", "user-3", "comm-2",
			"Contrôle alcoolémie négatif - Non-port ceinture",
			"CD-5678-CI", "Peugeot", "308", "VOITURE", "P234567", "Touré", "Ibrahim",
			-4 * time.Hour,
		},
		{
			"ctrl-3", "CTRL-2025-00003", "Adjamé Marché", "DOCUMENT", "TERMINE", "user-6", "comm-1",
			"Documents non conformes - Défaut d'assurance",
			"EF-9012-CI", "Yamaha", "YBR125", "MOTO", "P345678", "Konan", "Serge",
			-1 * time.Hour,
		},
		{
			"ctrl-4", "CTRL-2025-00004", "Zone Industrielle, Vridi", "DOCUMENT", "EN_COURS", "user-5", "comm-3",
			"Vérification documents en cours - Téléphone au volant",
			"GH-3456-CI", "Renault", "Clio", "VOITURE", "P456789", "Bamba", "Fatou",
			-30 * time.Minute,
		},
		{
			"ctrl-5", "CTRL-2025-00005", "Yopougon Zone 4", "GENERAL", "TERMINE", "user-7", "comm-5",
			"Contrôle deux-roues - Non-port casque",
			"IJ-7890-CI", "Honda", "CBR", "MOTO", "P567890", "Diallo", "Mamadou",
			-3 * time.Hour,
		},
		// Contrôles de la semaine
		{
			"ctrl-6", "CTRL-2025-00006", "Boulevard Lagunaire", "MIXTE", "TERMINE", "user-1", "comm-4",
			"Excès de vitesse + téléphone",
			"KL-2345-CI", "Hyundai", "Tucson", "VOITURE", "P678901", "Sylla", "Mariame",
			-48 * time.Hour,
		},
		{
			"ctrl-7", "CTRL-2025-00007", "Plateau Centre", "GENERAL", "CONFORME", "user-3", "comm-2",
			"Contrôle routine - RAS",
			"MN-6789-CI", "Kia", "Sportage", "VOITURE", "P789012", "Camara", "Lamine",
			-72 * time.Hour,
		},
		{
			"ctrl-8", "CTRL-2025-00008", "Treichville Port", "DOCUMENT", "TERMINE", "user-5", "comm-3",
			"Défaut de permis",
			"OP-0123-CI", "Ford", "Transit", "VOITURE", "P890123", "Ouédraogo", "Abdoul",
			-96 * time.Hour,
		},
		{
			"ctrl-9", "CTRL-2025-00009", "Cocody Riviera", "SECURITE", "TERMINE", "user-6", "comm-1",
			"Feu rouge grillé",
			"QR-4567-CI", "Mercedes", "Classe C", "VOITURE", "P901234", "Traoré", "Awa",
			-120 * time.Hour,
		},
		{
			"ctrl-10", "CTRL-2025-00010", "Marcory Zone 4", "GENERAL", "TERMINE", "user-7", "comm-5",
			"Stationnement interdit",
			"ST-8901-CI", "BMW", "Serie 3", "VOITURE", "P012345", "Konaté", "Drissa",
			-144 * time.Hour,
		},
	}

	for _, c := range controles {
		builder := client.Controle.Create().
			SetID(getOrCreateID(c.ID)).
			SetReference(c.Reference).
			SetDateControle(now.Add(c.DateOffset)).
			SetLieuControle(c.Location).
			SetTypeControle(controle.TypeControle(c.ControlType)).
			SetStatut(controle.Statut(c.Statut)).
			SetAgentID(getOrCreateID(c.AgentID)).
			SetObservations(c.Notes).
			SetVehiculeImmatriculation(c.VehiclePlate).
			SetVehiculeMarque(c.VehicleMake).
			SetVehiculeModele(c.VehicleModel).
			SetVehiculeType(controle.VehiculeType(c.VehicleType)).
			SetConducteurNumeroPermis(c.DriverLicense).
			SetConducteurPrenom(c.DriverFirstName).
			SetConducteurNom(c.DriverLastName)

		if c.CommissariatID != "" {
			builder = builder.SetCommissariatID(getOrCreateID(c.CommissariatID))
		}

		_, err := builder.Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Contrôle %s existe déjà ou erreur: %v\n", c.ID, err)
		} else {
			fmt.Printf("✅ Contrôle créé: %s - %s\n", c.Reference, c.Location)
		}
	}
}

func seedEquipes(ctx context.Context, client *ent.Client) {
	fmt.Println("\n👥 Création des équipes...")

	equipes := []struct {
		ID             string
		Nom            string
		Code           string
		Zone           string
		Description    string
		CommissariatID string
	}{
		{"equipe-1", "Équipe Alpha", "EQ-001", "Plateau", "Équipe de patrouille mobile du Plateau", "comm-1"},
		{"equipe-2", "Équipe Bravo", "EQ-002", "Cocody", "Équipe de contrôle routier Cocody", "comm-2"},
		{"equipe-3", "Équipe Charlie", "EQ-003", "Vridi", "Équipe de surveillance zone industrielle", "comm-3"},
		{"equipe-4", "Équipe Delta", "EQ-004", "Marcory", "Équipe d'intervention rapide", "comm-4"},
		{"equipe-5", "Équipe Echo", "EQ-005", "Yopougon", "Équipe de contrôle nocturne", "comm-5"},
	}

	for _, e := range equipes {
		_, err := client.Equipe.Create().
			SetID(getOrCreateID(e.ID)).
			SetNom(e.Nom).
			SetCode(e.Code).
			SetZone(e.Zone).
			SetDescription(e.Description).
			SetCommissariatID(getOrCreateID(e.CommissariatID)).
			SetActive(true).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Équipe %s existe déjà ou erreur: %v\n", e.Nom, err)
		} else {
			fmt.Printf("✅ Équipe créée: %s (%s)\n", e.Nom, e.Code)
		}
	}
}

func seedCompetences(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🎓 Création des compétences...")

	competences := []struct {
		ID             string
		Nom            string
		Type           string
		Description    string
		Organisme      string
		DateObtention  time.Time
		DateExpiration time.Time
	}{
		// Spécialités
		{"comp-1", "Contrôle routier", "SPECIALITE", "Formation aux contrôles routiers et vérification de documents", "École de Police", time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC), time.Time{}},
		{"comp-7", "Investigation", "SPECIALITE", "Techniques d'investigation et enquête", "École de Police", time.Date(2021, 4, 15, 0, 0, 0, 0, time.UTC), time.Time{}},
		{"comp-9", "Régulation trafic", "SPECIALITE", "Gestion de la circulation et régulation du trafic urbain", "École de Police", time.Time{}, time.Time{}},
		{"comp-10", "Accident de circulation", "SPECIALITE", "Constatation et rapport d'accidents de la circulation", "École de Police", time.Time{}, time.Time{}},
		{"comp-11", "Transport de marchandises", "SPECIALITE", "Contrôle des véhicules de transport de marchandises", "École de Police", time.Time{}, time.Time{}},
		{"comp-12", "Transport de personnes", "SPECIALITE", "Contrôle des véhicules de transport de personnes", "École de Police", time.Time{}, time.Time{}},
		{"comp-13", "Véhicules diplomatiques", "SPECIALITE", "Procédures spécifiques pour véhicules diplomatiques", "École de Police", time.Time{}, time.Time{}},

		// Certifications
		{"comp-2", "Alcoolémie", "CERTIFICATION", "Certification utilisation éthylotest homologué", "CNFP", time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
		{"comp-3", "Premiers secours", "CERTIFICATION", "PSC1 - Formation premiers secours civiques", "Croix-Rouge", time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC), time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)},
		{"comp-5", "Conduite moto", "CERTIFICATION", "Permis moto catégorie A", "Auto-école Police", time.Date(2019, 11, 5, 0, 0, 0, 0, time.UTC), time.Date(2029, 11, 5, 0, 0, 0, 0, time.UTC)},
		{"comp-6", "Tir", "CERTIFICATION", "Certification tir de service", "Centre de formation Police", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"comp-8", "Radar mobile", "CERTIFICATION", "Utilisation radar mobile homologué", "CNFP", time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		{"comp-14", "Conduite VIP", "CERTIFICATION", "Conduite de véhicules d'escorte et protection", "Centre de formation Police", time.Time{}, time.Time{}},
		{"comp-15", "Stupéfiants", "CERTIFICATION", "Détection et test de stupéfiants au volant", "CNFP", time.Time{}, time.Time{}},
		{"comp-16", "Chronotachygraphe", "CERTIFICATION", "Lecture et analyse des chronotachygraphes", "CNFP", time.Time{}, time.Time{}},

		// Formations
		{"comp-4", "Gestion de conflits", "FORMATION", "Formation à la médiation et gestion de conflits", "CNFP", time.Date(2023, 9, 10, 0, 0, 0, 0, time.UTC), time.Time{}},
		{"comp-17", "Procédure pénale", "FORMATION", "Formation sur les procédures pénales routières", "École de Police", time.Time{}, time.Time{}},
		{"comp-18", "Droit routier", "FORMATION", "Formation approfondie sur le code de la route", "École de Police", time.Time{}, time.Time{}},
		{"comp-19", "Communication", "FORMATION", "Techniques de communication et relations publiques", "CNFP", time.Time{}, time.Time{}},
		{"comp-20", "Encadrement", "FORMATION", "Formation à l'encadrement d'équipe", "CNFP", time.Time{}, time.Time{}},
		{"comp-21", "Informatique embarquée", "FORMATION", "Utilisation des systèmes informatiques embarqués", "CNFP", time.Time{}, time.Time{}},
		{"comp-22", "Anglais technique", "FORMATION", "Anglais professionnel pour contrôles internationaux", "CNFP", time.Time{}, time.Time{}},
	}

	for _, c := range competences {
		builder := client.Competence.Create().
			SetID(getOrCreateID(c.ID)).
			SetNom(c.Nom).
			SetType(c.Type).
			SetDescription(c.Description).
			SetOrganisme(c.Organisme).
			SetDateObtention(c.DateObtention).
			SetActive(true)

		if !c.DateExpiration.IsZero() {
			builder = builder.SetDateExpiration(c.DateExpiration)
		}

		_, err := builder.Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Compétence %s existe déjà ou erreur: %v\n", c.Nom, err)
		} else {
			fmt.Printf("✅ Compétence créée: %s (%s)\n", c.Nom, c.Type)
		}
	}

	// Associer les compétences aux agents
	fmt.Println("\n🔗 Association des compétences aux agents...")
	associations := []struct {
		UserID       string
		CompetenceID string
	}{
		{"user-1", "comp-1"}, {"user-1", "comp-3"}, {"user-1", "comp-6"},
		{"user-2", "comp-1"}, {"user-2", "comp-4"}, {"user-2", "comp-7"},
		{"user-3", "comp-1"}, {"user-3", "comp-2"}, {"user-3", "comp-8"},
		{"user-4", "comp-1"}, {"user-4", "comp-3"}, {"user-4", "comp-4"}, {"user-4", "comp-7"},
		{"user-5", "comp-1"}, {"user-5", "comp-5"}, {"user-5", "comp-7"},
		{"user-6", "comp-1"}, {"user-6", "comp-2"}, {"user-6", "comp-3"},
		{"user-7", "comp-1"}, {"user-7", "comp-5"}, {"user-7", "comp-8"},
		{"user-8", "comp-1"}, {"user-8", "comp-3"},
	}

	for _, a := range associations {
		_, err := client.User.UpdateOneID(getOrCreateID(a.UserID)).
			AddCompetenceIDs(getOrCreateID(a.CompetenceID)).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Association compétence %s -> %s erreur: %v\n", a.UserID, a.CompetenceID, err)
		}
	}
	fmt.Println("✅ Compétences associées aux agents")
}

func seedMissions(ctx context.Context, client *ent.Client) {
	fmt.Println("\n📋 Création des missions...")

	now := time.Now()

	missions := []struct {
		ID             string
		Type           string
		Titre          string
		DateDebut      time.Time
		DateFin        time.Time
		Duree          string
		Zone           string
		Statut         string
		Rapport        string
		AgentID        string
		CommissariatID string
		EquipeID       string
	}{
		// Missions en cours
		{
			"mission-1", "Patrouille mobile", "Patrouille Plateau Centre", now.Add(-2 * time.Hour), time.Time{},
			"2h00", "Plateau Centre", "EN_COURS", "", "user-1", "comm-1", "equipe-1",
		},
		{
			"mission-2", "Contrôle fixe", "Point de contrôle Cocody", now.Add(-3 * time.Hour), time.Time{},
			"3h00", "Cocody Riviera", "EN_COURS", "", "user-3", "comm-2", "equipe-2",
		},
		// Missions terminées aujourd'hui
		{
			"mission-3", "Contrôle routier", "Contrôle axe Vridi", now.Add(-8 * time.Hour), now.Add(-4 * time.Hour),
			"4h00", "Zone Industrielle Vridi", "TERMINEE", "15 véhicules contrôlés, 3 infractions", "user-5", "comm-3", "equipe-3",
		},
		{
			"mission-4", "Surveillance", "Surveillance nocturne Yopougon", now.Add(-12 * time.Hour), now.Add(-6 * time.Hour),
			"6h00", "Yopougon Zone 4", "TERMINEE", "RAS - Secteur calme", "user-7", "comm-5", "equipe-5",
		},
		// Missions planifiées
		{
			"mission-5", "Patrouille mobile", "Patrouille prévue Marcory", now.Add(24 * time.Hour), time.Time{},
			"4h00", "Marcory Zone 4", "PLANIFIEE", "", "user-6", "comm-4", "equipe-4",
		},
		{
			"mission-6", "Opération spéciale", "Opération alcoolémie weekend", now.Add(48 * time.Hour), time.Time{},
			"8h00", "Plateau - Cocody", "PLANIFIEE", "", "user-1", "comm-1", "equipe-1",
		},
		// Missions passées de la semaine
		{
			"mission-7", "Contrôle routier", "Contrôle documents Adjamé", now.Add(-48 * time.Hour), now.Add(-44 * time.Hour),
			"4h00", "Adjamé Marché", "TERMINEE", "25 contrôles, 5 défauts d'assurance", "user-6", "comm-1", "equipe-1",
		},
		{
			"mission-8", "Investigation", "Enquête accident Treichville", now.Add(-72 * time.Hour), now.Add(-68 * time.Hour),
			"4h00", "Treichville Port", "TERMINEE", "Rapport transmis au procureur", "user-5", "comm-3", "equipe-3",
		},
	}

	for _, m := range missions {
		builder := client.Mission.Create().
			SetID(getOrCreateID(m.ID)).
			SetType(m.Type).
			SetTitre(m.Titre).
			SetDateDebut(m.DateDebut).
			SetDuree(m.Duree).
			SetZone(m.Zone).
			SetStatut(m.Statut).
			AddAgentIDs(getOrCreateID(m.AgentID)).
			SetCommissariatID(getOrCreateID(m.CommissariatID)).
			SetEquipeID(getOrCreateID(m.EquipeID))

		if !m.DateFin.IsZero() {
			builder = builder.SetDateFin(m.DateFin)
		}
		if m.Rapport != "" {
			builder = builder.SetRapport(m.Rapport)
		}

		_, err := builder.Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Mission %s existe déjà ou erreur: %v\n", m.Titre, err)
		} else {
			fmt.Printf("✅ Mission créée: %s (%s)\n", m.Titre, m.Statut)
		}
	}
}

func seedObjectifs(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🎯 Création des objectifs...")

	now := time.Now()
	debutMois := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	finMois := debutMois.AddDate(0, 1, -1)

	objectifs := []struct {
		ID            string
		Titre         string
		Description   string
		Periode       string
		DateDebut     time.Time
		DateFin       time.Time
		Statut        string
		ValeurCible   int
		ValeurActuel  int
		Progression   float64
		AgentID       string
		AssigneParID  string
	}{
		// Objectifs mensuels
		{
			"obj-1", "Contrôles routiers mensuel", "Effectuer 50 contrôles routiers ce mois",
			"mois", debutMois, finMois, "EN_COURS", 50, 35, 70.0, "user-1", "user-2",
		},
		{
			"obj-2", "Infractions constatées", "Constater et verbaliser les infractions",
			"mois", debutMois, finMois, "EN_COURS", 20, 12, 60.0, "user-1", "user-2",
		},
		{
			"obj-3", "Heures de patrouille", "Effectuer 80 heures de patrouille",
			"mois", debutMois, finMois, "EN_COURS", 80, 65, 81.25, "user-3", "user-2",
		},
		{
			"obj-4", "Formation continue", "Suivre 2 formations ce trimestre",
			"trimestre", debutMois, debutMois.AddDate(0, 3, 0), "EN_COURS", 2, 1, 50.0, "user-5", "user-4",
		},
		// Objectifs atteints
		{
			"obj-5", "Contrôles deux-roues", "Focus sur les contrôles deux-roues",
			"mois", debutMois.AddDate(0, -1, 0), debutMois.AddDate(0, 0, -1), "ATTEINT", 30, 34, 100.0, "user-7", "user-4",
		},
		{
			"obj-6", "Taux de conformité", "Maintenir un taux de conformité documents > 90%",
			"mois", debutMois.AddDate(0, -1, 0), debutMois.AddDate(0, 0, -1), "ATTEINT", 90, 92, 100.0, "user-6", "user-2",
		},
		// Objectifs en attente
		{
			"obj-7", "Objectif annuel contrôles", "500 contrôles sur l'année",
			"annee", time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC), time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC),
			"EN_COURS", 500, 380, 76.0, "user-1", "user-2",
		},
		{
			"obj-8", "Réduction accidents zone", "Contribuer à réduire les accidents dans la zone",
			"trimestre", debutMois, debutMois.AddDate(0, 3, 0), "EN_COURS", 100, 75, 75.0, "user-3", "user-4",
		},
	}

	for _, o := range objectifs {
		builder := client.Objectif.Create().
			SetID(getOrCreateID(o.ID)).
			SetTitre(o.Titre).
			SetDescription(o.Description).
			SetPeriode(o.Periode).
			SetDateDebut(o.DateDebut).
			SetDateFin(o.DateFin).
			SetStatut(o.Statut).
			SetValeurCible(o.ValeurCible).
			SetValeurActuelle(o.ValeurActuel).
			SetProgression(o.Progression).
			SetAgentID(getOrCreateID(o.AgentID)).
			SetAssigneParID(getOrCreateID(o.AssigneParID))

		_, err := builder.Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Objectif %s existe déjà ou erreur: %v\n", o.Titre, err)
		} else {
			fmt.Printf("✅ Objectif créé: %s (%.0f%%)\n", o.Titre, o.Progression)
		}
	}
}

func seedObservations(ctx context.Context, client *ent.Client) {
	fmt.Println("\n📝 Création des observations...")

	now := time.Now()

	observations := []struct {
		ID           string
		Contenu      string
		Type         string
		Categorie    string
		VisibleAgent bool
		AgentID      string
		AuteurID     string
		DateOffset   time.Duration
	}{
		// Observations positives
		{
			"obs-1", "Excellent travail lors de l'opération de contrôle du weekend. A fait preuve de professionnalisme et de rigueur.",
			"FELICITATION", "Performance", true, "user-1", "user-2", -24 * time.Hour,
		},
		{
			"obs-2", "Très bonne gestion d'une situation conflictuelle avec un conducteur agressif. Calme et méthodique.",
			"POSITIVE", "Comportement", true, "user-3", "user-4", -48 * time.Hour,
		},
		{
			"obs-3", "A formé efficacement le nouveau collègue sur les procédures de contrôle.",
			"POSITIVE", "Performance", true, "user-5", "user-4", -72 * time.Hour,
		},
		// Observations neutres
		{
			"obs-4", "Point mensuel effectué. Objectifs en bonne voie d'être atteints.",
			"NEUTRE", "Performance", true, "user-1", "user-2", -168 * time.Hour,
		},
		{
			"obs-5", "Entretien de mi-parcours réalisé. Discussion sur les perspectives d'évolution.",
			"NEUTRE", "Performance", true, "user-6", "user-2", -120 * time.Hour,
		},
		// Observations négatives / avertissements
		{
			"obs-6", "Retard de 15 minutes constaté le 15 du mois. Premier avertissement verbal.",
			"AVERTISSEMENT", "Ponctualité", true, "user-8", "user-4", -96 * time.Hour,
		},
		{
			"obs-7", "Tenue non réglementaire observée. Rappel des règles effectué.",
			"NEGATIVE", "Discipline", true, "user-7", "user-4", -144 * time.Hour,
		},
		// Observation confidentielle (non visible par l'agent)
		{
			"obs-8", "Observation pour le dossier: excellent potentiel pour promotion au grade supérieur.",
			"NEUTRE", "Performance", false, "user-1", "user-2", -200 * time.Hour,
		},
	}

	for _, o := range observations {
		_, err := client.Observation.Create().
			SetID(getOrCreateID(o.ID)).
			SetContenu(o.Contenu).
			SetType(o.Type).
			SetCategorie(o.Categorie).
			SetVisibleAgent(o.VisibleAgent).
			SetAgentID(getOrCreateID(o.AgentID)).
			SetAuteurID(getOrCreateID(o.AuteurID)).
			SetCreatedAt(now.Add(o.DateOffset)).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Observation %s existe déjà ou erreur: %v\n", o.ID, err)
		} else {
			fmt.Printf("✅ Observation créée: %s (%s)\n", o.Type, o.Categorie)
		}
	}
}

func seedInfractions(ctx context.Context, client *ent.Client) {
	fmt.Println("\n⚠️ Création des infractions...")

	now := time.Now()

	infractions := []struct {
		ID            string
		NumeroPV      string
		Lieu          string
		Circonstances string
		MontantAmende float64
		PointsRetires int
		Statut        string
		ControleID    string
		TypeInfID     string
		VehiculeID    string
		ConducteurID  string
		DateOffset    time.Duration
	}{
		// Infractions liées aux contrôles d'aujourd'hui
		{"inf-1", "PV-2025-00001", "Boulevard Principal, Plateau", "Excès de vitesse de 25 km/h en zone 50", 50000, 2, "CONSTATEE", "ctrl-1", "inf-type-2", "veh-1", "cond-1", -2 * time.Hour},
		{"inf-2", "PV-2025-00002", "Avenue Centrale, Cocody", "Non-port de la ceinture de sécurité", 25000, 3, "CONSTATEE", "ctrl-2", "inf-type-9", "veh-2", "cond-2", -4 * time.Hour},
		{"inf-3", "PV-2025-00003", "Adjamé Marché", "Défaut d'assurance automobile", 100000, 0, "VALIDEE", "ctrl-3", "inf-type-11", "veh-5", "cond-3", -1 * time.Hour},
		{"inf-4", "PV-2025-00004", "Zone Industrielle, Vridi", "Usage du téléphone en conduisant", 35000, 3, "CONSTATEE", "ctrl-4", "inf-type-8", "veh-3", "cond-4", -30 * time.Minute},
		{"inf-5", "PV-2025-00005", "Yopougon Zone 4", "Non-port du casque", 25000, 3, "PAYEE", "ctrl-5", "inf-type-10", "veh-5", "cond-5", -3 * time.Hour},

		// Infractions de la semaine
		{"inf-6", "PV-2025-00006", "Boulevard Lagunaire", "Excès de vitesse de 35 km/h", 75000, 3, "CONSTATEE", "ctrl-6", "inf-type-3", "veh-6", "cond-1", -48 * time.Hour},
		{"inf-7", "PV-2025-00007", "Boulevard Lagunaire", "Téléphone au volant en plus de l'excès", 35000, 3, "CONSTATEE", "ctrl-6", "inf-type-8", "veh-6", "cond-1", -48 * time.Hour},
		{"inf-8", "PV-2025-00008", "Treichville Port", "Défaut de permis de conduire", 150000, 0, "VALIDEE", "ctrl-8", "inf-type-13", "veh-4", "cond-3", -96 * time.Hour},
		{"inf-9", "PV-2025-00009", "Cocody Riviera", "Non-respect du feu rouge", 50000, 4, "PAYEE", "ctrl-9", "inf-type-5", "veh-7", "cond-2", -120 * time.Hour},
		{"inf-10", "PV-2025-00010", "Marcory Zone 4", "Stationnement interdit", 15000, 0, "PAYEE", "ctrl-10", "inf-type-14", "veh-8", "cond-4", -144 * time.Hour},
	}

	for _, inf := range infractions {
		_, err := client.Infraction.Create().
			SetID(getOrCreateID(inf.ID)).
			SetNumeroPv(inf.NumeroPV).
			SetDateInfraction(now.Add(inf.DateOffset)).
			SetLieuInfraction(inf.Lieu).
			SetCirconstances(inf.Circonstances).
			SetMontantAmende(inf.MontantAmende).
			SetPointsRetires(inf.PointsRetires).
			SetStatut(inf.Statut).
			SetControleID(getOrCreateID(inf.ControleID)).
			SetTypeInfractionID(getOrCreateID(inf.TypeInfID)).
			SetVehiculeID(getOrCreateID(inf.VehiculeID)).
			SetConducteurID(getOrCreateID(inf.ConducteurID)).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  Infraction %s existe déjà ou erreur: %v\n", inf.NumeroPV, err)
		} else {
			fmt.Printf("✅ Infraction créée: %s - %.0f FCFA\n", inf.NumeroPV, inf.MontantAmende)
		}
	}
}

func seedCheckItems(ctx context.Context, client *ent.Client) {
	fmt.Println("\n✅ Création des points de vérification (CheckItems)...")

	items := []struct {
		ID           string
		Name         string
		Code         string
		Category     checkitem.ItemCategory
		ApplicableTo checkitem.ApplicableTo
		Description  string
		Icon         string
		IsMandatory  bool
		DisplayOrder int
		FineAmount   int
		PointsRetrait int
	}{
		// Documents
		{"chk-doc-1", "Permis de conduire", "DOC_PERMIS", checkitem.ItemCategoryDOCUMENT, checkitem.ApplicableToBOTH, "Vérification du permis de conduire valide", "id_card", true, 1, 150000, 0},
		{"chk-doc-2", "Carte grise", "DOC_CARTE_GRISE", checkitem.ItemCategoryDOCUMENT, checkitem.ApplicableToBOTH, "Vérification de la carte grise du véhicule", "description", true, 2, 50000, 0},
		{"chk-doc-3", "Assurance", "DOC_ASSURANCE", checkitem.ItemCategoryDOCUMENT, checkitem.ApplicableToBOTH, "Vérification de l'assurance en cours de validité", "security", true, 3, 100000, 0},
		{"chk-doc-4", "Contrôle technique", "DOC_CONTROLE_TECH", checkitem.ItemCategoryDOCUMENT, checkitem.ApplicableToBOTH, "Vérification du contrôle technique à jour", "build", true, 4, 50000, 0},
		{"chk-doc-5", "Vignette", "DOC_VIGNETTE", checkitem.ItemCategoryDOCUMENT, checkitem.ApplicableToCONTROL, "Vérification de la vignette automobile", "local_offer", false, 5, 25000, 0},

		// Sécurité
		{"chk-sec-1", "Freins", "SAFETY_FREINS", checkitem.ItemCategorySAFETY, checkitem.ApplicableToBOTH, "Vérification du système de freinage", "pan_tool", true, 10, 75000, 3},
		{"chk-sec-2", "Ceintures de sécurité", "SAFETY_CEINTURES", checkitem.ItemCategorySAFETY, checkitem.ApplicableToBOTH, "Vérification des ceintures de sécurité", "airline_seat_legroom_normal", true, 11, 25000, 3},
		{"chk-sec-3", "Pneus", "SAFETY_PNEUS", checkitem.ItemCategorySAFETY, checkitem.ApplicableToBOTH, "Vérification de l'état des pneumatiques", "tire_repair", true, 12, 35000, 2},
		{"chk-sec-4", "Direction", "SAFETY_DIRECTION", checkitem.ItemCategorySAFETY, checkitem.ApplicableToINSPECTION, "Vérification du système de direction", "swap_horiz", true, 13, 50000, 2},
		{"chk-sec-5", "Suspension", "SAFETY_SUSPENSION", checkitem.ItemCategorySAFETY, checkitem.ApplicableToINSPECTION, "Vérification du système de suspension", "height", false, 14, 40000, 0},

		// Éclairage
		{"chk-light-1", "Phares avant", "LIGHT_PHARES", checkitem.ItemCategoryLIGHTING, checkitem.ApplicableToBOTH, "Vérification des phares avant (codes et pleins phares)", "highlight", true, 20, 15000, 1},
		{"chk-light-2", "Feux arrière", "LIGHT_FEUX_AR", checkitem.ItemCategoryLIGHTING, checkitem.ApplicableToBOTH, "Vérification des feux arrière et stops", "wb_twilight", true, 21, 15000, 1},
		{"chk-light-3", "Clignotants", "LIGHT_CLIGNOTANTS", checkitem.ItemCategoryLIGHTING, checkitem.ApplicableToBOTH, "Vérification des clignotants", "turn_right", true, 22, 15000, 1},
		{"chk-light-4", "Feux de détresse", "LIGHT_DETRESSE", checkitem.ItemCategoryLIGHTING, checkitem.ApplicableToBOTH, "Vérification des feux de détresse (warning)", "warning_amber", false, 23, 10000, 0},

		// Équipements
		{"chk-equip-1", "Triangle de signalisation", "EQUIP_TRIANGLE", checkitem.ItemCategoryEQUIPMENT, checkitem.ApplicableToBOTH, "Présence du triangle de signalisation", "change_history", true, 30, 10000, 0},
		{"chk-equip-2", "Gilet de sécurité", "EQUIP_GILET", checkitem.ItemCategoryEQUIPMENT, checkitem.ApplicableToBOTH, "Présence du gilet de sécurité réfléchissant", "checkroom", true, 31, 10000, 0},
		{"chk-equip-3", "Extincteur", "EQUIP_EXTINCTEUR", checkitem.ItemCategoryEQUIPMENT, checkitem.ApplicableToBOTH, "Présence et validité de l'extincteur", "fire_extinguisher", false, 32, 15000, 0},
		{"chk-equip-4", "Trousse de secours", "EQUIP_TROUSSE", checkitem.ItemCategoryEQUIPMENT, checkitem.ApplicableToINSPECTION, "Présence de la trousse de premiers secours", "medical_services", false, 33, 5000, 0},
		{"chk-equip-5", "Roue de secours", "EQUIP_ROUE_SECOURS", checkitem.ItemCategoryEQUIPMENT, checkitem.ApplicableToINSPECTION, "Présence et état de la roue de secours", "donut_large", false, 34, 10000, 0},

		// Visibilité
		{"chk-vis-1", "Pare-brise", "VIS_PAREBRISE", checkitem.ItemCategoryVISIBILITY, checkitem.ApplicableToBOTH, "État du pare-brise (fissures, visibilité)", "window", true, 40, 25000, 0},
		{"chk-vis-2", "Rétroviseurs", "VIS_RETROVISEURS", checkitem.ItemCategoryVISIBILITY, checkitem.ApplicableToBOTH, "État et présence des rétroviseurs", "preview", true, 41, 15000, 1},
		{"chk-vis-3", "Essuie-glaces", "VIS_ESSUIEGLACES", checkitem.ItemCategoryVISIBILITY, checkitem.ApplicableToINSPECTION, "Fonctionnement des essuie-glaces", "water_drop", false, 42, 10000, 0},
		{"chk-vis-4", "Vitres teintées", "VIS_VITRES_TEINTEES", checkitem.ItemCategoryVISIBILITY, checkitem.ApplicableToCONTROL, "Conformité du niveau de teinte des vitres", "blur_on", false, 43, 25000, 0},
	}

	for _, item := range items {
		_, err := client.CheckItem.Create().
			SetID(getOrCreateID(item.ID)).
			SetItemName(item.Name).
			SetItemCode(item.Code).
			SetItemCategory(item.Category).
			SetApplicableTo(item.ApplicableTo).
			SetDescription(item.Description).
			SetIcon(item.Icon).
			SetIsMandatory(item.IsMandatory).
			SetIsActive(true).
			SetDisplayOrder(item.DisplayOrder).
			SetFineAmount(item.FineAmount).
			SetPointsRetrait(item.PointsRetrait).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  CheckItem %s existe déjà ou erreur: %v\n", item.Code, err)
		} else {
			fmt.Printf("✅ CheckItem créé: %s - %s\n", item.Code, item.Name)
		}
	}
}

func seedCheckOptions(ctx context.Context, client *ent.Client) {
	fmt.Println("\n📋 Création des résultats de vérifications (CheckOptions)...")

	now := time.Now()

	// Structure pour les résultats
	type checkResult struct {
		ID          string
		SourceType  checkoption.SourceType
		SourceID    string
		CheckItemID string
		Status      checkoption.ResultStatus
		Notes       string
		FineAmount  int
	}

	options := []checkResult{
		// Contrôle ctrl-1 (TERMINE - Excès de vitesse)
		{"chkopt-ctrl1-1", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-doc-1", checkoption.ResultStatusPASS, "Permis valide jusqu'en 2027", 0},
		{"chkopt-ctrl1-2", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise conforme", 0},
		{"chkopt-ctrl1-3", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-doc-3", checkoption.ResultStatusPASS, "Assurance NSIA valide", 0},
		{"chkopt-ctrl1-4", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-doc-4", checkoption.ResultStatusPASS, "Contrôle technique à jour", 0},
		{"chkopt-ctrl1-5", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-sec-1", checkoption.ResultStatusPASS, "Freins en bon état", 0},
		{"chkopt-ctrl1-6", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-sec-2", checkoption.ResultStatusPASS, "Ceintures fonctionnelles", 0},
		{"chkopt-ctrl1-7", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-light-1", checkoption.ResultStatusPASS, "Phares OK", 0},
		{"chkopt-ctrl1-8", checkoption.SourceTypeCONTROL, "ctrl-1", "chk-equip-1", checkoption.ResultStatusPASS, "Triangle présent", 0},

		// Contrôle ctrl-2 (TERMINE - Non-port ceinture)
		{"chkopt-ctrl2-1", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-doc-1", checkoption.ResultStatusPASS, "Permis valide", 0},
		{"chkopt-ctrl2-2", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise OK", 0},
		{"chkopt-ctrl2-3", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-doc-3", checkoption.ResultStatusPASS, "Assurance valide", 0},
		{"chkopt-ctrl2-4", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-doc-4", checkoption.ResultStatusWARNING, "Contrôle technique expire dans 15 jours", 0},
		{"chkopt-ctrl2-5", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-sec-2", checkoption.ResultStatusFAIL, "Conducteur non attaché lors du contrôle", 25000},
		{"chkopt-ctrl2-6", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-light-1", checkoption.ResultStatusPASS, "Éclairage OK", 0},
		{"chkopt-ctrl2-7", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-equip-1", checkoption.ResultStatusPASS, "Triangle présent", 0},
		{"chkopt-ctrl2-8", checkoption.SourceTypeCONTROL, "ctrl-2", "chk-equip-2", checkoption.ResultStatusPASS, "Gilet présent", 0},

		// Contrôle ctrl-3 (TERMINE - Défaut assurance)
		{"chkopt-ctrl3-1", checkoption.SourceTypeCONTROL, "ctrl-3", "chk-doc-1", checkoption.ResultStatusPASS, "Permis moto valide", 0},
		{"chkopt-ctrl3-2", checkoption.SourceTypeCONTROL, "ctrl-3", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise conforme", 0},
		{"chkopt-ctrl3-3", checkoption.SourceTypeCONTROL, "ctrl-3", "chk-doc-3", checkoption.ResultStatusFAIL, "Assurance expirée depuis 2 mois", 100000},
		{"chkopt-ctrl3-4", checkoption.SourceTypeCONTROL, "ctrl-3", "chk-light-1", checkoption.ResultStatusPASS, "Phare OK", 0},
		{"chkopt-ctrl3-5", checkoption.SourceTypeCONTROL, "ctrl-3", "chk-light-2", checkoption.ResultStatusWARNING, "Feu arrière faible", 0},

		// Contrôle ctrl-4 (EN_COURS - Téléphone au volant)
		{"chkopt-ctrl4-1", checkoption.SourceTypeCONTROL, "ctrl-4", "chk-doc-1", checkoption.ResultStatusPASS, "Permis vérifié", 0},
		{"chkopt-ctrl4-2", checkoption.SourceTypeCONTROL, "ctrl-4", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise OK", 0},
		{"chkopt-ctrl4-3", checkoption.SourceTypeCONTROL, "ctrl-4", "chk-doc-3", checkoption.ResultStatusPASS, "Assurance valide", 0},
		{"chkopt-ctrl4-4", checkoption.SourceTypeCONTROL, "ctrl-4", "chk-doc-4", checkoption.ResultStatusNOT_CHECKED, "", 0},

		// Contrôle ctrl-5 (TERMINE - Non-port casque moto)
		{"chkopt-ctrl5-1", checkoption.SourceTypeCONTROL, "ctrl-5", "chk-doc-1", checkoption.ResultStatusPASS, "Permis A valide", 0},
		{"chkopt-ctrl5-2", checkoption.SourceTypeCONTROL, "ctrl-5", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise OK", 0},
		{"chkopt-ctrl5-3", checkoption.SourceTypeCONTROL, "ctrl-5", "chk-doc-3", checkoption.ResultStatusPASS, "Assurance moto valide", 0},
		{"chkopt-ctrl5-4", checkoption.SourceTypeCONTROL, "ctrl-5", "chk-light-1", checkoption.ResultStatusPASS, "Phare fonctionnel", 0},
		{"chkopt-ctrl5-5", checkoption.SourceTypeCONTROL, "ctrl-5", "chk-sec-3", checkoption.ResultStatusWARNING, "Pneu arrière usé - à surveiller", 0},

		// Contrôle ctrl-7 (CONFORME - RAS)
		{"chkopt-ctrl7-1", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-doc-1", checkoption.ResultStatusPASS, "Permis en règle", 0},
		{"chkopt-ctrl7-2", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-doc-2", checkoption.ResultStatusPASS, "Carte grise conforme", 0},
		{"chkopt-ctrl7-3", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-doc-3", checkoption.ResultStatusPASS, "Assurance valide 1 an", 0},
		{"chkopt-ctrl7-4", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-doc-4", checkoption.ResultStatusPASS, "Contrôle technique récent", 0},
		{"chkopt-ctrl7-5", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-sec-1", checkoption.ResultStatusPASS, "Freins parfaits", 0},
		{"chkopt-ctrl7-6", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-sec-2", checkoption.ResultStatusPASS, "Ceintures OK", 0},
		{"chkopt-ctrl7-7", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-sec-3", checkoption.ResultStatusPASS, "Pneus neufs", 0},
		{"chkopt-ctrl7-8", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-light-1", checkoption.ResultStatusPASS, "Éclairage complet OK", 0},
		{"chkopt-ctrl7-9", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-light-2", checkoption.ResultStatusPASS, "Feux arrière OK", 0},
		{"chkopt-ctrl7-10", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-equip-1", checkoption.ResultStatusPASS, "Triangle présent", 0},
		{"chkopt-ctrl7-11", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-equip-2", checkoption.ResultStatusPASS, "Gilet présent", 0},
		{"chkopt-ctrl7-12", checkoption.SourceTypeCONTROL, "ctrl-7", "chk-vis-1", checkoption.ResultStatusPASS, "Pare-brise impeccable", 0},
	}

	for _, opt := range options {
		_, err := client.CheckOption.Create().
			SetID(getOrCreateID(opt.ID)).
			SetSourceType(opt.SourceType).
			SetSourceID(getOrCreateID(opt.SourceID).String()).
			SetCheckItemID(getOrCreateID(opt.CheckItemID)).
			SetResultStatus(opt.Status).
			SetNotes(opt.Notes).
			SetFineAmount(opt.FineAmount).
			SetCheckedAt(now).
			Save(ctx)

		if err != nil {
			fmt.Printf("⚠️  CheckOption %s existe déjà ou erreur: %v\n", opt.ID, err)
		} else {
			statusEmoji := "✅"
			if opt.Status == checkoption.ResultStatusFAIL {
				statusEmoji = "❌"
			} else if opt.Status == checkoption.ResultStatusWARNING {
				statusEmoji = "⚠️"
			} else if opt.Status == checkoption.ResultStatusNOT_CHECKED {
				statusEmoji = "⏸️"
			}
			fmt.Printf("%s CheckOption créé: %s (%s)\n", statusEmoji, opt.ID, opt.Status)
		}
	}
}

func seedInspections(ctx context.Context, client *ent.Client) {
	fmt.Println("\n🔍 Création des inspections...")

	now := time.Now()

	inspections := []struct {
		ID                       string
		Numero                   string
		Statut                   inspection.Statut
		Observations             string
		DateInspection           time.Time
		TotalVerifications       int
		VerificationsOk          int
		VerificationsAttention   int
		VerificationsEchec       int
		MontantTotalAmendes      int
		VehiculeImmatriculation  string
		VehiculeMarque           string
		VehiculeModele           string
		VehiculeAnnee            int
		VehiculeCouleur          string
		VehiculeNumeroChassis    string
		VehiculeType             inspection.VehiculeType
		ConducteurNumeroPermis   string
		ConducteurPrenom         string
		ConducteurNom            string
		ConducteurTelephone      string
		ConducteurAdresse        string
		AssuranceCompagnie       string
		AssuranceNumeroPolice    string
		AssuranceDateExpiration  time.Time
		AssuranceStatut          inspection.AssuranceStatut
		LieuInspection           string
		InspecteurID             string
		CommissariatID           string
	}{
		{
			"insp-1", "INS-2025-00001", inspection.StatutCONFORME, "Véhicule en parfait état, tous les documents en règle",
			now.AddDate(0, 0, -2), 10, 10, 0, 0, 0,
			"AB-123-CI", "Toyota", "Corolla", 2020, "Blanc", "JTDKN3DU5A0123456", inspection.VehiculeTypeVOITURE,
			"CI-2019-001234", "Kouassi", "Yao", "+225 07 01 02 03", "Cocody, Abidjan",
			"NSIA Assurance", "POL-2024-001234", now.AddDate(1, 0, 0), inspection.AssuranceStatutACTIVE,
			"Boulevard Lagunaire, Plateau", "user-1", "comm-1",
		},
		{
			"insp-2", "INS-2025-00002", inspection.StatutNON_CONFORME, "Assurance expirée, pneus usés, feux arrière défaillants",
			now.AddDate(0, 0, -5), 10, 5, 2, 3, 75000,
			"CD-456-CI", "Peugeot", "308", 2018, "Noir", "VF3LCYHZPJS123456", inspection.VehiculeTypeVOITURE,
			"CI-2017-005678", "Aminata", "Koné", "+225 07 04 05 06", "Marcory, Abidjan",
			"Saham Assurance", "POL-2023-005678", now.AddDate(0, -3, 0), inspection.AssuranceStatutEXPIREE,
			"Avenue Christiani, Adjamé", "user-3", "comm-2",
		},
		{
			"insp-3", "INS-2025-00003", inspection.StatutEN_COURS, "Inspection en cours, vérification des documents",
			now.AddDate(0, 0, -1), 10, 6, 0, 0, 0,
			"EF-789-CI", "Honda", "Civic", 2022, "Gris", "SHHFK2F40EU123456", inspection.VehiculeTypeVOITURE,
			"CI-2020-009012", "Mamadou", "Traoré", "+225 07 07 08 09", "Yopougon, Abidjan",
			"Allianz CI", "POL-2024-009012", now.AddDate(0, 8, 0), inspection.AssuranceStatutACTIVE,
			"Zone Industrielle, Vridi", "user-5", "comm-3",
		},
		{
			"insp-4", "INS-2025-00004", inspection.StatutTERMINE, "Inspection terminée avec avertissements mineurs",
			now.AddDate(0, 0, -10), 10, 8, 2, 0, 0,
			"GH-012-CI", "Renault", "Clio", 2019, "Rouge", "VF1RJA00461234567", inspection.VehiculeTypeVOITURE,
			"CI-2018-003456", "Fatou", "Diallo", "+225 07 10 11 12", "Treichville, Abidjan",
			"AXA Assurance", "POL-2024-003456", now.AddDate(0, 5, 0), inspection.AssuranceStatutACTIVE,
			"Boulevard de Marseille, Marcory", "user-1", "comm-1",
		},
		{
			"insp-5", "INS-2025-00005", inspection.StatutCONFORME, "Moto en bon état, tous les documents valides",
			now.AddDate(0, 0, -3), 8, 8, 0, 0, 0,
			"IJ-345-CI", "Yamaha", "MT-07", 2021, "Bleu", "JYARN23E0LA123456", inspection.VehiculeTypeMOTO,
			"CI-2019-007890", "Ibrahim", "Ouattara", "+225 07 13 14 15", "Cocody Angré, Abidjan",
			"SUNU Assurance", "POL-2024-007890", now.AddDate(0, 10, 0), inspection.AssuranceStatutACTIVE,
			"Avenue Franchet d'Espèrey, Cocody", "user-3", "comm-2",
		},
		{
			"insp-6", "INS-2025-00006", inspection.StatutEN_ATTENTE, "En attente de début d'inspection",
			now, 10, 0, 0, 0, 0,
			"KL-678-CI", "Mercedes", "Classe C", 2023, "Argent", "WDD2050021R123456", inspection.VehiculeTypeVOITURE,
			"CI-2021-001122", "Sekou", "Bamba", "+225 07 16 17 18", "Riviera 3, Abidjan",
			"NSIA Assurance", "POL-2024-001122", now.AddDate(1, 2, 0), inspection.AssuranceStatutACTIVE,
			"Rue des Jardins, Riviera 3", "user-5", "comm-3",
		},
		{
			"insp-7", "INS-2025-00007", inspection.StatutNON_CONFORME, "Défaillances multiples: freins, éclairage, assurance",
			now.AddDate(0, 0, -7), 10, 3, 1, 6, 150000,
			"MN-901-CI", "Nissan", "Altima", 2017, "Blanc", "1N4AL3AP6HC123456", inspection.VehiculeTypeVOITURE,
			"CI-2016-004455", "Aya", "Coulibaly", "+225 07 19 20 21", "Abobo, Abidjan",
			"Saham Assurance", "POL-2022-004455", now.AddDate(0, -6, 0), inspection.AssuranceStatutEXPIREE,
			"Commune d'Abobo", "user-1", "comm-1",
		},
		{
			"insp-8", "INS-2025-00008", inspection.StatutCONFORME, "Camion professionnel, tous documents à jour",
			now.AddDate(0, 0, -4), 12, 12, 0, 0, 0,
			"OP-234-CI", "Mercedes", "Actros", 2020, "Blanc", "WDB96340310123456", inspection.VehiculeTypeCAMION,
			"CI-2018-008877", "Moussa", "Sanogo", "+225 07 22 23 24", "Port Bouët, Abidjan",
			"AXA Assurance", "POL-2024-008877", now.AddDate(0, 9, 0), inspection.AssuranceStatutACTIVE,
			"Zone Portuaire, Port Bouët", "user-3", "comm-2",
		},
		{
			"insp-9", "INS-2025-00009", inspection.StatutTERMINE, "Bus scolaire inspecté, quelques remarques mineures",
			now.AddDate(0, 0, -15), 15, 13, 2, 0, 0,
			"QR-567-CI", "Mercedes", "Sprinter", 2019, "Jaune", "WDB9066571P123456", inspection.VehiculeTypeBUS,
			"CI-2017-006655", "Adjoua", "Fofana", "+225 07 25 26 27", "Bingerville",
			"SUNU Assurance", "POL-2024-006655", now.AddDate(0, 6, 0), inspection.AssuranceStatutACTIVE,
			"Route de Bingerville", "user-5", "comm-3",
		},
		{
			"insp-10", "INS-2025-00010", inspection.StatutEN_COURS, "Inspection de camionnette de livraison",
			now.AddDate(0, 0, 0), 10, 4, 0, 0, 0,
			"ST-890-CI", "Ford", "Transit", 2021, "Blanc", "WF0XXXGCDX1Y12345", inspection.VehiculeTypeCAMIONNETTE,
			"CI-2020-002233", "Koffi", "Dosso", "+225 07 28 29 30", "Plateau, Abidjan",
			"Allianz CI", "POL-2024-002233", now.AddDate(0, 11, 0), inspection.AssuranceStatutACTIVE,
			"Boulevard du Général de Gaulle, Plateau", "user-1", "comm-1",
		},
	}

	for _, insp := range inspections {
		builder := client.Inspection.Create().
			SetID(getOrCreateID(insp.ID)).
			SetNumero(insp.Numero).
			SetStatut(insp.Statut).
			SetObservations(insp.Observations).
			SetDateInspection(insp.DateInspection).
			SetTotalVerifications(insp.TotalVerifications).
			SetVerificationsOk(insp.VerificationsOk).
			SetVerificationsAttention(insp.VerificationsAttention).
			SetVerificationsEchec(insp.VerificationsEchec).
			SetMontantTotalAmendes(insp.MontantTotalAmendes).
			SetVehiculeImmatriculation(insp.VehiculeImmatriculation).
			SetVehiculeMarque(insp.VehiculeMarque).
			SetVehiculeModele(insp.VehiculeModele).
			SetVehiculeAnnee(insp.VehiculeAnnee).
			SetVehiculeCouleur(insp.VehiculeCouleur).
			SetVehiculeNumeroChassis(insp.VehiculeNumeroChassis).
			SetVehiculeType(insp.VehiculeType).
			SetConducteurNumeroPermis(insp.ConducteurNumeroPermis).
			SetConducteurPrenom(insp.ConducteurPrenom).
			SetConducteurNom(insp.ConducteurNom).
			SetConducteurTelephone(insp.ConducteurTelephone).
			SetConducteurAdresse(insp.ConducteurAdresse).
			SetAssuranceCompagnie(insp.AssuranceCompagnie).
			SetAssuranceNumeroPolice(insp.AssuranceNumeroPolice).
			SetAssuranceDateExpiration(insp.AssuranceDateExpiration).
			SetAssuranceStatut(insp.AssuranceStatut).
			SetLieuInspection(insp.LieuInspection).
			SetInspecteurID(getOrCreateID(insp.InspecteurID)).
			SetCommissariatID(getOrCreateID(insp.CommissariatID))

		_, err := builder.Save(ctx)
		if err != nil {
			fmt.Printf("⚠️  Inspection %s existe déjà ou erreur: %v\n", insp.Numero, err)
		} else {
			statutEmoji := "📋"
			switch insp.Statut {
			case inspection.StatutCONFORME:
				statutEmoji = "✅"
			case inspection.StatutNON_CONFORME:
				statutEmoji = "❌"
			case inspection.StatutEN_COURS:
				statutEmoji = "🔄"
			case inspection.StatutEN_ATTENTE:
				statutEmoji = "⏳"
			case inspection.StatutTERMINE:
				statutEmoji = "✔️"
			}
			fmt.Printf("%s Inspection créée: %s - %s (%s)\n", statutEmoji, insp.Numero, insp.VehiculeImmatriculation, insp.Statut)
		}
	}
}
