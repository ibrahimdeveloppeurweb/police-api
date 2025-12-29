#!/bin/bash

# Script de vérification rapide
# Usage: ./check-container-system.sh

echo "🔍 Vérification du Système de Contenants"
echo "========================================"
echo ""

BACKEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"
FRONTEND_DIR="/Users/ibrahim/Documents/police1/police-trafic-frontend-aligned"

# Vérifier le schéma Ent
echo "1️⃣  Vérification du schéma Ent..."
if grep -q "is_container" "$BACKEND_DIR/ent/schema/objet_perdu.go"; then
    echo "   ✅ Schéma contient 'is_container'"
else
    echo "   ❌ Schéma ne contient pas 'is_container'"
fi

if grep -q "container_details" "$BACKEND_DIR/ent/schema/objet_perdu.go"; then
    echo "   ✅ Schéma contient 'container_details'"
else
    echo "   ❌ Schéma ne contient pas 'container_details'"
fi

echo ""

# Vérifier les types
echo "2️⃣  Vérification des types..."
if grep -q "InventoryItem" "$BACKEND_DIR/internal/modules/objets-perdus/types.go"; then
    echo "   ✅ Type InventoryItem défini"
else
    echo "   ❌ Type InventoryItem non défini"
fi

if grep -q "ContainerDetails" "$BACKEND_DIR/internal/modules/objets-perdus/types.go"; then
    echo "   ✅ Type ContainerDetails défini"
else
    echo "   ❌ Type ContainerDetails non défini"
fi

echo ""

# Vérifier le hook frontend
echo "3️⃣  Vérification du hook frontend..."
if grep -q "isContainer" "$FRONTEND_DIR/src/hooks/useObjetPerduDetail.ts"; then
    echo "   ✅ Hook contient 'isContainer'"
else
    echo "   ❌ Hook ne contient pas 'isContainer'"
fi

if grep -q "ContainerDetails" "$FRONTEND_DIR/src/hooks/useObjetPerduDetail.ts"; then
    echo "   ✅ Hook contient 'ContainerDetails'"
else
    echo "   ❌ Hook ne contient pas 'ContainerDetails'"
fi

echo ""

# Vérifier la page de détail
echo "4️⃣  Vérification de la page de détail..."
if grep -q "isContainer" "$FRONTEND_DIR/src/app/gestion/objets-perdus/[id]/page.tsx"; then
    echo "   ✅ Page détail contient 'isContainer'"
else
    echo "   ❌ Page détail ne contient pas 'isContainer'"
fi

if grep -q "Inventaire du contenant" "$FRONTEND_DIR/src/app/gestion/objets-perdus/[id]/page.tsx"; then
    echo "   ✅ Page détail contient section inventaire"
else
    echo "   ❌ Page détail ne contient pas section inventaire"
fi

echo ""

# Vérifier si le serveur est en cours d'exécution
echo "5️⃣  Vérification du serveur..."
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "   ✅ Serveur en cours d'exécution sur le port 8080"
    
    # Tester l'API
    echo ""
    echo "6️⃣  Test de l'API..."
    
    RESPONSE=$(curl -s http://localhost:8080/health)
    if [ -n "$RESPONSE" ]; then
        echo "   ✅ API répond"
        
        # Test avec un objet perdu
        TEST_RESPONSE=$(curl -s http://localhost:8080/api/objets-perdus/7fa3287c-dd02-40d7-b650-47e9d7d8d296)
        
        if echo "$TEST_RESPONSE" | grep -q "isContainer"; then
            echo "   ✅ L'API retourne le champ 'isContainer'"
        else
            echo "   ❌ L'API ne retourne PAS le champ 'isContainer'"
            echo "   ⚠️  ACTION REQUISE: Exécutez ./scripts/fix-and-update-containers.sh"
        fi
    else
        echo "   ❌ API ne répond pas"
    fi
else
    echo "   ❌ Serveur n'est pas en cours d'exécution"
    echo "   ⚠️  Démarrez le serveur avec: cd $BACKEND_DIR && ./server"
fi

echo ""

# Vérifier les scripts
echo "7️⃣  Vérification des scripts..."
SCRIPTS=(
    "$BACKEND_DIR/scripts/fix-and-update-containers.sh"
    "$BACKEND_DIR/scripts/regenerate-ent.sh"
    "$BACKEND_DIR/scripts/migrate-containers-to-new-format.js"
    "$BACKEND_DIR/scripts/migrate_containers.sql"
)

for script in "${SCRIPTS[@]}"; do
    if [ -f "$script" ]; then
        echo "   ✅ $(basename $script)"
    else
        echo "   ❌ $(basename $script) manquant"
    fi
done

echo ""
echo "========================================"
echo "✨ Vérification terminée"
echo ""
echo "📝 Prochaines étapes:"
echo ""

# Si l'API ne retourne pas les champs
if ! curl -s http://localhost:8080/api/objets-perdus/7fa3287c-dd02-40d7-b650-47e9d7d8d296 2>/dev/null | grep -q "isContainer"; then
    echo "⚠️  IMPORTANT : L'API ne retourne pas encore les nouveaux champs"
    echo ""
    echo "   Exécutez cette commande pour corriger:"
    echo "   cd $BACKEND_DIR"
    echo "   chmod +x scripts/fix-and-update-containers.sh"
    echo "   ./scripts/fix-and-update-containers.sh"
else
    echo "✅ Système opérationnel !"
    echo ""
    echo "   Pour migrer les données existantes (optionnel):"
    echo "   cd $BACKEND_DIR"
    echo "   node scripts/migrate-containers-to-new-format.js"
fi

echo ""
