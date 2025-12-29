#!/bin/bash

# Script de correction et redémarrage du backend
# Police Trafic API - 10 Décembre 2025

clear

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                                                              ║"
echo "║        🔧 CORRECTION ET REDÉMARRAGE BACKEND                 ║"
echo "║        Police Trafic API                                     ║"
echo "║                                                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

# Vérifier qu'on est dans le bon répertoire
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ Erreur: Fichier go.mod non trouvé!${NC}"
    echo "   Veuillez exécuter ce script depuis la racine du projet backend."
    exit 1
fi

echo -e "${BLUE}▶ Étape 1/4 : Arrêt du serveur actuel...${NC}"
echo ""

# Arrêter le serveur s'il tourne
pkill -f "server" 2>/dev/null
pkill -f "go run ./cmd/api" 2>/dev/null
sleep 1

echo -e "${GREEN}✅ Serveur arrêté${NC}"
echo ""
sleep 1

echo -e "${BLUE}▶ Étape 2/4 : Nettoyage des anciens builds...${NC}"
echo ""

# Supprimer les anciens exécutables
rm -f server server-* 2>/dev/null

echo -e "${GREEN}✅ Nettoyage terminé${NC}"
echo ""
sleep 1

echo -e "${BLUE}▶ Étape 3/4 : Compilation du backend...${NC}"
echo ""

# Compiler le backend
go build -o server ./cmd/api

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erreur de compilation!${NC}"
    echo "   Vérifiez les erreurs ci-dessus."
    exit 1
fi

echo -e "${GREEN}✅ Backend compilé avec succès${NC}"
echo ""
sleep 1

echo -e "${BLUE}▶ Étape 4/4 : Démarrage du serveur...${NC}"
echo ""
echo -e "${YELLOW}Le serveur va démarrer. Appuyez sur Ctrl+C pour l'arrêter.${NC}"
echo ""
sleep 2

# Démarrer le serveur
./server
