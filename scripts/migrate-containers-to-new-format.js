#!/usr/bin/env node

/**
 * Script de migration pour convertir les objets perdus de type "contenant" en mode contenant
 * Usage: node migrate-containers-to-new-format.js
 */

const axios = require('axios');

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';

// Types d'objets qui devraient être des contenants
const CONTAINER_TYPES_MAPPING = {
  'Sac / Sacoche': 'sac',
  'Sac à main': 'sac',
  'Sac de voyage': 'sac',
  'Sac de sport': 'sac',
  'Valise': 'valise',
  'Portefeuille': 'portefeuille',
  'Porte-monnaie': 'portefeuille',
  'Mallette': 'mallette',
  'Sac à dos': 'sac_dos',
  'Porte-documents': 'mallette',
};

async function main() {
  try {
    console.log('🚀 Démarrage de la migration des contenants...\n');

    // 1. Récupérer tous les objets perdus
    console.log('📥 Récupération des objets perdus...');
    const response = await axios.get(`${API_BASE_URL}/objets-perdus`, {
      params: {
        limit: 1000,
      },
    });

    const objets = response.data?.data?.objets || response.data?.objets || [];
    console.log(`✅ ${objets.length} objets perdus récupérés\n`);

    // 2. Filtrer les objets qui devraient être des contenants
    const objetsAMigrer = objets.filter((objet) => {
      const typeObjet = objet.typeObjet;
      const isContainer = objet.isContainer || false;

      // Si déjà un contenant, on ne migre pas
      if (isContainer) {
        return false;
      }

      // Vérifier si le type correspond à un contenant
      return Object.keys(CONTAINER_TYPES_MAPPING).some((type) =>
        typeObjet.includes(type)
      );
    });

    console.log(`🎯 ${objetsAMigrer.length} objets à migrer\n`);

    if (objetsAMigrer.length === 0) {
      console.log('✨ Aucun objet à migrer. Migration terminée !');
      return;
    }

    // 3. Migrer chaque objet
    let success = 0;
    let errors = 0;

    for (const objet of objetsAMigrer) {
      try {
        // Déterminer le type de contenant
        let containerType = 'sac'; // Par défaut
        for (const [typeKey, typeValue] of Object.entries(
          CONTAINER_TYPES_MAPPING
        )) {
          if (objet.typeObjet.includes(typeKey)) {
            containerType = typeValue;
            break;
          }
        }

        // Préparer les détails du contenant
        const containerDetails = {
          type: containerType,
          couleur: objet.couleur || undefined,
          marque: objet.detailsSpecifiques?.marque || undefined,
          taille: undefined,
          signesDistinctifs: objet.description || undefined,
          inventory: [], // Inventaire vide par défaut
        };

        // Préparer la requête de mise à jour
        const updateData = {
          isContainer: true,
          containerDetails,
        };

        // Mettre à jour l'objet
        await axios.patch(
          `${API_BASE_URL}/objets-perdus/${objet.id}`,
          updateData
        );

        success++;
        console.log(
          `✅ [${success}/${objetsAMigrer.length}] Migré: ${objet.numero} - ${objet.typeObjet} → ${containerType}`
        );
      } catch (error) {
        errors++;
        console.error(
          `❌ Erreur lors de la migration de ${objet.numero}:`,
          error.response?.data || error.message
        );
      }
    }

    // 4. Résumé
    console.log('\n' + '='.repeat(60));
    console.log('📊 RÉSUMÉ DE LA MIGRATION');
    console.log('='.repeat(60));
    console.log(`✅ Objets migrés avec succès: ${success}`);
    console.log(`❌ Erreurs: ${errors}`);
    console.log(`📦 Total traité: ${objetsAMigrer.length}`);
    console.log('='.repeat(60));

    if (errors > 0) {
      process.exit(1);
    }
  } catch (error) {
    console.error('❌ Erreur fatale:', error.message);
    if (error.response) {
      console.error('Réponse API:', error.response.data);
    }
    process.exit(1);
  }
}

// Exécuter le script
main().catch((error) => {
  console.error('❌ Erreur non gérée:', error);
  process.exit(1);
});
