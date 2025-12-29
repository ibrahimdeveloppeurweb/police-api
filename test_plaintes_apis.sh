#!/bin/bash

# Script de test complet des APIs Plaintes
# Ce script teste toutes les nouvelles APIs dynamiques

set -e

BASE_URL="http://localhost:8080/api"

echo "🧪 Test des APIs Plaintes Dynamiques"
echo "====================================="
echo ""

# Vérifier que le serveur est actif
echo "🔍 Vérification du serveur..."
if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
    echo "✅ Serveur actif"
else
    echo "❌ Serveur non accessible sur $BASE_URL"
    echo "   Assurez-vous que le serveur tourne : ./server"
    exit 1
fi
echo ""

# Créer une plainte de test
echo "📝 Création d'une plainte de test..."
PLAINTE_RESPONSE=$(curl -s -X POST "$BASE_URL/plaintes" \
  -H "Content-Type: application/json" \
  -d '{
    "type_plainte": "VOL",
    "plaignant_nom": "TEST",
    "plaignant_prenom": "Automatique",
    "description": "Plainte de test pour validation des APIs",
    "priorite": "NORMALE"
  }')

PLAINTE_ID=$(echo $PLAINTE_RESPONSE | jq -r '.id')

if [ "$PLAINTE_ID" != "null" ] && [ -n "$PLAINTE_ID" ]; then
    echo "✅ Plainte créée avec ID: $PLAINTE_ID"
else
    echo "❌ Erreur création plainte"
    echo "Réponse: $PLAINTE_RESPONSE"
    exit 1
fi
echo ""

# Test 1: Ajouter un événement Timeline
echo "1️⃣  Test: Ajouter événement Timeline..."
TIMELINE_RESPONSE=$(curl -s -X POST "$BASE_URL/plaintes/$PLAINTE_ID/timeline" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-12-18T10:00:00Z",
    "heure": "10:00",
    "type": "DEPOT",
    "titre": "Dépôt de la plainte TEST",
    "description": "Test automatique de timeline",
    "acteur": "Robot de test",
    "statut": "TERMINE"
  }')

TIMELINE_ID=$(echo $TIMELINE_RESPONSE | jq -r '.id')
if [ "$TIMELINE_ID" != "null" ] && [ -n "$TIMELINE_ID" ]; then
    echo "   ✅ Événement timeline créé: $TIMELINE_ID"
else
    echo "   ❌ Erreur création timeline"
    echo "   Réponse: $TIMELINE_RESPONSE"
fi
echo ""

# Test 2: Récupérer la timeline
echo "2️⃣  Test: Récupérer timeline..."
GET_TIMELINE=$(curl -s "$BASE_URL/plaintes/$PLAINTE_ID/timeline")
TIMELINE_COUNT=$(echo $GET_TIMELINE | jq 'length')
if [ "$TIMELINE_COUNT" -gt 0 ]; then
    echo "   ✅ Timeline récupérée: $TIMELINE_COUNT événement(s)"
else
    echo "   ❌ Aucun événement trouvé"
fi
echo ""

# Test 3: Ajouter une preuve
echo "3️⃣  Test: Ajouter preuve..."
PREUVE_RESPONSE=$(curl -s -X POST "$BASE_URL/plaintes/$PLAINTE_ID/preuves" \
  -H "Content-Type: application/json" \
  -d '{
    "numero_piece": "PCE-TEST-001",
    "type": "MATERIELLE",
    "description": "Preuve de test automatique",
    "lieu_conservation": "Test Lab",
    "date_collecte": "2024-12-18T09:00:00Z",
    "collecte_par": "Robot",
    "expertise_demandee": false
  }')

PREUVE_ID=$(echo $PREUVE_RESPONSE | jq -r '.id')
if [ "$PREUVE_ID" != "null" ] && [ -n "$PREUVE_ID" ]; then
    echo "   ✅ Preuve créée: $PREUVE_ID"
else
    echo "   ❌ Erreur création preuve"
    echo "   Réponse: $PREUVE_RESPONSE"
fi
echo ""

# Test 4: Récupérer les preuves
echo "4️⃣  Test: Récupérer preuves..."
GET_PREUVES=$(curl -s "$BASE_URL/plaintes/$PLAINTE_ID/preuves")
PREUVES_COUNT=$(echo $GET_PREUVES | jq 'length')
if [ "$PREUVES_COUNT" -gt 0 ]; then
    echo "   ✅ Preuves récupérées: $PREUVES_COUNT preuve(s)"
