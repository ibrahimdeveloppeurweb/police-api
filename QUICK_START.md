# Guide de Démarrage Rapide

## 🚀 Installation et Configuration

### 1. Copier le schéma Ent

```bash
cd /Users/mat/Development/importants/police-traffic-back-front
cp -r police-trafic-api/ent police-trafic-api-frontend-aligned/
```

### 2. Installer les dépendances

```bash
cd police-trafic-api-frontend-aligned
go mod download
go mod tidy
```

### 3. Configurer la base de données

Éditer `config/config.yaml` :

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "votre_mot_de_passe"
  dbname: "police_traffic"
  sslmode: "disable"
```

### 4. Lancer l'application

```bash
make run
# ou
go run cmd/server/main.go
```

L'API sera disponible sur `http://localhost:8080`

## 📡 Endpoints Disponibles

### Authentification
- `POST /api/v1/auth/login` - Connexion
- `GET /api/v1/auth/me` - Utilisateur actuel
- `POST /api/v1/auth/logout` - Déconnexion
- `POST /api/v1/auth/refresh` - Rafraîchir token

### Contrôles
- `GET /api/v1/controles` - Liste
- `GET /api/v1/controles/:id` - Détails
- `POST /api/v1/controles` - Créer
- `PUT /api/v1/controles/:id` - Mettre à jour
- `DELETE /api/v1/controles/:id` - Supprimer
- `POST /api/v1/controles/:id/pv` - Générer PV

### PV
- `GET /api/v1/pv` - Liste
- `GET /api/v1/pv/:id` - Détails
- `PATCH /api/v1/pv/:id/paiement` - Mettre à jour paiement

### Admin
- `GET /api/v1/admin/statistiques` - Statistiques nationales
- `GET /api/v1/admin/commissariats` - Liste commissariats
- `GET /api/v1/admin/commissariats/:id` - Détails commissariat
- `GET /api/v1/admin/agents` - Liste agents

### Alertes
- `GET /api/v1/alertes` - Liste
- `GET /api/v1/alertes/:id` - Détails
- `POST /api/v1/alertes` - Créer
- `PUT /api/v1/alertes/:id` - Mettre à jour
- `PATCH /api/v1/alertes/:id/resolve` - Résoudre

### Commissariat
- `GET /api/v1/commissariat/:id/dashboard` - Dashboard
- `GET /api/v1/commissariat/:id/agents` - Agents
- `GET /api/v1/commissariat/:id/statistiques` - Statistiques

## 🔍 Vérification

### Health Check
```bash
curl http://localhost:8080/health
```

### Swagger Documentation
Ouvrir dans le navigateur :
```
http://localhost:8080/swagger/index.html
```

## ⚠️ Notes Importantes

1. **Schéma Ent** : Le dossier `ent/` doit être copié depuis le projet principal
2. **Base de données** : PostgreSQL doit être configuré et accessible
3. **Authentification** : Le module auth est basique, à compléter selon vos besoins
4. **Migrations** : Les migrations Ent s'exécutent automatiquement au démarrage

## 🐛 Dépannage

### Erreur : "ent package not found"
```bash
# Copier le dossier ent
cp -r ../police-trafic-api/ent .
```

### Erreur : "database connection failed"
Vérifier la configuration dans `config/config.yaml` et que PostgreSQL est démarré.

### Erreur : "module not found"
```bash
go mod download
go mod tidy
```




