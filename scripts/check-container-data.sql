-- Script pour vérifier les données de container_details dans la base
-- Usage: psql -h localhost -U postgres -d police_traffic -f check-container-data.sql

\echo '🔍 Vérification des données container_details'
\echo ''

-- 1. Afficher tous les objets avec is_container
\echo '1️⃣ Objets avec is_container = true:'
SELECT 
    id,
    numero,
    type_objet,
    is_container,
    container_details::text as container_details_raw
FROM objets_perdus
WHERE is_container = true
LIMIT 5;

\echo ''
\echo '2️⃣ Détail du container_details pour l''objet 7fa3287c-dd02-40d7-b650-47e9d7d8d296:'
SELECT 
    id,
    numero,
    type_objet,
    is_container,
    jsonb_pretty(container_details) as container_details_formatted
FROM objets_perdus
WHERE id = '7fa3287c-dd02-40d7-b650-47e9d7d8d296';

\echo ''
\echo '3️⃣ Statistiques sur les contenants:'
SELECT 
    COUNT(*) FILTER (WHERE is_container = true) as contenants,
    COUNT(*) FILTER (WHERE is_container = false) as objets_simples,
    COUNT(*) as total
FROM objets_perdus;

\echo ''
\echo '4️⃣ Types de contenants:'
SELECT 
    container_details->>'type' as type_contenant,
    COUNT(*) as nombre
FROM objets_perdus
WHERE is_container = true
GROUP BY container_details->>'type'
ORDER BY nombre DESC;
