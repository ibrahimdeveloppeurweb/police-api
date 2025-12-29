#!/bin/bash

# Script de correction des erreurs de types dans service.go

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Correction des erreurs de types dans service.go..."

# Créer un fichier de correction
cat > /tmp/fix_service.sed << 'EOF'
# Ligne 430: SetDonneesCompletes attend []byte pas map
s/createBuilder\.SetDonneesCompletes(donneesCompletesJSON)/createBuilder.SetDonneesCompletes(donneesCompletes)/g

# Ligne 443: SetHistorique attend []byte pas []map
s/createBuilder\.SetHistorique(historiqueJSON)/createBuilder.SetHistorique(historiqueInitial)/g

# Ligne 565: Historique est déjà []map pas []byte
s/json\.Unmarshal(conv\.Historique, &historique)/historique = conv.Historique/g

# Ligne 580: SetHistorique attend []byte pas []map  
s/updateBuilder\.SetHistorique(historiqueJSON)/updateBuilder.SetHistorique(historique)/g
EOF

# Appliquer les corrections
sed -i.bak -f /tmp/fix_service.sed internal/modules/convocations/service.go

echo "✅ Corrections appliquées"
echo "📦 Compilation..."

go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie!"
    echo "🔄 Redémarrage..."
    pkill -f "./server" || true
    sleep 2
    nohup ./server > server.log 2>&1 &
    sleep 3
    echo "✅ Serveur démarré!"
else
    echo "❌ Il reste des erreurs"
    echo "Restauration du backup..."
    mv internal/modules/convocations/service.go.bak internal/modules/convocations/service.go
fi
