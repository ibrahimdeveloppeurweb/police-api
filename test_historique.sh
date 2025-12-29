#!/bin/bash

echo "🔍 Vérification de l'historique des plaintes"
echo "=============================================="
echo ""

# Configuration
API_BASE="http://localhost:8080/api"
PLAINTE_ID=""

# Couleurs
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "📋 Étape 1: Récupérer une plainte existante"
echo ""

# Récupérer la liste des plaintes
PLAINTES_RESPONSE=$(curl -s "${API_BASE}/plaintes?limit=1")
echo "Réponse API plaintes:"
echo "$PLAINTES_RESPONSE" | jq '.'
echo ""

# Extraire l'ID de la première plainte
PLAINTE_ID=$(echo "$PLAINTES_RESPONSE" | jq -r '.plaintes[0].id // .data[0].id // .[0].id // empty')

if [ -z "$PLAINTE_ID" ] || [ "$PLAINTE_ID" = "null" ]; then
    echo -e "${RED}❌ Aucune plainte trouvée${NC}"
    echo "Créez d'abord une plainte via l'interface"
    exit 1
fi

echo -e "${GREEN}✅ Plainte trouvée: ${PLAINTE_ID}${NC}"
echo ""

# Test 1: Vérifier l'endpoint historique
echo "📋 Test 1: GET /plaintes/${PLAINTE_ID}/historique"
echo ""
HISTORIQUE_RESPONSE=$(curl -s "${API_BASE}/plaintes/${PLAINTE_ID}/historique")
echo "Réponse:"
echo "$HISTORIQUE_RESPONSE" | jq '.'
echo ""

if [ "$HISTORIQUE_RESPONSE" = "null" ]; then
    echo -e "${RED}❌ L'API retourne null${NC}"
    echo "La table historique_action_plaintes n'existe probablement pas"
    echo ""
    echo "Solution: Exécutez ce SQL dans PostgreSQL:"
    echo ""
    cat << 'EOF'
CREATE TABLE IF NOT EXISTS historique_action_plaintes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plainte_id UUID NOT NULL REFERENCES plaintes(id) ON DELETE CASCADE,
    type_action VARCHAR(50) NOT NULL,
    ancienne_valeur VARCHAR(255),
    nouvelle_valeur VARCHAR(255) NOT NULL,
    observations TEXT,
    effectue_par UUID,
    effectue_par_nom VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_historique_plainte_id ON historique_action_plaintes(plainte_id);
CREATE INDEX IF NOT EXISTS idx_historique_created_at ON historique_action_plaintes(created_at DESC);
EOF
    echo ""
elif [ "$HISTORIQUE_RESPONSE" = "[]" ]; then
    echo -e "${GREEN}✅ L'API retourne un tableau vide (correct)${NC}"
    echo "La table existe mais est vide"
else
    echo -e "${GREEN}✅ L'API retourne des données${NC}"
    COUNT=$(echo "$HISTORIQUE_RESPONSE" | jq 'length')
    echo "Nombre d'entrées: $COUNT"
fi
echo ""

# Test 2: Tester le changement d'étape
echo "📋 Test 2: Changement d'étape (doit créer une entrée dans l'historique)"
echo ""

ETAPE_RESPONSE=$(curl -s -X PATCH "${API_BASE}/plaintes/${PLAINTE_ID}/etape" \
  -H "Content-Type: application/json" \
  -d '{
    "etape": "ENQUETE",
    "observations": "Test automatique - changement étape"
  }')

echo "Réponse changement d'étape:"
echo "$ETAPE_RESPONSE" | jq '.'
echo ""

if echo "$ETAPE_RESPONSE" | jq -e '.error' > /dev/null; then
    echo -e "${RED}❌ Erreur lors du changement d'étape${NC}"
else
    echo -e "${GREEN}✅ Changement d'étape effectué${NC}"
fi
echo ""

# Test 3: Vérifier si l'historique a été créé
echo "📋 Test 3: Vérifier l'historique après le changement"
echo ""
sleep 1

