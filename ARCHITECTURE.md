# Architecture du Projet

## 📐 Structure Modulaire

Le projet suit une architecture modulaire avec séparation claire des responsabilités :

```
┌─────────────────────────────────────────┐
│           Application Layer             │
│         (cmd/server/main.go)            │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│            App Configuration             │
│          (internal/app/app.go)           │
│  - Dependency Injection (Fx)            │
│  - Module Registration                  │
└─────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
┌──────────────┐      ┌──────────────────┐
│  Core Layer  │      │ Infrastructure   │
│              │      │      Layer        │
│ - Router     │      │ - Config         │
│ - Server     │      │ - Database        │
│ - Interfaces │      │ - Logger          │
└──────────────┘      └──────────────────┘
        │
        ▼
┌─────────────────────────────────────────┐
│         Business Modules Layer          │
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌────────┐│
│  │ Controles│  │    PV    │  │ Admin  ││
│  └──────────┘  └──────────┘  └────────┘│
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌────────┐│
│  │ Alertes  │  │Commissar.│  │  Auth  ││
│  └──────────┘  └──────────┘  └────────┘│
└─────────────────────────────────────────┘
```

## 🏗️ Architecture d'un Module

Chaque module suit la même structure :

```
module/
├── dto.go          # Data Transfer Objects (alignés frontend)
├── repository.go   # Accès aux données (Ent ORM)
├── service.go      # Logique métier
├── controller.go   # Endpoints HTTP (Echo)
└── module.go       # Configuration Fx
```

### Flux de Données

```
HTTP Request
    │
    ▼
Controller (controller.go)
    │ - Validation
    │ - Binding
    │
    ▼
Service (service.go)
    │ - Logique métier
    │ - Transformation DTO
    │
    ▼
Repository (repository.go)
    │ - Requêtes Ent
    │ - Mapping Ent ↔ DTO
    │
    ▼
Database (PostgreSQL via Ent)
```

## 🔄 Pattern DTO

Tous les modules utilisent des DTOs alignés avec le frontend :

```go
// Frontend TypeScript
interface Controle {
  id: string;
  numero: string;
  type: TypeControle;
  // ...
}

// Backend Go DTO
type ControleResponseDTO struct {
    ID     string      `json:"id"`
    Numero string      `json:"numero"`
    Type   TypeControle `json:"type"`
    // ...
}
```

## 📦 Injection de Dépendances (Fx)

Chaque module est enregistré via Fx :

```go
var Module = fx.Module("controles",
    fx.Provide(
        NewRepository,    // Repository avec DB
        NewService,       // Service avec Repository
        fx.Annotate(
            NewController, // Controller avec Service
            fx.As(new(interfaces.Controller)),
            fx.ResultTags(`group:"controllers"`),
        ),
    ),
)
```

## 🔌 Interfaces

Tous les controllers implémentent `interfaces.Controller` :

```go
type Controller interface {
    RegisterRoutes(router *echo.Group, middleware ...echo.MiddlewareFunc)
    GetPrefix() string
    GetVersion() string
}
```

## 🗄️ Base de Données

- **ORM** : Ent (entgo.io/ent)
- **Database** : PostgreSQL
- **Migrations** : Automatiques au démarrage

## 📝 Validation

- **Validator** : go-playground/validator
- **Middleware** : Intégré dans Echo
- **Validation** : Tags struct dans les DTOs

## 🎯 Principes

1. **Séparation des responsabilités** : Chaque couche a un rôle précis
2. **DTOs alignés** : Correspondance exacte avec le frontend
3. **Modularité** : Chaque module est indépendant
4. **Réutilisabilité** : Services partagés via Fx
5. **Maintenabilité** : Structure claire et documentée




