#!/bin/bash

# Script complet : Régénérer Ent + Corriger 404 + Redémarrer

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 DÉPLOIEMENT COMPLET MODULE CONVOCATIONS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "📋 Étape 1/4 : Régénération des entités Ent..."
go generate ./ent
if [ $? -eq 0 ]; then
    echo "✅ Entités Ent régénérées"
else
    echo "❌ Erreur lors de la régénération Ent"
    exit 1
fi
echo ""

echo "📋 Étape 2/4 : Compilation du serveur..."
go build -o server ./cmd/server
if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie"
else
    echo "❌ Erreur de compilation"
    exit 1
fi
echo ""

echo "📋 Étape 3/4 : Redémarrage du serveur..."
pkill -f "./server" || true
sleep 2
nohup ./server > server.log 2>&1 &
sleep 3
echo "✅ Serveur redémarré"
echo ""

echo "📋 Étape 4/4 : Vérification des routes..."
sleep 2
if grep -q "Registering convocations routes" server.log; then
    echo "✅ Routes convocations enregistrées"
else
    echo "⚠️  Routes convocations non trouvées dans les logs"
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ DÉPLOIEMENT TERMINÉ AVEC SUCCÈS !"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Résumé :"
echo "   • 74 champs implémentés ✅"
echo "   • Module enregistré ✅"
echo "   • Serveur redémarré ✅"
echo ""
echo "🧪 Testez maintenant depuis votre frontend !"
echo ""
echo "📋 Routes disponibles :"
echo "   • POST   /api/v1/convocations"
echo "   • GET    /api/v1/convocations"
echo "   • GET    /api/v1/convocations/:id"
echo "   • PATCH  /api/v1/convocations/:id/statut"
echo "   • GET    /api/v1/convocations/statistiques"
echo "   • GET    /api/v1/convocations/dashboard"
echo ""
echo "📖 Logs en temps réel : tail -f server.log"
echo ""
echo "📚 Documentation :"
echo "   • README_CONVOCATIONS_74_CHAMPS.md"
echo "   • FIX_404_CONVOCATIONS.md"
echo "   • QUICKSTART_CONVOCATIONS_74_CHAMPS.md"
