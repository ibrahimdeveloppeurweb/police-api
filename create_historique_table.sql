-- =====================================================
-- CRÉATION DE LA TABLE historique_action_plaintes
-- =====================================================

-- Supprimer la table si elle existe déjà
DROP TABLE IF EXISTS historique_action_plaintes CASCADE;

-- Créer la table
CREATE TABLE historique_action_plaintes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plainte_id UUID NOT NULL,
    type_action VARCHAR(50) NOT NULL,
    ancienne_valeur VARCHAR(255),
    nouvelle_valeur VARCHAR(255) NOT NULL,
    observations TEXT,
    effectue_par UUID,
    effectue_par_nom VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Clé étrangère vers la table plaintes
    CONSTRAINT fk_plainte
        FOREIGN KEY (plainte_id)
        REFERENCES plaintes(id)
        ON DELETE CASCADE,
    
    -- Clé étrangère vers la table users (optionnelle)
    CONSTRAINT fk_user
        FOREIGN KEY (effectue_par)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- Créer les index pour améliorer les performances
CREATE INDEX idx_historique_plainte_id ON historique_action_plaintes(plainte_id);
CREATE INDEX idx_historique_created_at ON historique_action_plaintes(created_at DESC);
CREATE INDEX idx_historique_type_action ON historique_action_plaintes(type_action);

-- Ajouter un commentaire
COMMENT ON TABLE historique_action_plaintes IS 'Historique détaillé des actions effectuées sur les plaintes';
COMMENT ON COLUMN historique_action_plaintes.type_action IS 'Type: CHANGEMENT_ETAPE, CHANGEMENT_STATUT, ASSIGNATION_AGENT, CONVOCATION';

-- =====================================================
-- AJOUT DES CHAMPS MANQUANTS DANS LA TABLE plaintes
-- =====================================================

-- Ajouter nombre_convocations si n'existe pas
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='plaintes' AND column_name='nombre_convocations') THEN
        ALTER TABLE plaintes ADD COLUMN nombre_convocations INTEGER DEFAULT 0;
    END IF;
END $$;

-- Ajouter decision_finale si n'existe pas
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='plaintes' AND column_name='decision_finale') THEN
        ALTER TABLE plaintes ADD COLUMN decision_finale TEXT;
    END IF;
END $$;

-- =====================================================
-- DONNÉES DE TEST (OPTIONNEL)
-- =====================================================

-- Exemple d'insertion
-- INSERT INTO historique_action_plaintes 
-- (plainte_id, type_action, ancienne_valeur, nouvelle_valeur, observations, effectue_par_nom)
-- VALUES 
-- ('VOTRE-UUID-PLAINTE', 'CHANGEMENT_ETAPE', 'DEPOT', 'ENQUETE', 'Début de l''enquête terrain', 'Jean Dupont');

-- Vérification
SELECT 
    COUNT(*) as total_actions,
    type_action,
    COUNT(*) as count_per_type
FROM historique_action_plaintes
GROUP BY type_action
ORDER BY count_per_type DESC;

-- =====================================================
-- FIN DU SCRIPT
-- =====================================================

\echo '✅ Table historique_action_plaintes créée avec succès !'
\echo '✅ Index créés'
\echo '✅ Champs ajoutés dans la table plaintes'
\echo ''
\echo 'Prochaines étapes :'
\echo '1. Redémarrer le backend'
\echo '2. Tester l''endpoint GET /plaintes/:id/historique'
