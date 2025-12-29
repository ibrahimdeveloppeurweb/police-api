#!/bin/bash

# Script d'initialisation de la base de données PostgreSQL
# Usage: ./scripts/database/init_db.sh

set -e

echo "🗄️  Initialisation de la base de données PostgreSQL..."

# Vérifier si PostgreSQL est installé et en cours d'exécution
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL n'est pas installé. Installez-le d'abord:"
    echo "   - macOS: brew install postgresql"
    echo "   - Ubuntu: sudo apt-get install postgresql"
    echo "   - Docker: docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres"
    exit 1
fi

# Vérifier si le serveur PostgreSQL est en cours d'exécution
if ! pg_isready -h localhost -p 5432 &> /dev/null; then
    echo "❌ Le serveur PostgreSQL n'est pas en cours d'exécution."
    echo "   Démarrez-le avec:"
    echo "   - macOS: brew services start postgresql"
    echo "   - Ubuntu: sudo systemctl start postgresql"
    echo "   - Docker: docker start postgres"
    exit 1
fi

echo "✅ PostgreSQL est disponible"

# Créer la base de données
echo "📦 Création de la base de données..."
psql -h localhost -U postgres -f scripts/database/create_db.sql

echo "🎉 Base de données initialisée avec succès!"
echo ""
echo "📋 Informations de connexion:"
echo "   Host: localhost"
echo "   Port: 5432"
echo "   Database: police_traffic"
echo "   User: postgres"
echo ""
echo "🚀 Vous pouvez maintenant lancer l'API avec:"
echo "   go run ./cmd/server"