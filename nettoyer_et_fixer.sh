#!/bin/bash

echo "🧹 Nettoyage et Application de la Solution PlainteHistorique"
echo "============================================================="
echo ""

BASE_DIR="/Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned"
cd "$BASE_DIR"

# Étape 1: Supprimer le fichier historique_action_plainte.go si existe
echo "📋 Étape 1: Nettoyage des fichiers inutiles"
echo ""

if [ -f "ent/schema/historique_action_plainte.go" ]; then
    echo "Suppression de historique_action_plainte.go..."
    rm "ent/schema/historique_action_plainte.go"
    echo "✅ Fichier supprimé"
else
    echo "✅ Fichier déjà absent"
fi
echo ""

# Étape 2: Retirer l'edge de plainte.go
echo "📋 Étape 2: Vérification du fichier plainte.go"
echo ""

if grep -q "historique_actions" "ent/schema/plainte.go"; then
    echo "⚠️  L'edge historique_actions existe dans plainte.go"
    echo "Création d'une sauvegarde..."
    cp "ent/schema/plainte.go" "ent/schema/plainte.go.bak"
    
    # Supprimer la ligne
    sed -i.tmp '/historique_actions/d' "ent/schema/plainte.go"
    rm "ent/schema/plainte.go.tmp"
    
    echo "✅ Edge supprimé (sauvegarde: plainte.go.bak)"
else
    echo "✅ Pas d'edge historique_actions à supprimer"
fi
echo ""

# Étape 3: Vérifier service_extended.go
echo "📋 Étape 3: Vérification du service"
echo ""

SERVICE_FILE="internal/modules/plainte/service_extended.go"

if grep -q "GetHistorique" "$SERVICE_FILE"; then
    echo "✅ La méthode GetHistorique existe déjà"
    
    # Vérifier si elle retourne bien un tableau
    if grep -q "return \[\]HistoriqueResponse{}" "$SERVICE_FILE"; then
        echo "✅ La méthode retourne déjà un tableau vide en cas d'erreur"
    else
        echo "⚠️  La méthode pourrait retourner nil"
        echo ""
        echo "📝 Modification recommandée dans GetHistorique:"
        echo ""
        cat << 'EOF'
// Remplacer tous les "return nil, err" par "return []HistoriqueResponse{}, nil"
// Exemple:

if err != nil {
    s.logger.Error("Failed to query plainte", zap.Error(err))
    return []HistoriqueResponse{}, nil  // Au lieu de: return nil, err
}
EOF
    fi
else
    echo "⚠️  La méthode GetHistorique n'existe pas"
    echo "Suivez le guide SOLUTION_AVEC_PLAINTE_HISTORIQUE.md pour l'ajouter"
fi
echo ""

# Étape 4: Vérifier le contrôleur
echo "📋 Étape 4: Vérification du contrôleur"
echo ""

CONTROLLER_FILE="internal/modules/plainte/controller.go"

if grep -q "GetHistorique" "$CONTROLLER_FILE"; then
    echo "✅ Le contrôleur a une méthode GetHistorique"
    
    # Vérifier qu'elle appelle bien le service
    if grep -q "c.service.GetHistorique" "$CONTROLLER_FILE"; then
        echo "✅ Le contrôleur appelle le service correctement"
    else
        echo "⚠️  Le contrôleur n'appelle peut-être pas le service"
    fi
else
    echo "⚠️  Pas de méthode GetHistorique dans le contrôleur"
fi
echo ""

# Étape 5: Compilation
echo "📋 Étape 5: Test de compilation"
echo ""

echo "Compilation du projet..."
go build -o server cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie"
    echo ""
    echo "========================================="
    echo "✅ NETTOYAGE TERMINÉ"
    echo "========================================="
    echo ""
    echo "📝 Prochaines étapes:"
    echo ""
    echo "1. Vérifiez que GetHistorique retourne []HistoriqueResponse{} en cas d'erreur"
    echo "2. Ajoutez CreateHistorique dans les méthodes (voir SOLUTION_AVEC_PLAINTE_HISTORIQUE.md)"
    echo "3. Testez l'endpoint:"
    echo "   curl http://localhost:8080/api/plaintes/UUID/historique"
    echo ""
    echo "Pour démarrer le serveur:"
    echo "   ./server"
else
    echo "❌ Erreur de compilation"
    echo ""
    echo "Vérifiez les erreurs ci-dessus"
    echo "Si l'entité HistoriqueActionPlainte est référencée, relancez:"
    echo "   go generate ./ent"
    exit 1
fi
