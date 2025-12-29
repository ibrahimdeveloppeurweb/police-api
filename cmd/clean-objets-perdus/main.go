package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🧹 Nettoyage de la table objets_perdus...\n")

	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Erreur lors du chargement de la configuration: %v", err)
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
		log.Fatalf("❌ Erreur lors de l'ouverture de la connexion: %v", err)
	}
	defer drv.Close()

	// Créer le client Ent
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	// Compter les objets perdus avant suppression
	countBefore, err := client.ObjetPerdu.Query().Count(ctx)
	if err != nil {
		log.Fatalf("❌ Erreur lors du comptage: %v", err)
	}

	fmt.Printf("📊 Nombre d'objets perdus avant suppression: %d\n", countBefore)

	if countBefore == 0 {
		fmt.Println("✅ La table est déjà vide")
		os.Exit(0)
	}

	// Supprimer tous les objets perdus
	fmt.Println("🗑️  Suppression de tous les objets perdus...")
	deleted, err := client.ObjetPerdu.Delete().Exec(ctx)
	if err != nil {
		log.Fatalf("❌ Erreur lors de la suppression: %v", err)
	}

	fmt.Printf("✅ %d objets perdus supprimés avec succès\n", deleted)

	// Vérifier que la table est vide
	countAfter, err := client.ObjetPerdu.Query().Count(ctx)
	if err != nil {
		log.Fatalf("❌ Erreur lors du comptage final: %v", err)
	}

	if countAfter == 0 {
		fmt.Println("✅ La table objets_perdus a été nettoyée avec succès")
	} else {
		fmt.Printf("⚠️  Il reste %d objets perdus dans la table\n", countAfter)
	}

	fmt.Println("\n🎉 Nettoyage terminé!")
}

