#!/usr/bin/env python3
"""
Script de correction automatique pour service.go
Corrige toutes les erreurs de types
"""

import re

# Lire le fichier
with open('internal/modules/convocations/service.go', 'r') as f:
    content = f.read()

# Correction 1: Ligne 430 - SetDonneesCompletes
content = content.replace(
    'createBuilder.SetDonneesCompletes(donneesCompletesJSON)',
    'createBuilder.SetDonneesCompletes(donneesCompletes)'
)

# Correction 2: Ligne 443 - SetHistorique  
content = content.replace(
    'createBuilder.SetHistorique(historiqueJSON)',
    'createBuilder.SetHistorique(historiqueInitial)'
)

# Correction 3: Ligne 565 - Historique est déjà []map
content = content.replace(
    'json.Unmarshal(conv.Historique, &historique)',
    'historique = conv.Historique'
)

# Correction 4: Ligne 580 - SetHistorique
content = content.replace(
    'updateBuilder.SetHistorique(historiqueJSON)',
    'updateBuilder.SetHistorique(historique)'
)

# Correction 5: Supprimer les lignes de Marshal inutiles
content = re.sub(r'\s*donneesCompletesJSON, _ := json\.Marshal\(donneesCompletes\)\n', '', content)
content = re.sub(r'\s*historiqueJSON, _ := json\.Marshal\(historiqueInitial\)\n', '', content)
content = re.sub(r'\s*historiqueJSON, _ := json\.Marshal\(historique\)\n', '', content)

# Correction 6: Ligne 640 - QualiteConvoque est *string
content = content.replace(
    'QualiteConvoque:   conv.QualiteConvoque,',
    '''QualiteConvoque:   func() string {
			if conv.QualiteConvoque != nil {
				return *conv.QualiteConvoque
			}
			return conv.StatutPersonne
		}(),'''
)

# Correction 7: Lignes 651-652 - ConvoqueEmail est *string
old_email = '''	if conv.ConvoqueEmail != "" {
		response.ConvoqueEmail = &conv.ConvoqueEmail
	}'''
new_email = '''	if conv.ConvoqueEmail != nil && *conv.ConvoqueEmail != "" {
		response.ConvoqueEmail = conv.ConvoqueEmail
	}'''
content = content.replace(old_email, new_email)

# Correction 8: Lignes 654-655 - ConvoqueAdresse est *string
old_adresse = '''	if conv.ConvoqueAdresse != "" {
		response.ConvoqueAdresse = &conv.ConvoqueAdresse
	}'''
new_adresse = '''	if conv.ConvoqueAdresse != nil && *conv.ConvoqueAdresse != "" {
		response.ConvoqueAdresse = conv.ConvoqueAdresse
	}'''
content = content.replace(old_adresse, new_adresse)

# Correction 9: Ligne 658 - DateRdv est *time.Time
old_date_rdv = '''	if !conv.DateRdv.IsZero() {
		response.DateRdv = &conv.DateRdv
	}'''
new_date_rdv = '''	if conv.DateRdv != nil && !conv.DateRdv.IsZero() {
		response.DateRdv = conv.DateRdv
	}'''
content = content.replace(old_date_rdv, new_date_rdv)

# Correction 10: HeureRdv est *string
old_heure = '''	if conv.HeureRdv != "" {
		response.HeureRdv = &conv.HeureRdv
	}'''
new_heure = '''	if conv.HeureRdv != nil && *conv.HeureRdv != "" {
		response.HeureRdv = conv.HeureRdv
	}'''
content = content.replace(old_heure, new_heure)

# Correction 11: DateEnvoi est *time.Time
old_date_envoi = '''	if !conv.DateEnvoi.IsZero() {
		response.DateEnvoi = &conv.DateEnvoi
	}'''
new_date_envoi = '''	if conv.DateEnvoi != nil && !conv.DateEnvoi.IsZero() {
		response.DateEnvoi = conv.DateEnvoi
	}'''
content = content.replace(old_date_envoi, new_date_envoi)

# Correction 12: DateHonoration est *time.Time
old_date_hon = '''	if !conv.DateHonoration.IsZero() {
		response.DateHonoration = &conv.DateHonoration
	}'''
new_date_hon = '''	if conv.DateHonoration != nil && !conv.DateHonoration.IsZero() {
		response.DateHonoration = conv.DateHonoration
	}'''
content = content.replace(old_date_hon, new_date_hon)

# Correction 13: Observations est *string
old_obs = '''	if conv.Observations != "" {
		response.Observations = &conv.Observations
	}'''
new_obs = '''	if conv.Observations != nil && *conv.Observations != "" {
		response.Observations = conv.Observations
	}'''
content = content.replace(old_obs, new_obs)

# Correction 14: ResultatAudition est *string
old_result = '''	if conv.ResultatAudition != "" {
		response.ResultatAudition = &conv.ResultatAudition
	}'''
new_result = '''	if conv.ResultatAudition != nil && *conv.ResultatAudition != "" {
		response.ResultatAudition = conv.ResultatAudition
	}'''
content = content.replace(old_result, new_result)

# Correction 15: AffaireLiee est *string
old_affaire = '''	if conv.AffaireLiee != "" {
		response.AffaireLiee = &conv.AffaireLiee
	}'''
new_affaire = '''	if conv.AffaireLiee != nil && *conv.AffaireLiee != "" {
		response.AffaireLiee = conv.AffaireLiee
	}'''
content = content.replace(old_affaire, new_affaire)

# Écrire le fichier corrigé
with open('internal/modules/convocations/service.go', 'w') as f:
    f.write(content)

print("✅ Fichier corrigé avec succès!")
print("📦 Lancez maintenant: make run")
