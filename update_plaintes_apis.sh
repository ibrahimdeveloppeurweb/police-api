#!/bin/bash

# Script de mise à jour complète des APIs Plaintes
# Ce script génère les entités Ent et crée les migrations nécessaires

set -e  # Arrêter en cas d'erreur

echo "🚀 Début de la mise à jour des APIs Plaintes..."
echo ""

# 1. Se positionner dans le bon répertoire
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# 2. Générer le code Ent
echo "📦 Génération du code Ent..."
go generate ./ent
if [ $? -eq 0 ]; then
    echo "✅ Code Ent généré avec succès"
else
    echo "❌ Erreur lors de la génération du code Ent"
    exit 1
fi
echo ""

# 3. Compiler le backend pour vérifier qu'il n'y a pas d'erreurs
echo "🔨 Compilation du backend..."
go build -o server cmd/server/main.go
if [ $? -eq 0 ]; then
    echo "✅ Backend compilé avec succès"
else
    echo "❌ Erreur lors de la compilation"
    exit 1
fi
echo ""

# 4. Créer la migration
echo "🗄️  Création de la migration..."
atlas migrate diff add_plaintes_preuves_actes_timeline \
  --dir "file://ent/migrate/migrations" \
  --to "ent://ent/schema" \
  --dev-url "sqlite://file?mode=memory&_fk=1"

if [ $? -eq 0 ]; then
    echo "✅ Migration créée avec succès"
else
    echo "⚠️  Attention: Erreur lors de la création de la migration"
    echo "   Vous devrez peut-être créer la migration manuellement"
fi
echo ""

# 5. Appliquer la migration
echo "🔄 Application de la migration..."
atlas migrate apply \
  --dir "file://ent/migrate/migrations" \
  --url "sqlite://police_trafic.db"

if [ $? -eq 0 ]; then
    echo "✅ Migration appliquée avec succès"
else
    echo "⚠️  Attention: Erreur lors de l'application de la migration"
fi
echo ""

# 6. Redémarrer le serveur
echo "🔄 Redémarrage du serveur..."
pkill -f "./server" 2>/dev/null || true
sleep 2
./server &
SERVER_PID=$!
echo "✅ Serveur redémarré (PID: $SERVER_PID)"
echo ""

echo "✨ Mise à jour terminée avec succès !"
echo ""
echo "📝 Prochaines étapes :"
echo "   1. Testez les APIs avec curl ou Postman"
echo "   2. Vérifiez que les données sont bien enregistrées"
echo "   3. Testez le frontend pour confirmer que tout fonctionne"
echo ""
echo "🔗 URLs à tester :"
echo "   - POST http://localhost:8080/api/plaintes/{id}/timeline"
echo "   - POST http://localhost:8080/api/plaintes/{id}/preuves"
echo "   - POST http://localhost:8080/api/plaintes/{id}/actes-enquete"
echo "   - GET  http://localhost:8080/api/plaintes/alertes"
echo "   - GET  http://localhost:8080/api/plaintes/top-agents"
echo ""
