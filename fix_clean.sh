#!/bin/bash

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Restauration et correction propre..."

# Si on a un backup, le restaurer
if [ -f internal/modules/convocations/service.go.backup ]; then
    cp internal/modules/convocations/service.go.backup internal/modules/convocations/service.go
fi

# Faire les corrections proprement avec sed
sed -i.bak \
    -e 's/SetDonneesCompletes(donneesCompletesJSON)/SetDonneesCompletes(donneesCompletes)/g' \
    -e '/donneesCompletesJSON, _ := json\.Marshal(donneesCompletes)/d' \
    -e 's/SetHistorique(historiqueJSON)/SetHistorique(historiqueInitial)/1' \
    -e '/historiqueJSON, _ := json\.Marshal(historiqueInitial)/d' \
    -e 's/json\.Unmarshal(conv\.Historique, \&historique)/\/\/ historique = conv.Historique (déjà le bon type)/' \
    -e 's/updateBuilder\.SetHistorique(historiqueJSON)/updateBuilder.SetHistorique(historique)/' \
    -e '/historiqueJSON, _ := json\.Marshal(historique)/d' \
    internal/modules/convocations/service.go

# Ajouter l'assignation historique après la déclaration
sed -i.bak2 '/var historique \[\]map\[string\]interface{}/a\
	if len(conv.Historique) > 0 {\
		historique = conv.Historique\
	}
' internal/modules/convocations/service.go

# Supprimer les lignes commentées
sed -i.bak3 '/\/\/ historique = conv.Historique/d' internal/modules/convocations/service.go

echo "✅ Corrections appliquées"
echo "📦 Compilation..."

go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie!"
    pkill -f "./server" || true
    sleep 2
    nohup ./server > server.log 2>&1 &
    sleep 3
    echo "✅ Serveur démarré!"
    echo "🧪 Testez : POST /api/v1/convocations"
else
    echo "❌ Erreurs restantes"
    go build -o server ./cmd/server 2>&1 | grep "error:" | head -10
fi
