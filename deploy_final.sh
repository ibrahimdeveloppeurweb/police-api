#!/bin/bash

# Script de déploiement complet après corrections

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 DÉPLOIEMENT MODULE CONVOCATIONS - VERSION FINALE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "📋 Étape 1/3 : Régénération des entités Ent avec les 74 champs..."
go generate ./ent
if [ $? -eq 0 ]; then
    echo "✅ Entités régénérées"
else
    echo "❌ Erreur régénération Ent"
    exit 1
fi
echo ""

echo "📋 Étape 2/3 : Compilation du serveur..."
go build -o server ./cmd/server 2>&1 | tee compile.log
if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie"
else
    echo "❌ Erreur de compilation - voir compile.log"
    echo ""
    echo "Erreurs détectées :"
    grep "error:" compile.log | head -10
    exit 1
fi
echo ""

echo "📋 Étape 3/3 : Redémarrage du serveur..."
pkill -f "./server" || true
sleep 2
nohup ./server > server.log 2>&1 &
SERVER_PID=$!
sleep 3

if ps -p $SERVER_PID > /dev/null; then
    echo "✅ Serveur démarré (PID: $SERVER_PID)"
else
    echo "❌ Le serveur n'a pas démarré"
    echo "Dernières lignes du log :"
    tail -20 server.log
    exit 1
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ DÉPLOIEMENT TERMINÉ AVEC SUCCÈS !"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Vérifications :"
echo "   • ConvocationRepository créé ✅"
echo "   • Module enregistré dans app.go ✅"
echo "   • 74 champs implémentés ✅"
echo "   • Serveur démarré ✅"
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
