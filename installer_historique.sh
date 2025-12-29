#!/bin/bash

set -e

echo "🚀 Installation automatique de l'historique des actions pour les plaintes"
echo "=========================================================================="
echo ""

BASE_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"
cd "$BASE_DIR"

# Étape 1: Génération des entités
echo "📝 Étape 1/4 : Génération des entités Ent..."
go generate ./ent
echo "✅ Entités générées"
echo ""

# Étape 2: Vérification
echo "🔍 Étape 2/4 : Vérification des fichiers générés..."
if [ -f "ent/historiqueactionplainte.go" ]; then
    echo "✅ Fichier ent/historiqueactionplainte.go créé"
else
    echo "⚠️  Fichier ent/historiqueactionplainte.go non trouvé"
fi
echo ""

# Étape 3: Afficher les instructions pour les modifications manuelles
echo "📋 Étape 3/4 : Modifications manuelles nécessaires"
echo ""
echo "Veuillez suivre le guide GUIDE_HISTORIQUE_ACTIONS_BACKEND.md pour :"
echo "  1. Ajouter les types dans types.go"
echo "  2. Ajouter les méthodes dans service_extended.go"
echo "  3. Modifier le contrôleur dans controller.go"
echo "  4. Modifier les endpoints existants"
echo ""
echo "Une fois les modifications faites, appuyez sur ENTRÉE pour continuer..."
read

# Étape 4: Compilation et redémarrage
echo "🔨 Étape 4/4 : Compilation et redémarrage du backend..."
echo ""

# Compiler
echo "Compilation..."
go build -o server cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie"
    echo ""
    echo "Pour démarrer le serveur :"
    echo "  ./server"
    echo ""
    echo "Puis testez avec :"
    echo "  curl http://localhost:8080/api/plaintes/{ID}/historique"
    echo ""
    echo "✅ Installation terminée !"
else
    echo "❌ Erreur de compilation"
    echo "Vérifiez les logs ci-dessus pour plus de détails"
    exit 1
fi
