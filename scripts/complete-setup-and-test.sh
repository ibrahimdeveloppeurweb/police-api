#!/bin/bash

# Script complet pour tout mettre à jour et tester
# Usage: ./complete-setup-and-test.sh

set -e

BACKEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"

echo "════════════════════════════════════════════════════════════════"
echo "🚀 Setup Complet du Système de Contenants"
echo "════════════════════════════════════════════════════════════════"
echo ""

cd "$BACKEND_DIR"

# Étape 1 : Régénérer Ent
echo "1️⃣  Régénération des entités Ent..."
if go generate ./ent; then
    echo "✅ Ent régénéré"
else
    echo "❌ Erreur Ent"
    exit 1
fi
echo ""

# Étape 2 : Compiler
echo "2️⃣  Compilation du backend..."
if go build -v -o server ./cmd/server; then
    echo "✅ Backend compilé"
else
    echo "❌ Erreur compilation"
    exit 1
fi
echo ""

# Étape 3 : Arrêter l'ancien serveur
echo "3️⃣  Gestion du serveur..."
SERVER_PID=$(lsof -ti:8080 2>/dev/null || echo "")
if [ -n "$SERVER_PID" ]; then
    echo "   Arrêt de l'ancien serveur (PID: $SERVER_PID)..."
    kill -9 $SERVER_PID 2>/dev/null || true
    sleep 2
fi
echo ""

# Étape 4 : Démarrer le nouveau serveur
echo "4️⃣  Démarrage du serveur..."
./server > /tmp/police-server.log 2>&1 &
NEW_PID=$!
echo "   Serveur démarré (PID: $NEW_PID)"
echo "   Attente du démarrage..."

COUNTER=0
until curl -s http://localhost:8080/health > /dev/null 2>&1; do
    sleep 1
    COUNTER=$((COUNTER + 1))
    if [ $COUNTER -gt 30 ]; then
        echo "❌ Timeout"
        exit 1
    fi
done
echo "✅ Serveur opérationnel"
echo ""

# Étape 5 : Tester l'API
echo "5️⃣  Test de l'API..."

# Test 1 : Vérifier qu'un objet existant retourne isContainer
echo "   Test 1 : Objet existant..."
RESPONSE=$(curl -s http://localhost:8080/api/objets-perdus/7fa3287c-dd02-40d7-b650-47e9d7d8d296)

if echo "$RESPONSE" | grep -q "isContainer"; then
    IS_CONTAINER=$(echo "$RESPONSE" | grep -o '"isContainer":[^,}]*' | cut -d: -f2)
    echo "   ✅ isContainer présent (valeur: $IS_CONTAINER)"
else
    echo "   ❌ isContainer absent"
    exit 1
fi
echo ""

# Test 2 : Créer un nouvel objet contenant via l'API
echo "   Test 2 : Création d'un objet contenant..."

CREATE_PAYLOAD='{
  "typeObjet": "Sac / Sacoche",
  "description": "Test de création d'\''un contenant",
  "valeurEstimee": "5000 FCFA",
  "couleur": "Noir",
  "isContainer": true,
  "containerDetails": {
    "type": "sac",
    "couleur": "Noir",
    "marque": "Nike",
    "taille": "Moyen",
    "signesDistinctifs": "Logo Nike blanc sur le côté",
    "inventory": [
      {
        "category": "telephone",
        "icon": "smartphone",
        "name": "iPhone 13",
        "color": "Noir",
        "brand": "Apple",
        "serial": "ABC123456"
      },
      {
        "category": "carte",
        "icon": "credit-card",
        "name": "Visa SGBCI",
        "color": "Bleue",
        "cardType": "VISA",
        "cardBank": "SGBCI",
        "cardLast4": "1234"
      }
    ]
  },
  "declarant": {
    "nom": "TEST",
    "prenom": "Utilisateur",
    "telephone": "+2250700000000",
    "email": "test@test.com",
    "adresse": "Adresse test",
    "cni": "TEST123"
  },
  "lieuPerte": "Test Lieu",
  "datePerte": "'$(date +%Y-%m-%d)'",
  "observations": "Test automatique"
}'

CREATE_RESPONSE=$(curl -s -X POST \
  http://localhost:8080/api/objets-perdus \
  -H "Content-Type: application/json" \
  -d "$CREATE_PAYLOAD" 2>&1)

if echo "$CREATE_RESPONSE" | grep -q '"id"'; then
    NEW_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "   ✅ Objet créé avec succès (ID: $NEW_ID)"
    
    # Vérifier que le nouvel objet a bien containerDetails
    echo ""
    echo "   Test 3 : Vérification du nouvel objet..."
    NEW_RESPONSE=$(curl -s "http://localhost:8080/api/objets-perdus/$NEW_ID")
    
    if echo "$NEW_RESPONSE" | grep -q '"containerDetails"'; then
        echo "   ✅ containerDetails présent"
        
        # Afficher l'inventaire
        INVENTORY_COUNT=$(echo "$NEW_RESPONSE" | grep -o '"inventory":\[' | wc -l)
        echo "   ✅ Inventaire détecté"
        
        echo ""
        echo "   📊 Détails du nouvel objet:"
        echo "$NEW_RESPONSE" | jq '{
          numero: .data.numero,
          typeObjet: .data.typeObjet,
          isContainer: .data.isContainer,
          containerDetails: .data.containerDetails
        }' 2>/dev/null || echo "$NEW_RESPONSE" | head -20
    else
        echo "   ⚠️  containerDetails absent"
    fi
else
    echo "   ❌ Erreur lors de la création"
    echo "   Réponse: $CREATE_RESPONSE" | head -10
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✨ SETUP TERMINÉ"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "📊 Résultats:"
echo "   • Serveur PID: $NEW_PID"
echo "   • API disponible: http://localhost:8080"
echo "   • Frontend: http://localhost:3000"
echo ""

if [ -n "$NEW_ID" ]; then
    echo "🎯 Testez le nouvel objet dans l'interface:"
    echo "   http://localhost:3000/gestion/objets-perdus/$NEW_ID"
    echo ""
fi

echo "📝 Commandes utiles:"
echo "   • Logs serveur: tail -f /tmp/police-server.log"
echo "   • Arrêter serveur: kill $NEW_PID"
echo "   • Migrer objets existants: node scripts/migrate-containers-to-new-format.js"
echo ""
echo "════════════════════════════════════════════════════════════════"