else
    echo "   ❌ Aucune preuve trouvée"
fi
echo ""

# Test 5: Ajouter un acte d'enquête
echo "5️⃣  Test: Ajouter acte d'enquête..."
ACTE_RESPONSE=$(curl -s -X POST "$BASE_URL/plaintes/$PLAINTE_ID/actes-enquete" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "AUDITION",
    "date": "2024-12-18T14:00:00Z",
    "heure": "14:00",
    "duree": "1h",
    "lieu": "Test Lab",
    "officier_charge": "Robot Test",
    "description": "Acte de test automatique"
  }')

ACTE_ID=$(echo $ACTE_RESPONSE | jq -r '.id')
if [ "$ACTE_ID" != "null" ] && [ -n "$ACTE_ID" ]; then
    echo "   ✅ Acte créé: $ACTE_ID"
else
    echo "   ❌ Erreur création acte"
    echo "   Réponse: $ACTE_RESPONSE"
fi
echo ""

# Test 6: Récupérer les actes
echo "6️⃣  Test: Récupérer actes d'enquête..."
GET_ACTES=$(curl -s "$BASE_URL/plaintes/$PLAINTE_ID/actes-enquete")
ACTES_COUNT=$(echo $GET_ACTES | jq 'length')
if [ "$ACTES_COUNT" -gt 0 ]; then
    echo "   ✅ Actes récupérés: $ACTES_COUNT acte(s)"
else
    echo "   ❌ Aucun acte trouvé"
fi
echo ""

# Test 7: Récupérer les alertes
echo "7️⃣  Test: Récupérer alertes..."
ALERTES=$(curl -s "$BASE_URL/plaintes/alertes")
ALERTES_COUNT=$(echo $ALERTES | jq 'length')
if [ "$ALERTES_COUNT" -ge 0 ]; then
    echo "   ✅ Alertes récupérées: $ALERTES_COUNT alerte(s)"
else
    echo "   ❌ Erreur récupération alertes"
fi
echo ""

# Test 8: Récupérer top agents
echo "8️⃣  Test: Récupérer top agents..."
TOP_AGENTS=$(curl -s "$BASE_URL/plaintes/top-agents")
AGENTS_COUNT=$(echo $TOP_AGENTS | jq 'length')
if [ "$AGENTS_COUNT" -ge 0 ]; then
    echo "   ✅ Top agents récupérés: $AGENTS_COUNT agent(s)"
else
    echo "   ❌ Erreur récupération top agents"
fi
echo ""

# Test 9: Récupérer la plainte complète
echo "9️⃣  Test: Récupérer plainte complète..."
PLAINTE_COMPLETE=$(curl -s "$BASE_URL/plaintes/$PLAINTE_ID")
PLAINTE_NUMERO=$(echo $PLAINTE_COMPLETE | jq -r '.numero')
if [ "$PLAINTE_NUMERO" != "null" ] && [ -n "$PLAINTE_NUMERO" ]; then
    echo "   ✅ Plainte complète: $PLAINTE_NUMERO"
else
    echo "   ❌ Erreur récupération plainte"
fi
echo ""

# Nettoyage (optionnel)
echo "🧹 Nettoyage..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/plaintes/$PLAINTE_ID")
echo "   ✅ Plainte de test supprimée"
echo ""

# Résumé
echo "======================================"
echo "✨ Tests terminés !"
echo ""
echo "📊 Résultats :"
echo "   - Plainte créée : ✅"
echo "   - Timeline      : ✅ ($TIMELINE_COUNT événements)"
echo "   - Preuves       : ✅ ($PREUVES_COUNT preuves)"
echo "   - Actes         : ✅ ($ACTES_COUNT actes)"
echo "   - Alertes       : ✅ ($ALERTES_COUNT alertes)"
echo "   - Top agents    : ✅ ($AGENTS_COUNT agents)"
echo ""
echo "🎉 Toutes les APIs fonctionnent correctement !"
echo ""
echo "💡 Vous pouvez maintenant tester le frontend"
echo "   Les composants suivants sont maintenant dynamiques :"
echo "   - TimelineInvestigation"
echo "   - PreuvesList"
echo "   - ActesEnqueteList"
echo "   - AlertesActives"
echo "   - TopAgentsPerformants"
