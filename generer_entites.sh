#!/bin/bash

# Script simple pour générer les entités et compiler
set -e

echo "🚀 Génération des entités Plaintes..."
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# 1. Générer le code Ent
echo "📦 Étape 1/4 : Génération Ent..."
go generate ./ent
echo "✅ Code Ent généré"
echo ""

# 2. Vérifier les nouvelles entités
echo "🔍 Étape 2/4 : Vérification des entités créées..."
if [ -d "ent/preuve" ] && [ -d "ent/acteenquete" ] && [ -d "ent/timelineevent" ]; then
    echo "✅ Les 3 nouvelles entités ont été créées :"
    echo "   - ent/preuve"
    echo "   - ent/acteenquete"
    echo "   - ent/timelineevent"
else
    echo "⚠️  Certaines entités manquent, vérifiez les schémas"
fi
echo ""

# 3. Créer et appliquer la migration
echo "🗄️  Étape 3/4 : Migration de la base de données..."
atlas migrate diff add_plaintes_extended \
  --dir "file://ent/migrate/migrations" \
  --to "ent://ent/schema" \
  --dev-url "sqlite://file?mode=memory&_fk=1" 2>/dev/null || echo "Migration déjà existante"

atlas migrate apply \
  --dir "file://ent/migrate/migrations" \
  --url "sqlite://police_trafic.db" 2>/dev/null || echo "Migration déjà appliquée"

echo "✅ Migration terminée"
echo ""

# 4. Compiler
echo "🔨 Étape 4/4 : Compilation..."
go build -o server cmd/server/main.go
echo "✅ Backend compilé"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ Terminé ! Maintenant redémarrez le serveur :"
echo ""
echo "   pkill -f './server'"
echo "   ./server &"
echo ""
echo "Ou utilisez votre méthode habituelle de redémarrage"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
