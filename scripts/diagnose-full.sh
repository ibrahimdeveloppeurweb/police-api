#!/bin/bash

# Script de diagnostic complet
# Usage: ./diagnose-full.sh

echo "════════════════════════════════════════════════════════════════"
echo "🔬 DIAGNOSTIC COMPLET DU SYSTÈME DE CONTENANTS"
echo "════════════════════════════════════════════════════════════════"
echo ""

BACKEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"
TEST_ID="7fa3287c-dd02-40d7-b650-47e9d7d8d296"

cd "$BACKEND_DIR"

# 1. Vérifier le schéma Ent
echo "1️⃣  SCHÉMA ENT"
echo "───────────────────────────────────────────────────────────────"
if grep -q "is_container" ent/schema/objet_perdu.go; then
    echo "✅ Champ 'is_container' défini dans le schéma"
else
    echo "❌ Champ 'is_container' ABSENT du schéma"
fi

if grep -q "container_details" ent/schema/objet_perdu.go; then
    echo "✅ Champ 'container_details' défini dans le schéma"
else
    echo "❌ Champ 'container_details' ABSENT du schéma"
fi
echo ""

# 2. Vérifier le code généré par Ent
echo "2️⃣  CODE GÉNÉRÉ PAR ENT"
echo "───────────────────────────────────────────────────────────────"

if [ -f "ent/objetperdu.go" ]; then
    if grep -q "IsContainer" ent/objetperdu.go; then
        echo "✅ 'IsContainer' présent dans le code généré"
    else
        echo "❌ 'IsContainer' ABSENT du code généré"
        echo "   ⚠️  ACTION: Exécutez 'go generate ./ent'"
    fi
    
    if grep -q "ContainerDetails" ent/objetperdu.go; then
        echo "✅ 'ContainerDetails' présent dans le code généré"
    else
        echo "❌ 'ContainerDetails' ABSENT du code généré"
        echo "   ⚠️  ACTION: Exécutez 'go generate ./ent'"
    fi
    
    # Vérifier la date de modification
    ENT_SCHEMA_DATE=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" ent/schema/objet_perdu.go 2>/dev/null || stat -c "%y" ent/schema/objet_perdu.go 2>/dev/null)
    ENT_GEN_DATE=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" ent/objetperdu.go 2>/dev/null || stat -c "%y" ent/objetperdu.go 2>/dev/null)
    
    echo ""
    echo "   📅 Dates de modification:"
    echo "      Schéma:  $ENT_SCHEMA_DATE"
    echo "      Généré:  $ENT_GEN_DATE"
else
    echo "❌ Fichier ent/objetperdu.go INTROUVABLE"
fi
echo ""

# 3. Vérifier les types
echo "3️⃣  TYPES (types.go)"
echo "───────────────────────────────────────────────────────────────"

if grep -q "type InventoryItem struct" internal/modules/objets-perdus/types.go; then
    echo "✅ Type 'InventoryItem' défini"
else
    echo "❌ Type 'InventoryItem' ABSENT"
fi

if grep -q "type ContainerDetails struct" internal/modules/objets-perdus/types.go; then
    echo "✅ Type 'ContainerDetails' défini"
else
    echo "❌ Type 'ContainerDetails' ABSENT"
fi

if grep -q "IsContainer.*bool.*\`json:\"isContainer\"\`" internal/modules/objets-perdus/types.go; then
    echo "✅ 'IsContainer' dans CreateObjetPerduRequest"
else
    echo "❌ 'IsContainer' ABSENT de CreateObjetPerduRequest"
fi

if grep -q "IsContainer.*bool.*\`json:\"isContainer\"\`" internal/modules/objets-perdus/types.go; then
    # Compter les occurrences
    COUNT=$(grep -c "IsContainer" internal/modules/objets-perdus/types.go)
    echo "   📊 'IsContainer' apparaît $COUNT fois dans types.go"
fi
echo ""

