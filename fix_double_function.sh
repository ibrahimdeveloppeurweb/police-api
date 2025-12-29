#!/bin/bash

# Script de correction - Supprime la fonction toResponse en double

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

echo "🔧 Suppression de la fonction toResponse en double dans service.go..."

# Lire tout le fichier jusqu'à "// toResponse converts" et le sauvegarder
head -n $(grep -n "^// toResponse converts ent.Convocation to ConvocationResponse" internal/modules/convocations/service.go | tail -1 | cut -d: -f1 | awk '{print $1-1}') internal/modules/convocations/service.go > internal/modules/convocations/service_temp.go

# Remplacer
mv internal/modules/convocations/service_temp.go internal/modules/convocations/service.go

echo "✅ Fonction supprimée"
echo "📦 Compilation..."

make run
