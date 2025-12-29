#!/bin/bash

# Script pour régénérer Ent et redémarrer le backend

echo "🔄 Régénération des entités Ent..."
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Régénérer Ent
go generate ./ent

if [ $? -eq 0 ]; then
    echo "✅ Entités Ent régénérées avec succès"
    
    echo ""
    echo "🔨 Recompilation du backend..."
    go build -v -o server ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo "✅ Backend recompilé avec succès"
        echo ""
        echo "🚀 Pour redémarrer le serveur, exécutez:"
        echo "   ./server"
        echo ""
        echo "Ou utilisez:"
        echo "   make run"
    else
        echo "❌ Erreur lors de la compilation"
        exit 1
    fi
else
    echo "❌ Erreur lors de la génération Ent"
    exit 1
fi
