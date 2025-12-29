#!/usr/bin/env python3
"""
Script de correction des erreurs de syntaxe dans service.go
Ajoute des sauts de ligne manquants
"""

# Lire le fichier
with open('internal/modules/convocations/service.go', 'r') as f:
    lines = f.readlines()

# Corriger ligne par ligne
corrected_lines = []
for i, line in enumerate(lines):
    # Si la ligne se termine par } et la suivante commence par createBuilder ou updateBuilder
    if line.rstrip().endswith('}'):
        corrected_lines.append(line)
        # Vérifier la ligne suivante
        if i + 1 < len(lines):
            next_line = lines[i + 1]
            # Si la ligne suivante commence par une tabulation suivie de createBuilder ou updateBuilder
            # mais qu'il n'y a pas de ligne vide entre les deux
            if (next_line.strip().startswith('createBuilder') or 
                next_line.strip().startswith('updateBuilder')):
                # Vérifier qu'on n'a pas déjà ajouté de ligne vide
                if not line.strip() == '}':
                    continue  # La ligne } est déjà seule
    else:
        corrected_lines.append(line)

# En fait, approche plus simple : remplacer directement les patterns problématiques
with open('internal/modules/convocations/service.go', 'r') as f:
    content = f.read()

# Correction 1: Ligne ~427
content = content.replace(
    '}\tcreateBuilder.SetDonneesCompletes(donneesCompletes)',
    '}\n\tcreateBuilder.SetDonneesCompletes(donneesCompletes)'
)

# Correction 2: Ligne ~438
content = content.replace(
    '}\tcreateBuilder.SetHistorique(historiqueInitial)',
    '}\n\tcreateBuilder.SetHistorique(historiqueInitial)'
)

# Correction 3: Ligne ~573
content = content.replace(
    '}\tupdateBuilder.SetHistorique(historique)',
    '}\n\tupdateBuilder.SetHistorique(historique)'
)

# Variantes possibles avec espaces au lieu de tabs
content = content.replace(
    '}    createBuilder.SetDonneesCompletes(donneesCompletes)',
    '}\n    createBuilder.SetDonneesCompletes(donneesCompletes)'
)

content = content.replace(
    '}    createBuilder.SetHistorique(historiqueInitial)',
    '}\n    createBuilder.SetHistorique(historiqueInitial)'
)

content = content.replace(
    '}    updateBuilder.SetHistorique(historique)',
    '}\n    updateBuilder.SetHistorique(historique)'
)

# Sauvegarder
with open('internal/modules/convocations/service.go', 'w') as f:
    f.write(content)

print("✅ Fichier corrigé!")
print("📦 Relancez: make run")
