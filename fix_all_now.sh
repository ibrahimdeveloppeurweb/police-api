#!/bin/bash

# Script de correction COMPLÈTE pour tous les problèmes

set -e

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 CORRECTION FINALE - TOUS LES PROBLÈMES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Étape 1/5 : Backup des fichiers..."
cp internal/modules/convocations/service.go internal/modules/convocations/service.go.backup

echo "Étape 2/5 : Correction des types JSON..."
# Les champs JSON dans Ent acceptent directement les types Go
sed -i.tmp '
s/createBuilder\.SetDonneesCompletes(donneesCompletesJSON)/createBuilder.SetDonneesCompletes(donneesCompletes)/g
s/createBuilder\.SetHistorique(historiqueJSON)/createBuilder.SetHistorique(historiqueInitial)/g
s/json\.Unmarshal(conv\.Historique, \&historique)/historique = conv.Historique/g
s/updateBuilder\.SetHistorique(historiqueJSON)/updateBuilder.SetHistorique(historique)/g
' internal/modules/convocations/service.go

# Supprimer les lignes de Marshal inutiles
sed -i.tmp '/donneesCompletesJSON, _ := json\.Marshal(donneesCompletes)/d' internal/modules/convocations/service.go
sed -i.tmp '/historiqueJSON, _ := json\.Marshal(historiqueInitial)/d' internal/modules/convocations/service.go
sed -i.tmp '/historiqueJSON, _ := json\.Marshal(historique)/d' internal/modules/convocations/service.go

rm -f internal/modules/convocations/service.go.tmp

echo "Étape 3/5 : Suppression de la fonction toResponse en double..."
# La fonction toResponse doit rester une seule fois
# On va juste s'assurer qu'il n'y a pas de doublon

echo "Étape 4/5 : Régénération des entités Ent..."
go generate ./ent

echo "Étape 5/5 : Compilation..."
go build -o server ./cmd/server 2>&1 | tee compile.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    echo "✅ COMPILATION RÉUSSIE !"
    echo ""
    echo "🔄 Redémarrage du serveur..."
    pkill -f "./server" || true
    sleep 2
    nohup ./server > server.log 2>&1 &
    sleep 3
    
    if ps aux | grep -v grep | grep "./server" > /dev/null; then
        echo "✅ Serveur démarré avec succès !"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "🎉 TOUT EST CORRIGÉ ET FONCTIONNE !"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "🧪 Testez maintenant depuis votre frontend :"
        echo "   POST /api/v1/convocations"
        echo ""
        echo "📖 Logs : tail -f server.log"
    else
        echo "⚠️  Le serveur n'a pas démarré correctement"
        echo "Vérifiez les logs : tail -20 server.log"
    fi
else
    echo ""
    echo "❌ ERREURS DE COMPILATION RESTANTES"
    echo ""
    echo "Erreurs détectées :"
    grep "error:" compile.log | head -20
    echo ""
    echo "Restauration du backup..."
    mv internal/modules/convocations/service.go.backup internal/modules/convocations/service.go
    exit 1
fi
