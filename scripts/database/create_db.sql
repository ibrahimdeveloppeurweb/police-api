-- Script pour créer la base de données PostgreSQL pour l'API Police Traffic

-- Créer la base de données
CREATE DATABASE police_traffic;

-- Se connecter à la base
\c police_traffic;

-- Créer un utilisateur pour l'application (optionnel)
-- CREATE USER police_api WITH PASSWORD 'police_api_password';
-- GRANT ALL PRIVILEGES ON DATABASE police_traffic TO police_api;

-- Vérification
SELECT 'Base de données police_traffic créée avec succès!' as message;