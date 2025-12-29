#!/bin/bash

# Script de correction rapide - Suppression fichier en trop

set -e

echo "🔧 Correction de l'erreur de compilation..."

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "✅ Suppression du fichier service_toresponse.go en double..."
rm -f internal/modules/convocations/service_toresponse.go

echo "✅ Compilation..."
go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie !"
    
    echo "🔄 Redémarrage du serveur..."
    pkill -f "./server" || true
    sleep 2
    nohup ./server > server.log 2>&1 &
    sleep 3
    
    echo ""
    echo "✅ Serveur démarré !"
    echo "🧪 Testez : POST /api/v1/convocations"
    echo "📖 Logs : tail -f server.log"
else
    echo "❌ Erreur de compilation"
    exit 1
fi
