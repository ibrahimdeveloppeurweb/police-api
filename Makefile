# Makefile pour Police Traffic API

.PHONY: help build run test clean db-setup db-migrate db-seed db-reset docker-up docker-down deps generate swagger

# Configuration
APP_NAME := police-traffic-api
GO_MODULE := police-trafic-api-frontend-aligned

# Couleurs pour l'affichage
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

help: ## Afficher cette aide
	@echo "$(CYAN)Police Traffic API - Commandes disponibles:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# Application
run: ## Lancer le serveur de développement
	@echo "$(CYAN)🚀 Démarrage du serveur...$(RESET)"
	@go run ./cmd/server

build: ## Compiler l'application
	@echo "$(CYAN)🔨 Compilation de l'application...$(RESET)"
	@go build -v -o bin/server ./cmd/server
	@go build -v -o bin/migrate ./cmd/migrate  
	@go build -v -o bin/seed ./cmd/seed
	@echo "$(GREEN)✅ Compilation terminée$(RESET)"

test: ## Exécuter les tests
	@echo "$(CYAN)🧪 Exécution des tests...$(RESET)"
	@go test -v ./...

clean: ## Nettoyer les fichiers de build
	@echo "$(CYAN)🧹 Nettoyage...$(RESET)"
	@rm -rf bin/
	@rm -f main server
	@go clean -cache
	@echo "$(GREEN)✅ Nettoyage terminé$(RESET)"

deps: ## Installer/mettre à jour les dépendances
	@echo "$(CYAN)📦 Installation des dépendances...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dépendances mises à jour$(RESET)"

# Base de données
db-setup: ## Configuration complète de la base de données
	@echo "$(CYAN)🗄️  Configuration de la base de données...$(RESET)"
	@./scripts/database/setup_complete.sh

db-migrate: ## Exécuter les migrations uniquement
	@echo "$(CYAN)📦 Exécution des migrations...$(RESET)"
	@go run ./cmd/migrate

db-seed: ## Insérer les données de test uniquement
	@echo "$(CYAN)🌱 Insertion des données de test...$(RESET)"
	@go run ./cmd/seed

db-reset: ## Supprimer et recréer la base de données
	@echo "$(YELLOW)⚠️  Suppression de la base de données...$(RESET)"
	@psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS police_traffic;" || true
	@psql -h localhost -U postgres -c "CREATE DATABASE police_traffic;"
	@echo "$(GREEN)✅ Base de données réinitialisée$(RESET)"
	@$(MAKE) db-migrate
	@$(MAKE) db-seed

# Ent
generate: ## Régénérer les entités Ent
	@echo "$(CYAN)⚡ Génération des entités Ent...$(RESET)"
	@go generate ./ent
	@echo "$(GREEN)✅ Entités générées$(RESET)"

ent-new: ## Créer un nouveau schéma Ent (usage: make ent-new SCHEMA=MyEntity)
	@echo "$(CYAN)📝 Création du schéma $(SCHEMA)...$(RESET)"
	@go run entgo.io/ent/cmd/ent new $(SCHEMA)

# Docker (optionnel)
docker-up: ## Démarrer PostgreSQL avec Docker
	@echo "$(CYAN)🐳 Démarrage de PostgreSQL avec Docker...$(RESET)"
	@docker run --name postgres-police \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=police_traffic \
		-p 5432:5432 \
		-d postgres:15
	@echo "$(GREEN)✅ PostgreSQL démarré sur localhost:5432$(RESET)"

docker-down: ## Arrêter PostgreSQL Docker
	@echo "$(CYAN)🐳 Arrêt de PostgreSQL...$(RESET)"
	@docker stop postgres-police || true
	@docker rm postgres-police || true

# Swagger
install-swag: ## Installer swag pour Swagger
	@echo "$(CYAN)📚 Installation de swag...$(RESET)"
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✅ Swag installé$(RESET)"

swagger: ## Générer la documentation Swagger
	@echo "$(CYAN)📚 Génération de la documentation Swagger...$(RESET)"
	@swag init -g cmd/server/main.go -o docs
	@echo "$(GREEN)✅ Documentation générée$(RESET)"

# Développement
dev: db-setup ## Configuration complète pour développement
	@echo "$(GREEN)🎉 Environnement de développement prêt!$(RESET)"
	@echo "$(CYAN)Lancer le serveur avec: make run$(RESET)"

lint: ## Vérifier le code avec golangci-lint
	@echo "$(CYAN)🔍 Vérification du code...$(RESET)"
	@golangci-lint run || echo "$(YELLOW)⚠️  golangci-lint n'est pas installé$(RESET)"

fmt: ## Formatter le code
	@echo "$(CYAN)💅 Formatage du code...$(RESET)"
	@go fmt ./...
	@go mod tidy

# Informations
info: ## Afficher les informations du projet
	@echo "$(CYAN)📋 Informations du projet:$(RESET)"
	@echo "  Nom: $(APP_NAME)"
	@echo "  Module: $(GO_MODULE)"
	@echo "  Version Go: $(shell go version)"
	@echo "  Répertoire: $(PWD)"
	@echo ""
	@echo "$(CYAN)🗄️  Base de données:$(RESET)"
	@echo "  Host: localhost:5432"
	@echo "  Database: police_traffic"
	@echo "  User: postgres"
	@echo ""
	@echo "$(CYAN)🚀 Endpoints:$(RESET)"
	@echo "  Health: http://localhost:8080/health"
	@echo "  API: http://localhost:8080/api/*"
	@echo "  Swagger: http://localhost:8080/swagger/index.html"


