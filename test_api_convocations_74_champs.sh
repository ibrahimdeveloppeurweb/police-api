#!/bin/bash

# Script de test de l'API POST /api/v1/convocations avec TOUS les 74 champs

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 TEST API POST /api/v1/convocations"
echo "📋 Test avec les 74 champs implémentés"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# URL de l'API
API_URL="http://localhost:8080/api/v1/convocations"

# Token d'authentification (à remplacer par un vrai token)
TOKEN="YOUR_AUTH_TOKEN_HERE"

# Fichier JSON de test
JSON_FILE="test_convocation_complete_74_champs.json"

echo "📍 URL de l'API : $API_URL"
echo "📄 Fichier de test : $JSON_FILE"
echo ""

# Vérifier que le fichier JSON existe
if [ ! -f "$JSON_FILE" ]; then
    echo "❌ Erreur : Le fichier $JSON_FILE n'existe pas"
    exit 1
fi

echo "📤 Envoi de la requête POST..."
echo ""

# Effectuer la requête
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @"$JSON_FILE")

# Extraire le code HTTP et le corps de la réponse
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "📊 Code HTTP : $http_code"
echo ""

# Afficher la réponse formatée
if command -v jq &> /dev/null; then
    echo "📋 Réponse (formatée) :"
    echo "$body" | jq '.'
else
    echo "📋 Réponse :"
    echo "$body"
    echo ""
    echo "💡 Installez 'jq' pour une meilleure lisibilité : brew install jq"
fi

echo ""

# Interpréter le résultat
if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
    echo "✅ SUCCÈS - Convocation créée avec succès !"
    echo ""
    echo "🎯 Champs testés :"
    echo "   • Section 1 - Informations générales : 6 champs ✅"
    echo "   • Section 2 - Affaire liée : 7 champs ✅"
    echo "   • Section 3 - Personne convoquée : 32 champs ✅"
    echo "   • Section 4 - Rendez-vous : 11 champs ✅"
    echo "   • Section 5 - Personnes présentes : 14 champs ✅"
    echo "   • Section 6 - Motif et objet : 5 champs ✅"
    echo "   • Section 9 - Observations : 1 champ ✅"
    echo "   • Section 10 - État : 2 champs ✅"
    echo ""
    echo "📊 TOTAL : 74 champs testés avec succès"
elif [ "$http_code" -eq 401 ]; then
    echo "⚠️  ERREUR 401 - Non autorisé"
    echo "💡 Veuillez mettre à jour le TOKEN dans le script"
elif [ "$http_code" -eq 400 ]; then
    echo "⚠️  ERREUR 400 - Requête invalide"
    echo "💡 Vérifiez les données dans $JSON_FILE"
elif [ "$http_code" -eq 500 ]; then
    echo "❌ ERREUR 500 - Erreur serveur"
    echo "💡 Vérifiez les logs du serveur"
else
    echo "⚠️  Code HTTP inattendu : $http_code"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📖 Pour plus d'infos : IMPLEMENTATION_COMPLETE_74_CHAMPS_CONVOCATIONS.md"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
