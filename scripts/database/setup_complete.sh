#!/bin/bash

# Script de configuration complète de la base de données
# Usage: ./scripts/database/setup_complete.sh

set -e

echo "🚀 Configuration complète de la base de données PostgreSQL"
echo "=========================================================="

# Étape 1: Vérifier PostgreSQL
echo ""
echo "1️⃣  Vérification de PostgreSQL..."
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL n'est pas installé."
    echo ""
    echo "📦 Installation rapide:"
    echo "   macOS:   brew install postgresql && brew services start postgresql"
    echo "   Ubuntu:  sudo apt-get install postgresql postgresql-contrib"
    echo "   Docker:  docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres"
    exit 1
fi

if ! pg_isready -h localhost -p 5432 &> /dev/null; then
    echo "❌ PostgreSQL n'est pas en cours d'exécution."
    echo ""
    echo "🔧 Pour le démarrer:"
    echo "   macOS:   brew services start postgresql"
    echo "   Ubuntu:  sudo systemctl start postgresql"
    echo "   Docker:  docker start postgres"
    exit 1
fi

echo "✅ PostgreSQL est disponible"

# Étape 2: Créer la base de données
echo ""
echo "2️⃣  Création de la base de données..."
if psql -h localhost -U postgres -lqt | cut -d \| -f 1 | grep -qw police_traffic; then
    echo "⚠️  La base 'police_traffic' existe déjà"
    read -p "🤔 Voulez-vous la supprimer et la recréer? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️  Suppression de l'ancienne base..."
        psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS police_traffic;"
    else
        echo "⏭️  Utilisation de la base existante"
    fi
fi

if ! psql -h localhost -U postgres -lqt | cut -d \| -f 1 | grep -qw police_traffic; then
    echo "📦 Création de la nouvelle base..."
    psql -h localhost -U postgres -c "CREATE DATABASE police_traffic;"
fi

echo "✅ Base de données prête"

# Étape 3: Exécuter les migrations
echo ""
echo "3️⃣  Exécution des migrations Ent..."
if ! go run ./cmd/migrate; then
    echo "❌ Erreur lors des migrations"
    exit 1
fi

echo "✅ Migrations terminées"

# Étape 4: Insérer les données de test
echo ""
echo "4️⃣  Insertion des données de test..."
read -p "🤔 Voulez-vous insérer des données de test? (Y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    if ! go run ./cmd/seed; then
        echo "❌ Erreur lors de l'insertion des données"
        exit 1
    fi
    echo "✅ Données de test insérées"
else
    echo "⏭️  Données de test ignorées"
fi

# Étape 5: Test de connexion
echo ""
echo "5️⃣  Test de connexion..."
echo "🔍 Vérification des tables créées:"
psql -h localhost -U postgres -d police_traffic -c "\\dt"

echo ""
echo "🎉 Configuration terminée avec succès!"
echo ""
echo "📋 Informations de connexion:"
echo "   Host:     localhost"
echo "   Port:     5432"
echo "   Database: police_traffic"
echo "   User:     postgres"
echo ""
echo "🚀 Vous pouvez maintenant lancer l'API:"
echo "   go run ./cmd/server"
echo ""
echo "🧪 Ou tester la connexion directement:"
echo "   psql -h localhost -U postgres -d police_traffic"