# 4. Vérifier la compilation
echo "4️⃣  COMPILATION"
echo "───────────────────────────────────────────────────────────────"

if [ -f "server" ]; then
    SERVER_DATE=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" server 2>/dev/null || stat -c "%y" server 2>/dev/null)
    TYPES_DATE=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" internal/modules/objets-perdus/types.go 2>/dev/null || stat -c "%y" internal/modules/objets-perdus/types.go 2>/dev/null)
    
    echo "   📅 Dates:"
    echo "      Types:   $TYPES_DATE"
    echo "      Server:  $SERVER_DATE"
    
    # Test de compilation rapide
    echo ""
    echo "   🔨 Test de compilation..."
    if go build -o /tmp/test-server ./cmd/server 2>&1 | head -5; then
        echo "   ✅ Compilation réussie"
        rm -f /tmp/test-server
    else
        echo "   ❌ Erreur de compilation"
    fi
else
    echo "❌ Binaire 'server' INTROUVABLE"
    echo "   ⚠️  ACTION: Exécutez 'go build -o server ./cmd/server'"
fi
echo ""

# 5. Vérifier le serveur
echo "5️⃣  SERVEUR"
echo "───────────────────────────────────────────────────────────────"

if lsof -ti:8080 > /dev/null 2>&1; then
    SERVER_PID=$(lsof -ti:8080)
    echo "✅ Serveur en cours d'exécution (PID: $SERVER_PID)"
    
    # Tester l'API
    echo ""
    echo "   🧪 Test de l'API..."
    
    HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
    if [ -n "$HEALTH_RESPONSE" ]; then
        echo "   ✅ API répond (/health)"
    else
        echo "   ❌ API ne répond pas"
    fi
    
    echo ""
    echo "   📊 Test de l'objet $TEST_ID..."
    
    API_RESPONSE=$(curl -s "http://localhost:8080/api/objets-perdus/$TEST_ID")
    
    if echo "$API_RESPONSE" | jq . > /dev/null 2>&1; then
        echo "   ✅ Réponse JSON valide"
        
        # Vérifier les champs
        HAS_IS_CONTAINER=$(echo "$API_RESPONSE" | jq -r '.data.isContainer' 2>/dev/null)
        HAS_CONTAINER_DETAILS=$(echo "$API_RESPONSE" | jq -r '.data.containerDetails' 2>/dev/null)
        
        echo ""
        echo "   📋 Champs dans la réponse:"
        
        if [ "$HAS_IS_CONTAINER" != "null" ] && [ -n "$HAS_IS_CONTAINER" ]; then
            echo "   ✅ isContainer: $HAS_IS_CONTAINER"
        else
            echo "   ❌ isContainer: ABSENT ou null"
            echo "      ⚠️  Le code Ent n'a probablement pas été régénéré"
        fi
        
        if [ "$HAS_CONTAINER_DETAILS" != "null" ]; then
            if [ "$HAS_CONTAINER_DETAILS" = "{}" ] || [ -z "$HAS_CONTAINER_DETAILS" ]; then
                echo "   ⚠️  containerDetails: présent mais vide"
            else
                echo "   ✅ containerDetails: présent"
                echo ""
                echo "   📦 Détails du contenant:"
                echo "$API_RESPONSE" | jq -r '.data.containerDetails' 2>/dev/null | head -10
            fi
        else
            echo "   ❌ containerDetails: ABSENT ou null"
        fi
        
        echo ""
        echo "   📄 Tous les champs retournés:"
        echo "$API_RESPONSE" | jq -r '.data | keys[]' 2>/dev/null | sed 's/^/      - /'
        
    else
        echo "   ❌ Réponse invalide ou erreur"
        echo ""
        echo "   📄 Réponse brute:"
        echo "$API_RESPONSE" | head -20
    fi
    
else
    echo "❌ Serveur N'EST PAS en cours d'exécution"
    echo "   ⚠️  ACTION: Démarrez le serveur avec './server'"
fi
echo ""

