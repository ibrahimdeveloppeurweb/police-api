#!/bin/bash

echo "🔄 Régénération du code Ent pour le module plaintes..."

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Générer le code Ent
go generate ./ent

if [ $? -eq 0 ]; then
    echo "✅ Code Ent régénéré avec succès !"
    echo ""
    echo "📋 Nouveaux champs ajoutés au schéma plainte:"
    echo "  - suspects (JSON)"
    echo "  - temoins (JSON)"
    echo ""
    echo "🚀 Vous pouvez maintenant:"
    echo "  1. Compiler le projet: make build"
    echo "  2. Lancer le serveur: make run"
    echo ""
    echo "📝 Le formulaire frontend peut maintenant envoyer:"
    echo "  - Une liste de suspects avec nom, prénom, description, adresse"
    echo "  - Une liste de témoins avec nom, prénom, téléphone, adresse"
else
    echo "❌ Erreur lors de la régénération du code Ent"
    exit 1
fi
