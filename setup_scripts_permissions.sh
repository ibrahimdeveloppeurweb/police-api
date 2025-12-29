#!/bin/bash

# Script pour rendre tous les scripts exécutables
# Module Convocations - 74 champs

echo "🔧 Configuration des permissions pour les scripts..."
echo ""

cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Rendre les scripts exécutables
chmod +x regenerer_convocations_74_champs.sh
chmod +x test_api_convocations_74_champs.sh
chmod +x restart-backend.sh

echo "✅ Permissions configurées :"
echo "   - regenerer_convocations_74_champs.sh"
echo "   - test_api_convocations_74_champs.sh"
echo "   - restart-backend.sh"
echo ""
echo "🎯 Vous pouvez maintenant exécuter :"
echo "   ./regenerer_convocations_74_champs.sh"
echo "   ./restart-backend.sh"
echo "   ./test_api_convocations_74_champs.sh"
