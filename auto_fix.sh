#!/bin/bash

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Correction automatique des erreurs de syntaxe..."

# Lire le fichier ligne par ligne et corriger
awk '
/^}	createBuilder\.SetDonneesCompletes/ {
    print "}"
    print "\tcreateBuild" substr($0, 3)
    next
}
/^}	createBuilder\.SetHistorique/ {
    print "}"
    print "\tcreateBuilder" substr($0, 3)
    next
}
/^}	updateBuilder\.SetHistorique/ {
    print "}"
    print "\tupdateBuilder" substr($0, 3)
    next
}
{ print }
' internal/modules/convocations/service.go > internal/modules/convocations/service_fixed.go

mv internal/modules/convocations/service_fixed.go internal/modules/convocations/service.go

echo "✅ Fichier corrigé!"
echo "📦 Compilation..."
make run
