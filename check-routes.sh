#!/bin/bash

echo "🔍 Vérification des routes objets-perdus..."
echo ""

# Vérifier que le serveur est en cours d'exécution
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Le serveur n'est pas en cours d'exécution sur le port 8080"
    echo "   Veuillez démarrer le serveur avec: go run ./cmd/server"
    exit 1
fi

echo "✅ Serveur en cours d'exécution"
echo ""

# Tester l'endpoint objets-perdus
echo "📡 Test de l'endpoint POST /api/objets-perdus..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/objets-perdus \
  -H "Content-Type: application/json" \
  -d '{}')

if [ "$response" = "401" ]; then
    echo "✅ Route trouvée ! (401 = Non autorisé, ce qui est normal sans token)"
    echo "   L'endpoint fonctionne, il nécessite juste une authentification"
elif [ "$response" = "404" ]; then
    echo "❌ Route non trouvée (404)"
    echo "   Le module objets-perdus n'est probablement pas chargé"
    echo "   Veuillez redémarrer le serveur avec: go run ./cmd/server"
    exit 1
else
    echo "⚠️  Réponse inattendue: $response"
fi

echo ""
echo "✅ Vérification terminée"

