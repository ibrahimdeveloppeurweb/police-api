#!/bin/bash

# Script de correction finale - Supprime les lignes problématiques

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Correction des erreurs de syntaxe..."

# Créer une version corrigée
cat internal/modules/convocations/service.go | \
  # Remplacer SetDonneesCompletes(donneesCompletesJSON) par SetDonneesCompletes(donneesCompletes)
  sed 's/\.SetDonneesCompletes(donneesCompletesJSON)/.SetDonneesCompletes(donneesCompletes)/g' | \
  # Remplacer SetHistorique(historiqueJSON) par SetHistorique(historiqueInitial) dans Create
  sed '0,/\.SetHistorique(historiqueJSON)/s/\.SetHistorique(historiqueJSON)/.SetHistorique(historiqueInitial)/' | \
  # Remplacer SetHistorique(historiqueJSON) par SetHistorique(historique) dans UpdateStatut
  sed 's/updateBuilder\.SetHistorique(historiqueJSON)/updateBuilder.SetHistorique(historique)/g' | \
  # Supprimer la ligne json.Unmarshal(conv.Historique
  sed '/json\.Unmarshal(conv\.Historique, &historique)/d' | \
  # Ajouter historique = conv.Historique à la place
  sed '/var historique \[\]map\[string\]interface{}/a\	historique = conv.Historique' | \
  # Supprimer les lignes Marshal inutiles
  grep -v 'donneesCompletesJSON, _ := json.Marshal(donneesCompletes)' | \
  grep -v 'historiqueJSON, _ := json.Marshal(historiqueInitial)' | \
  grep -v 'historiqueJSON, _ := json.Marshal(historique)' \
  > internal/modules/convocations/service_fixed.go

mv internal/modules/convocations/service_fixed.go internal/modules/convocations/service.go

echo "✅ Syntaxe corrigée"
echo "📦 Compilation..."
make run
