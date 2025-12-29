#!/bin/bash

echo "🔧 Génération des entités Ent pour l'historique des actions..."

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Générer les entités
echo "📝 Génération du code Ent..."
go generate ./ent

echo "✅ Génération terminée !"
echo ""
echo "📋 Prochaines étapes:"
echo "1. Vérifier que l'entité HistoriqueActionPlainte a été générée dans ent/"
echo "2. Redémarrer le backend pour appliquer les changements"
echo "3. Tester l'endpoint GET /plaintes/:id/historique"
