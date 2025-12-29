#!/bin/bash

# 🚀 SCRIPT D'INTÉGRATION - Mode Contenant avec Inventaire
# Ce script applique toutes les modifications nécessaires au backend

echo "======================================"
echo "🚀 INTÉGRATION MODE CONTENANT"
echo "======================================"
echo ""

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages
print_step() {
    echo ""
    echo -e "${YELLOW}📌 ÉTAPE $1: $2${NC}"
    echo "--------------------------------------"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Se placer dans le répertoire du backend
cd /Users/ibrahim/Documents/police1/police-trafic-api-frontend-aligned

# Étape 1 : Vérifier que les modifications sont présentes
print_step "1" "Vérification des fichiers modifiés"

if grep -q "is_container" ent/schema/objet_perdu.go; then
    print_success "Schéma Ent modifié"
else
    print_error "Le schéma Ent n'a pas été modifié"
    exit 1
fi

if grep -q "InventoryItem" internal/modules/objets-perdus/types.go; then
    print_success "Types modifiés"
else
    print_error "Les types n'ont pas été modifiés"
    exit 1
fi

if grep -q "IsContainer" internal/infrastructure/repository/objet_perdu_repository.go; then
    print_success "Repository modifié"
else
    print_error "Le repository n'a pas été modifié"
    exit 1
fi

if grep -q "containerDetails" internal/modules/objets-perdus/service.go; then
    print_success "Service modifié"
else
    print_error "Le service n'a pas été modifié"
    exit 1
fi

# Étape 2 : Régénérer le code Ent
print_step "2" "Régénération du code Ent"

if go generate ./ent; then
    print_success "Code Ent régénéré avec succès"
else
    print_error "Erreur lors de la régénération du code Ent"
    exit 1
fi

# Étape 3 : Nettoyer les dépendances
print_step "3" "Nettoyage des dépendances Go"

go mod tidy
print_success "Dépendances nettoyées"

# Étape 4 : Compiler le code
print_step "4" "Compilation du code"

if go build ./...; then
    print_success "Compilation réussie"
else
    print_error "Erreur de compilation"
    exit 1
fi

# Étape 5 : Créer et appliquer la migration
print_step "5" "Création et application de la migration"

echo "⚠️  Attention: Cette étape va modifier la base de données"
read -p "Voulez-vous continuer? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if go run cmd/migrate/main.go; then
        print_success "Migration appliquée avec succès"
    else
        print_error "Erreur lors de la migration"
        exit 1
    fi
else
    echo "Migration annulée"
    echo "⚠️  Vous devrez exécuter manuellement: go run cmd/migrate/main.go"
fi

# Étape 6 : Vérification finale
print_step "6" "Vérification finale"

echo "Vérification de la structure de la base de données..."
psql -U postgres -d police_trafic_db -c "\d objets_perdus" > /tmp/db_check.txt 2>&1

if grep -q "is_container" /tmp/db_check.txt; then
    print_success "Colonne 'is_container' trouvée dans la base de données"
else
    print_error "Colonne 'is_container' NON trouvée - La migration n'a peut-être pas été appliquée"
fi

if grep -q "container_details" /tmp/db_check.txt; then
    print_success "Colonne 'container_details' trouvée dans la base de données"
else
    print_error "Colonne 'container_details' NON trouvée - La migration n'a peut-être pas été appliquée"
fi

rm /tmp/db_check.txt

# Récapitulatif
echo ""
echo "======================================"
echo "✅ INTÉGRATION TERMINÉE"
echo "======================================"
echo ""
echo "📋 Résumé:"
echo "  ✅ Fichiers modifiés vérifiés"
echo "  ✅ Code Ent régénéré"
echo "  ✅ Dépendances nettoyées"
echo "  ✅ Code compilé"
echo "  ✅ Migration appliquée (si confirmée)"
echo "  ✅ Base de données vérifiée"
echo ""
echo "🚀 Prochaines étapes:"
echo ""
echo "  1. Démarrer le serveur:"
echo "     go run cmd/server/main.go"
echo ""
echo "  2. Tester avec cURL (objet simple):"
echo "     curl -X POST http://localhost:8080/api/objets-perdus \\"
echo "       -H 'Content-Type: application/json' \\"
echo "       -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "       -d '{...}'"
echo ""
echo "  3. Tester avec le frontend:"
echo "     - Ouvrir http://localhost:3000/gestion/objets-perdus/nouveau"
echo "     - Créer un objet simple"
echo "     - Créer un contenant avec inventaire"
echo ""
echo "📚 Documentation:"
echo "  - GUIDE_INTEGRATION_INVENTAIRE_OBJETS_PERDUS.md"
echo "  - RESUME_MODIFICATIONS_INVENTAIRE.md"
echo ""
echo "======================================"
