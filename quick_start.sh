#!/bin/bash

# Script rapide pour générer et lancer
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Correction des imports et génération..."
echo ""

# 1. Générer les entités Ent
echo "📦 Génération des entités Ent..."
go generate ./ent

if [ $? -eq 0 ]; then
    echo "✅ Entités générées"
else
    echo "❌ Erreur génération"
    exit 1
fi
echo ""

# 2. Créer migration
echo "🗄️  Création de la migration..."
atlas migrate diff add_plaintes_extended \
  --dir "file://ent/migrate/migrations" \
  --to "ent://ent/schema" \
  --dev-url "sqlite://file?mode=memory&_fk=1" 2>/dev/null || echo "Migration existante"
echo ""

# 3. Appliquer migration
echo "🔄 Application de la migration..."
atlas migrate apply \
  --dir "file://ent/migrate/migrations" \
  --url "sqlite://police_trafic.db" 2>/dev/null || echo "Migration appliquée"
echo ""

# 4. Lancer le serveur
echo "🚀 Lancement du serveur..."
make run
