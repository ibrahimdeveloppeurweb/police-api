#!/bin/bash

# Script automatique pour corriger et mettre à jour le système de contenants
# Usage: ./fix-and-update-containers.sh

set -e  # Arrêter en cas d'erreur

BACKEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"
FRONTEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-frontend-aligned"

echo "════════════════════════════════════════════════════════════════"
echo "🔧 Correction et Mise à Jour du Système de Contenants"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Étape 1 : Backend - Régénération Ent
echo "📦 ÉTAPE 1/5 : Régénération des entités Ent..."
cd "$BACKEND_DIR"

if go generate ./ent; then
    echo "✅ Entités Ent régénérées avec succès"
else
    echo "❌ Erreur lors de la génération Ent"
    exit 1
fi

echo ""

# Étape 2 : Backend - Compilation
echo "🔨 ÉTAPE 2/5 : Compilation du backend..."

if go build -v -o server ./cmd/server; then
    echo "✅ Backend compilé avec succès"
else
    echo "❌ Erreur lors de la compilation du backend"
    exit 1
fi

echo ""

# Étape 3 : Vérifier que le serveur n'est pas déjà en cours
echo "🔍 ÉTAPE 3/5 : Vérification du serveur..."

SERVER_PID=$(lsof -ti:8080 2>/dev/null || echo "")

if [ -n "$SERVER_PID" ]; then
    echo "⚠️  Un serveur est déjà en cours d'exécution sur le port 8080 (PID: $SERVER_PID)"
    echo "   Arrêt du serveur actuel..."
    kill -9 $SERVER_PID 2>/dev/null || true
    sleep 2
    echo "✅ Serveur arrêté"
fi

echo ""

# Étape 4 : Démarrer le nouveau serveur en arrière-plan
echo "🚀 ÉTAPE 4/5 : Démarrage du nouveau serveur..."

./server > /tmp/police-server.log 2>&1 &
SERVER_NEW_PID=$!

echo "   Serveur démarré (PID: $SERVER_NEW_PID)"
echo "   Logs disponibles dans: /tmp/police-server.log"
echo "   Attente du démarrage..."

# Attendre que le serveur soit prêt (max 30 secondes)
COUNTER=0
until curl -s http://localhost:8080/health > /dev/null 2>&1; do
    sleep 1
    COUNTER=$((COUNTER + 1))
    if [ $COUNTER -gt 30 ]; then
        echo "❌ Le serveur n'a pas démarré dans les temps"
        echo "   Vérifiez les logs: tail -f /tmp/police-server.log"
        exit 1
    fi
done

echo "✅ Serveur opérationnel"
echo ""

# Étape 5 : Test de l'API
echo "🧪 ÉTAPE 5/5 : Test de l'API..."

# Tester avec l'ID fourni
TEST_ID="7fa3287c-dd02-40d7-b650-47e9d7d8d296"

echo "   Test de l'endpoint: /api/objets-perdus/$TEST_ID"

RESPONSE=$(curl -s http://localhost:8080/api/objets-perdus/$TEST_ID)

# Vérifier si isContainer est présent
if echo "$RESPONSE" | grep -q "isContainer"; then
    echo "✅ Le champ 'isContainer' est présent dans la réponse"
    
    # Extraire la valeur
    IS_CONTAINER=$(echo "$RESPONSE" | grep -o '"isContainer":[^,}]*' | cut -d: -f2)
    echo "   Valeur actuelle: isContainer = $IS_CONTAINER"
else
    echo "❌ Le champ 'isContainer' est toujours absent"
    echo "   Vérifiez les logs du serveur:"
    echo "   tail -f /tmp/police-server.log"
    exit 1
fi

# Vérifier si containerDetails est présent
if echo "$RESPONSE" | grep -q "containerDetails"; then
    echo "✅ Le champ 'containerDetails' est présent dans la réponse"
else
    echo "⚠️  Le champ 'containerDetails' est absent (peut être normal si null)"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✨ CORRECTION TERMINÉE AVEC SUCCÈS"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "📊 Résumé:"
echo "   • Serveur PID: $SERVER_NEW_PID"
echo "   • URL: http://localhost:8080"
echo "   • Logs: tail -f /tmp/police-server.log"
echo ""
echo "🔄 Prochaines étapes (OPTIONNEL):"
echo ""
echo "   1. Migrer les objets existants en contenants:"
echo "      cd $BACKEND_DIR"
echo "      node scripts/migrate-containers-to-new-format.js"
echo ""
echo "   2. Ouvrir l'interface web:"
echo "      http://localhost:3000"
echo ""
echo "   3. Tester un objet de type 'Sac / Sacoche':"
echo "      http://localhost:3000/gestion/objets-perdus/$TEST_ID"
echo ""
echo "   4. Arrêter le serveur:"
echo "      kill $SERVER_NEW_PID"
echo ""
echo "════════════════════════════════════════════════════════════════"
