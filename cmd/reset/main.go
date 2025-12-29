package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🗑️  Réinitialisation de la base de données...")

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

	// Ajouter le mot de passe si fourni
	if cfg.Database.Password != "" {
		dsn += fmt.Sprintf(" password=%s", cfg.Database.Password)
	}

	fmt.Printf("📡 Connexion à la base de données: %s:%d/%s\n",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Ouvrir la connexion directe
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Erreur d'ouverture de la connexion: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Tables à supprimer dans l'ordre (pour respecter les contraintes de clé étrangère)
	tables := []string{
		// Many-to-many junction tables first
		"user_competences",
		"mission_agents",
		"competence_agents",
		// Then child tables
		"check_options",
		"check_items",
		"infractions",
		"paiements",
		"recours",
		"proces_verbals",
		"inspections",
		"controles",
		"documents",
		"plaintes",
		"alerte_securitaires",
		"audit_logs",
		"observations",
		"objectifs",
		"missions",
		"equipes",
		"competences",
		"vehicules",
		"conducteurs",
		"infraction_types",
		"users",
		"commissariats",
	}

	fmt.Println("\n📦 Suppression des tables...")
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			fmt.Printf("   ⚠️  Erreur pour %s: %v\n", table, err)
		} else {
			fmt.Printf("   ✓ %s supprimée\n", table)
		}
	}

	fmt.Println("\n✅ Base de données réinitialisée!")
	fmt.Println("Exécutez maintenant: go run cmd/migrate/main.go && go run cmd/seed/main.go")
}
