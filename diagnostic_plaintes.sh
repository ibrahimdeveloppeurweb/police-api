#!/bin/bash

# Script de diagnostic pour vérifier l'état du backend

echo "🔍 DIAGNOSTIC BACKEND PLAINTES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# 1. Vérifier les schémas
echo "1️⃣  Schémas Ent dans ent/schema/ :"
ls -1 ent/schema/*.go | grep -E "preuve|acte|timeline" && echo "   ✅ Schémas trouvés" || echo "   ❌ Schémas manquants"
echo ""

# 2. Vérifier les entités générées
echo "2️⃣  Entités générées dans ent/ :"
if [ -d "ent/preuve" ]; then
    echo "   ✅ ent/preuve/"
else
    echo "   ❌ ent/preuve/ MANQUANT"
fi

if [ -d "ent/acteenquete" ]; then
    echo "   ✅ ent/acteenquete/"
else
    echo "   ❌ ent/acteenquete/ MANQUANT"
fi

if [ -d "ent/timelineevent" ]; then
    echo "   ✅ ent/timelineevent/"
else
    echo "   ❌ ent/timelineevent/ MANQUANT"
fi
echo ""

# 3. Vérifier les tables dans la DB
echo "3️⃣  Tables dans la base de données :"
if command -v sqlite3 &> /dev/null; then
    TABLES=$(sqlite3 police_trafic.db "SELECT name FROM sqlite_master WHERE type='table' AND name IN ('preuves','actes_enquete','timeline_events');" 2>/dev/null)
    if [ -n "$TABLES" ]; then
        echo "$TABLES" | while read table; do
            COUNT=$(sqlite3 police_trafic.db "SELECT COUNT(*) FROM $table;" 2>/dev/null)
            echo "   ✅ $table ($COUNT enregistrements)"
        done
    else
        echo "   ❌ Tables manquantes dans la DB"
    fi
else
    echo "   ⚠️  sqlite3 non installé, impossible de vérifier"
fi
echo ""

# 4. Vérifier le service_extended.go
echo "4️⃣  Service backend :"
if grep -q "GetPreuves" internal/modules/plainte/service_extended.go 2>/dev/null; then
    echo "   ✅ GetPreuves trouvé"
else
    echo "   ❌ GetPreuves manquant"
fi

if grep -q "AddPreuve" internal/modules/plainte/service_extended.go 2>/dev/null; then
    echo "   ✅ AddPreuve trouvé"
else
    echo "   ❌ AddPreuve manquant"
fi

if grep -q "GetTimeline" internal/modules/plainte/service_extended.go 2>/dev/null; then
    echo "   ✅ GetTimeline trouvé"
else
    echo "   ❌ GetTimeline manquant"
fi
echo ""

# 5. Vérifier si le serveur tourne
echo "5️⃣  Serveur :"
if ps aux | grep -v grep | grep "./server" > /dev/null; then
    PID=$(ps aux | grep -v grep | grep "./server" | awk '{print $2}')
    echo "   ✅ Serveur actif (PID: $PID)"
else
    echo "   ❌ Serveur non actif"
fi
echo ""

# 6. Test API simple
echo "6️⃣  Test API :"
if curl -s -f http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "   ✅ API accessible"
    
    # Tester une plainte
    PLAINTE=$(curl -s http://localhost:8080/api/plaintes 2>/dev/null | head -c 100)
    if [ -n "$PLAINTE" ]; then
        echo "   ✅ API plaintes répond"
    else
        echo "   ⚠️  API plaintes ne répond pas"
    fi
else
    echo "   ❌ API non accessible sur http://localhost:8080"
fi
echo ""

# Résumé et solution
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 RÉSUMÉ :"
echo ""

MISSING_ENTITIES=0
[ ! -d "ent/preuve" ] && MISSING_ENTITIES=1
[ ! -d "ent/acteenquete" ] && MISSING_ENTITIES=1
[ ! -d "ent/timelineevent" ] && MISSING_ENTITIES=1

if [ $MISSING_ENTITIES -eq 1 ]; then
    echo "❌ PROBLÈME : Les entités Ent ne sont pas générées"
    echo ""
    echo "💡 SOLUTION :"
    echo "   chmod +x generer_entites.sh"
    echo "   ./generer_entites.sh"
else
    echo "✅ Les entités sont générées"
    echo ""
    echo "💡 Si les APIs retournent toujours des données statiques :"
    echo "   1. Vérifiez que le serveur utilise le nouveau binaire"
    echo "   2. Redémarrez le serveur :"
    echo "      pkill -f './server' && ./server &"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
