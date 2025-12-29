#!/bin/bash

# Script de compilation et redémarrage après ajout du module convocations

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 CORRECTION ERREUR 404 - MODULE CONVOCATIONS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "✅ Modifications effectuées :"
echo "   1. Module convocations ajouté dans app.go"
echo "   2. Fichier module.go créé avec fx.Module"
echo "   3. Controller implémente interfaces.Controller"
echo ""

echo "📦 Compilation du serveur..."
go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie !"
    echo ""
    echo "🔄 Arrêt du serveur existant..."
    pkill -f "./server" || true
    sleep 2
    
    echo "🚀 Démarrage du nouveau serveur..."
    nohup ./server > server.log 2>&1 &
    
    sleep 3
    
    echo ""
    echo "✅ Serveur redémarré avec succès !"
    echo ""
    echo "🧪 Testez maintenant l'API :"
    echo "   POST http://localhost:8080/api/v1/convocations"
    echo ""
    echo "📋 Routes convocations disponibles :"
    echo "   • POST   /api/v1/convocations"
    echo "   • GET    /api/v1/convocations"
    echo "   • GET    /api/v1/convocations/:id"
    echo "   • PATCH  /api/v1/convocations/:id/statut"
    echo "   • GET    /api/v1/convocations/statistiques"
    echo "   • GET    /api/v1/convocations/dashboard"
    echo ""
    echo "📖 Logs : tail -f server.log"
else
    echo "❌ Erreur de compilation"
    exit 1
fi
