#!/bin/bash

# Script de correction finale pour service.go

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned/internal/modules/convocations

echo "🔧 Correction de service.go..."

# Backup
cp service.go service.go.backup

# Correction des lignes problématiques
cat service.go | \
  # Ligne 430: SetDonneesCompletes attend map[string]interface{} directement
  sed 's/createBuilder\.SetDonneesCompletes(donneesCompletesJSON)/createBuilder.SetDonneesCompletes(donneesCompletes)/' | \
  # Ligne 443: SetHistorique attend []map[string]interface{} directement  
  sed 's/createBuilder\.SetHistorique(historiqueJSON)/createBuilder.SetHistorique(historiqueInitial)/' | \
  # Ligne 565: conv.Historique est déjà []map[string]interface{}
  sed 's/json\.Unmarshal(conv\.Historique, \&historique)/historique = conv.Historique/' | \
  # Ligne 580: SetHistorique attend []map[string]interface{} directement
  sed 's/updateBuilder\.SetHistorique(historiqueJSON)/updateBuilder.SetHistorique(historique)/' | \
  # Supprimer les lignes inutiles de Marshal
  sed '/donneesCompletesJSON, _ := json\.Marshal(donneesCompletes)/d' | \
  sed '/historiqueJSON, _ := json\.Marshal(historiqueInitial)/d' | \
  sed '/historiqueJSON, _ := json\.Marshal(historique)/d' \
  > service_fixed.go

mv service_fixed.go service.go

echo "✅ service.go corrigé"
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned
echo "📦 Compilation..."
make run
