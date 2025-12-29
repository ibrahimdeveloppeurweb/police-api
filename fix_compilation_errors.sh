#!/bin/bash

# Script de correction complète pour les erreurs de compilation du module convocations

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 CORRECTION DES ERREURS DE COMPILATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "✅ Étape 1/4 : Régénération des entités Ent..."
go generate ./ent
echo ""

echo "✅ Étape 2/4 : Suppression des fichiers temporaires..."
rm -f internal/modules/convocations/service_toresponse.go
echo ""

echo "✅ Étape 3/4 : Compilation..."
go build -o server ./cmd/server
if [ $? -ne 0 ]; then
    echo "❌ Erreur de compilation"
    echo ""
    echo "📋 Correction manuelle nécessaire pour toResponse()"
    exit 1
fi
echo ""

echo "✅ Étape 4/4 : Redémarrage du serveur..."
pkill -f "./server" || true
sleep 2
nohup ./server > server.log 2>&1 &
sleep 3
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ CORRECTION TERMINÉE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🧪 Testez maintenant : POST /api/v1/convocations"
echo "📖 Logs : tail -f server.log"
