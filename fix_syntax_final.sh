#!/bin/bash

# Script de correction des erreurs de syntaxe - Version finale

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Correction des erreurs de syntaxe..."

# Sauvegarder l'original
cp internal/modules/convocations/service.go internal/modules/convocations/service.go.broken

# Correction avec sed - Ajouter des sauts de ligne manquants
sed -i.tmp '
# Ligne 427: Ajouter saut de ligne après }
s/}	createBuilder\.SetDonneesCompletes/}\
	createBuilder.SetDonneesCompletes/g

# Ligne 438: Ajouter saut de ligne après }  
s/}	createBuilder\.SetHistorique/}\
	createBuilder.SetHistorique/g

# Ligne 573: Ajouter saut de ligne après }
s/}	updateBuilder\.SetHistorique/}\
	updateBuilder.SetHistorique/g
' internal/modules/convocations/service.go

rm -f internal/modules/convocations/service.go.tmp

echo "✅ Corrections syntaxiques appliquées"
echo "📦 Compilation..."

go build -o server ./cmd/server 2>&1 | tee compile.log

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ COMPILATION RÉUSSIE!"
    echo ""
    pkill -f "./server" || true
    sleep 2
    nohup ./server > server.log 2>&1 &
    sleep 3
    echo "✅ Serveur démarré!"
    echo "🧪 Testez: POST /api/v1/convocations"
else
    echo ""
    echo "❌ Erreurs restantes:"
    grep "error:" compile.log | head -10
fi