HISTORIQUE_AFTER=$(curl -s "${API_BASE}/plaintes/${PLAINTE_ID}/historique")
echo "Historique après changement:"
echo "$HISTORIQUE_AFTER" | jq '.'
echo ""

if [ "$HISTORIQUE_AFTER" = "null" ]; then
    echo -e "${RED}❌ Toujours null - La méthode CreateHistoriqueAction n'est pas appelée${NC}"
    echo ""
    echo "Le code backend ne crée pas automatiquement l'historique"
    echo "Vérifiez que les méthodes ont été modifiées dans service_extended.go"
elif [ "$HISTORIQUE_AFTER" = "[]" ]; then
    echo -e "${YELLOW}⚠️  Tableau vide - L'endpoint existe mais rien n'est enregistré${NC}"
    echo ""
    echo "Causes possibles:"
    echo "1. La méthode CreateHistoriqueAction n'est pas appelée dans ChangerEtape"
    echo "2. Il y a une erreur silencieuse lors de la création"
    echo "3. Le code attend GetHistoriqueActions au lieu de GetHistorique"
else
    COUNT_AFTER=$(echo "$HISTORIQUE_AFTER" | jq 'length')
    echo -e "${GREEN}✅ Historique créé avec succès !${NC}"
    echo "Nombre d'entrées: $COUNT_AFTER"
    echo ""
    echo "Dernière entrée:"
    echo "$HISTORIQUE_AFTER" | jq '.[0]'
fi
echo ""

# Test 4: Insertion SQL directe
echo "📋 Test 4: Test d'insertion SQL directe"
echo ""
echo "Tentative d'insertion d'un enregistrement de test..."
echo ""

# Note: Nécessite psql configuré
DB_NAME="police_nationale"
DB_USER="postgres"

echo "Pour tester l'insertion SQL directement, exécutez:"
echo ""
cat << EOF
psql -U $DB_USER -d $DB_NAME << 'SQL'
INSERT INTO historique_action_plaintes 
(plainte_id, type_action, ancienne_valeur, nouvelle_valeur, observations, effectue_par_nom)
VALUES 
('${PLAINTE_ID}', 'CHANGEMENT_STATUT', 'EN_COURS', 'RESOLU', 'Test manuel insertion', 'Test User');

SELECT * FROM historique_action_plaintes WHERE plainte_id = '${PLAINTE_ID}';
SQL
EOF
echo ""

# Résumé
echo "================================================"
echo "📊 RÉSUMÉ DES TESTS"
echo "================================================"
echo ""
echo "Plainte testée: ${PLAINTE_ID}"
echo ""

if [ "$HISTORIQUE_RESPONSE" = "null" ]; then
    echo -e "${RED}❌ PROBLÈME: La table n'existe pas ou GetHistoriqueActions n'est pas implémenté${NC}"
    echo ""
    echo "📝 ACTIONS À FAIRE:"
    echo "1. Créer la table avec create_historique_table.sql"
    echo "2. Implémenter GetHistoriqueActions dans service_extended.go"
    echo "3. Modifier GetHistorique dans controller.go pour appeler GetHistoriqueActions"
    echo "4. Redémarrer le backend"
elif [ "$HISTORIQUE_AFTER" = "[]" ] && [ "$HISTORIQUE_RESPONSE" = "[]" ]; then
    echo -e "${YELLOW}⚠️  PROBLÈME: La table existe mais rien n'est enregistré${NC}"
    echo ""
    echo "📝 ACTIONS À FAIRE:"
    echo "1. Ajouter CreateHistoriqueAction dans ChangerEtape"
    echo "2. Ajouter CreateHistoriqueAction dans ChangerStatut"
    echo "3. Ajouter CreateHistoriqueAction dans AssignerAgent"
    echo "4. Redémarrer le backend"
else
    echo -e "${GREEN}✅ SUCCÈS: L'historique fonctionne correctement !${NC}"
fi
echo ""
