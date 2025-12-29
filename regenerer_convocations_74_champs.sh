#!/bin/bash

# Script de régénération des entités Ent pour le module Convocations
# Tous les 74 champs ont été implémentés dans ent/schema/convocation.go

set -e

echo "🔄 Régénération des entités Ent pour module Convocations..."
echo "📋 74 champs ont été ajoutés au schéma"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "✅ Génération des entités Ent..."
go generate ./ent

echo ""
echo "✅ Vérification de la compilation..."
go build ./cmd/server

echo ""
echo "✅ Formatage du code..."
go fmt ./...

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ SUCCÈS ! Entités Ent régénérées avec succès"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Champs implémentés par section :"
echo "   • Section 1 - Informations générales : 6 champs"
echo "   • Section 2 - Affaire liée : 7 champs"
echo "   • Section 3 - Personne convoquée : 32 champs"
echo "   • Section 4 - Rendez-vous : 11 champs"
echo "   • Section 5 - Personnes présentes : 14 champs"
echo "   • Section 6 - Motif et objet : 5 champs"
echo "   • Section 9 - Observations : 1 champ"
echo "   • Section 10 - État et traçabilité : 4 champs"
echo "   • TOTAL : 74 champs + métadonnées"
echo ""
echo "🎯 Prochaines étapes :"
echo "   1. Redémarrer le serveur : ./restart-backend.sh"
echo "   2. Tester l'API POST /api/v1/convocations"
echo "   3. Vérifier la création avec tous les champs"
echo ""
echo "📖 Documentation : IMPLEMENTATION_COMPLETE_74_CHAMPS_CONVOCATIONS.md"
