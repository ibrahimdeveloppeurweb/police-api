package main

import (
	"context"
	"fmt"
	"log"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/migrate"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔄 Démarrage des migrations...")

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

	// Ouvrir la connexion
	drv, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		log.Fatalf("❌ Erreur d'ouverture de la connexion: %v", err)
	}
	defer drv.Close()

	// Créer le client Ent
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	// Exécuter les migrations avec options de mise à jour du schéma
	ctx := context.Background()
	fmt.Println("📦 Exécution des migrations...")

	// Utiliser DropColumn et DropIndex pour mettre à jour le schéma existant
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	); err != nil {
		log.Fatalf("❌ Erreur lors de la migration: %v", err)
	}

	fmt.Println("✅ Migrations exécutées avec succès!")
	
	// Afficher les tables créées
	fmt.Println("\n📋 Tables créées:")
	tables := []string{
		"users",
		"infraction_types", 
		"controles",
		// Ajouter d'autres tables quand elles seront créées
	}
	
	for _, table := range tables {
		fmt.Printf("   ✓ %s\n", table)
	}
	
	fmt.Println("\n🎉 Migration terminée!")
}