# 6. Vérifier la base de données
echo "6️⃣  BASE DE DONNÉES"
echo "───────────────────────────────────────────────────────────────"

if command -v psql > /dev/null 2>&1; then
    echo "   🔍 Vérification de l'objet $TEST_ID dans la BDD..."
    
    DB_RESULT=$(psql -h localhost -U postgres -d police_traffic -t -c "
        SELECT 
            is_container,
            container_details IS NOT NULL as has_details,
            jsonb_pretty(container_details) 
        FROM objets_perdus 
        WHERE id = '$TEST_ID';
    " 2>&1)
    
    if echo "$DB_RESULT" | grep -q "ERROR"; then
        echo "   ❌ Erreur de connexion à la base de données"
        echo "   $DB_RESULT"
    else
        echo "   ✅ Connexion à la base de données réussie"
        echo ""
        echo "   📊 Données dans la BDD:"
        echo "$DB_RESULT"
    fi
else
    echo "   ⚠️  psql non disponible, impossible de vérifier la BDD"
fi
echo ""

# 7. DIAGNOSTIC FINAL
echo "════════════════════════════════════════════════════════════════"
echo "📋 DIAGNOSTIC FINAL"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Déterminer l'action requise
NEEDS_ENT_REGEN=false
NEEDS_RECOMPILE=false
NEEDS_RESTART=false
NEEDS_MIGRATION=false

if ! grep -q "IsContainer" ent/objetperdu.go 2>/dev/null; then
    NEEDS_ENT_REGEN=true
fi

if [ "$NEEDS_ENT_REGEN" = true ] || ! [ -f "server" ]; then
    NEEDS_RECOMPILE=true
fi

if [ "$HAS_IS_CONTAINER" = "null" ] || [ -z "$HAS_IS_CONTAINER" ]; then
    if lsof -ti:8080 > /dev/null 2>&1; then
        NEEDS_RESTART=true
    fi
fi

if [ "$HAS_IS_CONTAINER" = "false" ] && [ "$HAS_CONTAINER_DETAILS" = "null" ]; then
    NEEDS_MIGRATION=true
fi

echo "🎯 ACTIONS REQUISES:"
echo ""

if [ "$NEEDS_ENT_REGEN" = true ]; then
    echo "1️⃣  ⚠️  CRITIQUE: Régénérer Ent"
    echo "   Commande: go generate ./ent"
    echo ""
fi

if [ "$NEEDS_RECOMPILE" = true ]; then
    echo "2️⃣  ⚠️  CRITIQUE: Recompiler le backend"
    echo "   Commande: go build -o server ./cmd/server"
    echo ""
fi

if [ "$NEEDS_RESTART" = true ]; then
    echo "3️⃣  ⚠️  CRITIQUE: Redémarrer le serveur"
    echo "   Commande: killall server && ./server"
    echo ""
fi

if [ "$NEEDS_MIGRATION" = true ]; then
    echo "4️⃣  📌 OPTIONNEL: Migrer les données existantes"
    echo "   Commande: node scripts/migrate-containers-to-new-format.js"
    echo ""
fi

if [ "$NEEDS_ENT_REGEN" = false ] && [ "$NEEDS_RECOMPILE" = false ] && [ "$NEEDS_RESTART" = false ]; then
    echo "✅ TOUT EST OK !"
    echo ""
    echo "   Le système est correctement configuré."
    
    if [ "$NEEDS_MIGRATION" = true ]; then
        echo ""
        echo "   💡 Note: L'objet testé a été créé avant la mise à jour."
        echo "   Pour tester le système complet, créez un nouvel objet via:"
        echo "   http://localhost:3000/gestion/objets-perdus/form"
    fi
else
    echo "🚀 SOLUTION RAPIDE (1 commande):"
    echo ""
    echo "   chmod +x scripts/complete-setup-and-test.sh"
    echo "   ./scripts/complete-setup-and-test.sh"
    echo ""
    echo "   Cette commande fait TOUT automatiquement !"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
