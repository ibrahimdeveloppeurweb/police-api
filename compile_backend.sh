#!/bin/bash

echo "🔧 Compilation du backend avec tous les champs..."
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "📦 Installation des dépendances..."
go mod tidy

echo "🏗️  Compilation..."
go build -o bin/server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie!"
    echo ""
    echo "Pour démarrer le serveur:"
    echo "  ./bin/server"
    echo ""
    echo "Ou directement:"
    echo "  go run cmd/server/main.go"
else
    echo "❌ Erreur de compilation"
    exit 1
fi
