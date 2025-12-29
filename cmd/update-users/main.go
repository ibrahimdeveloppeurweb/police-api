package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
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
	fmt.Println("🔄 Mise à jour des utilisateurs...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Erreur de chargement de la configuration: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.DBName,
	)

	if cfg.Database.Password != "" {
		dsn += fmt.Sprintf(" password=%s", cfg.Database.Password)
	}

	drv, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		log.Fatalf("❌ Erreur d'ouverture de la connexion: %v", err)
	}
	defer drv.Close()

	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	users := []struct {
		ID            string
		DateNaissance time.Time
		CNI           string
		Adresse       string
		DateEntree    time.Time
		GpsPrecision  float64
		TempsService  string
		EquipeID      string
		SuperieurID   string
	}{
		{
			"user-2",
			time.Date(1975, 8, 20, 0, 0, 0, 0, time.UTC), "CI-1975-020820", "Cocody Riviera 3, Villa 45",
			time.Date(1998, 9, 1, 0, 0, 0, 0, time.UTC), 98.5, "8h15", "", "",
		},
		{
			"user-4",
			time.Date(1982, 3, 12, 0, 0, 0, 0, time.UTC), "CI-1982-120382", "Marcory Zone 4, Immeuble B",
			time.Date(2005, 6, 15, 0, 0, 0, 0, time.UTC), 95.0, "4h30", "", "user-2",
		},
		{
			"user-1",
			time.Date(1988, 5, 15, 0, 0, 0, 0, time.UTC), "CI-1988-150588", "Plateau, Rue du Commerce 23",
			time.Date(2010, 3, 1, 0, 0, 0, 0, time.UTC), 92.5, "6h45", "equipe-1", "user-4",
		},
		{
			"user-3",
			time.Date(1985, 11, 3, 0, 0, 0, 0, time.UTC), "CI-1985-031185", "Cocody Angré, Résidence Soleil",
			time.Date(2008, 1, 15, 0, 0, 0, 0, time.UTC), 88.0, "7h20", "equipe-2", "user-2",
		},
		{
			"user-5",
			time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC), "CI-1990-220790", "Vridi Canal, Bloc 12",
			time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC), 95.5, "5h30", "equipe-3", "user-4",
		},
		{
			"user-6",
			time.Date(1987, 2, 28, 0, 0, 0, 0, time.UTC), "CI-1987-280287", "Adjamé Liberté, Rue 15",
			time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC), 90.0, "7h00", "equipe-1", "user-4",
		},
		{
			"user-7",
			time.Date(1992, 9, 10, 0, 0, 0, 0, time.UTC), "CI-1992-100992", "Yopougon Selmer, Villa 8",
			time.Date(2018, 2, 1, 0, 0, 0, 0, time.UTC), 97.0, "6h15", "equipe-5", "user-4",
		},
		{
			"user-8",
			time.Date(1995, 12, 5, 0, 0, 0, 0, time.UTC), "CI-1995-051295", "Marcory Remblais, Apt 302",
			time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC), 0.0, "0h00", "equipe-4", "user-4",
		},
	}

	for _, u := range users {
		update := client.User.UpdateOneID(getOrCreateID(u.ID)).
			SetDateNaissance(u.DateNaissance).
			SetCni(u.CNI).
			SetAdresse(u.Adresse).
			SetDateEntree(u.DateEntree).
			SetGpsPrecision(u.GpsPrecision).
			SetTempsService(u.TempsService)

		if u.SuperieurID != "" {
			update = update.SetSuperieurID(getOrCreateID(u.SuperieurID))
		}
		if u.EquipeID != "" {
			update = update.SetEquipeID(getOrCreateID(u.EquipeID))
		}

		_, err := update.Save(ctx)
		if err != nil {
			fmt.Printf("⚠️  Erreur mise à jour %s: %v\n", u.ID, err)
		} else {
			fmt.Printf("✅ Utilisateur mis à jour: %s\n", u.ID)
		}
	}

	fmt.Println("\n🎉 Mise à jour terminée!")
}
