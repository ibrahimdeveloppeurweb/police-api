-- Migration: Convertir les objets perdus de type contenant en mode contenant
-- Cette migration identifie les objets perdus qui devraient être des contenants
-- et met à jour leurs champs is_container et container_details

UPDATE objets_perdus
SET 
    is_container = true,
    container_details = jsonb_build_object(
        'type', 
        CASE 
            WHEN type_objet ILIKE '%sac%sacoche%' THEN 'sac'
            WHEN type_objet ILIKE '%valise%' OR type_objet ILIKE '%bagage%' THEN 'valise'
            WHEN type_objet ILIKE '%portefeuille%' THEN 'portefeuille'
            WHEN type_objet ILIKE '%mallette%' THEN 'mallette'
            WHEN type_objet ILIKE '%sac%dos%' OR type_objet ILIKE '%sac à dos%' THEN 'sac_dos'
            ELSE 'sac'
        END,
        'couleur', couleur,
        'marque', COALESCE(details_specifiques->>'marque', ''),
        'taille', '',
        'signesDistinctifs', description,
        'inventory', '[]'::jsonb
    ),
    updated_at = NOW()
WHERE 
    (
        type_objet ILIKE '%sac%sacoche%' OR
        type_objet ILIKE '%valise%' OR
        type_objet ILIKE '%bagage%' OR
        type_objet ILIKE '%portefeuille%' OR
        type_objet ILIKE '%mallette%' OR
        type_objet ILIKE '%sac%dos%' OR
        type_objet ILIKE '%sac à dos%'
    )
    AND is_container = false;

-- Afficher le résultat
SELECT 
    id,
    numero,
    type_objet,
    is_container,
    container_details->'type' as container_type
FROM objets_perdus
WHERE is_container = true
ORDER BY created_at DESC;
