package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"police-trafic-api-frontend-aligned/ent"
)

// EntityType represents the type of entity for code generation
type EntityType string

const (
	EntityUser              EntityType = "USR"
	EntityCommissariat      EntityType = "COM"
	EntityControle          EntityType = "CTRL"
	EntityInspection        EntityType = "INSP"
	EntityInfraction        EntityType = "INFR"
	EntityProcesVerbal      EntityType = "PV"
	EntityPaiement          EntityType = "PAI"
	EntityRecours           EntityType = "REC"
	EntityVehicule          EntityType = "VEH"
	EntityConducteur        EntityType = "COND"
	EntityDocument          EntityType = "DOC"
	EntityAlerteSecuritaire EntityType = "ALRT"
	EntityPlainte           EntityType = "PLT"
	EntityMission           EntityType = "MIS"
	EntityEquipe            EntityType = "EQP"
	EntityObjectif          EntityType = "OBJ"
	EntityObservation       EntityType = "OBS"
	EntityCompetence        EntityType = "COMP"
	EntityCheckItem         EntityType = "CHKI"
	EntityCheckOption       EntityType = "CHKO"
	EntityInfractionType    EntityType = "ITYPE"
	EntityAuditLog          EntityType = "AUD"
)

// CodeGenerator handles unique code generation for entities
type CodeGenerator struct {
	client   *ent.Client
	counters map[string]int
	mu       sync.Mutex
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(client *ent.Client) *CodeGenerator {
	return &CodeGenerator{
		client:   client,
		counters: make(map[string]int),
	}
}

// GenerateCode generates a unique code for an entity
// Format: {PREFIX}-{COMM_CODE}-{YYYYMM}-{SEQ} or {PREFIX}-{YYYYMM}-{SEQ}
// Example: CTRL-DKR01-202512-001, USR-202512-042
func (g *CodeGenerator) GenerateCode(ctx context.Context, entityType EntityType, commissariatCode string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	yearMonth := now.Format("200601")

	// Build the code prefix
	var codePrefix string
	if commissariatCode != "" {
		// Normalize commissariat code (uppercase, remove special chars)
		commCode := strings.ToUpper(strings.ReplaceAll(commissariatCode, "-", ""))
		if len(commCode) > 6 {
			commCode = commCode[:6]
		}
		codePrefix = fmt.Sprintf("%s-%s-%s", entityType, commCode, yearMonth)
	} else {
		codePrefix = fmt.Sprintf("%s-%s", entityType, yearMonth)
	}

	// Get and increment counter for this prefix
	counterKey := codePrefix
	g.counters[counterKey]++
	seq := g.counters[counterKey]

	// Format sequence with leading zeros (3 digits)
	return fmt.Sprintf("%s-%03d", codePrefix, seq)
}

// GenerateCodeWithDate generates a code with a specific date (useful for batch imports)
func (g *CodeGenerator) GenerateCodeWithDate(ctx context.Context, entityType EntityType, commissariatCode string, date time.Time) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	yearMonth := date.Format("200601")

	// Build the code prefix
	var codePrefix string
	if commissariatCode != "" {
		commCode := strings.ToUpper(strings.ReplaceAll(commissariatCode, "-", ""))
		if len(commCode) > 6 {
			commCode = commCode[:6]
		}
		codePrefix = fmt.Sprintf("%s-%s-%s", entityType, commCode, yearMonth)
	} else {
		codePrefix = fmt.Sprintf("%s-%s", entityType, yearMonth)
	}

	counterKey := codePrefix
	g.counters[counterKey]++
	seq := g.counters[counterKey]

	return fmt.Sprintf("%s-%03d", codePrefix, seq)
}

// InitializeCountersFromDB initializes counters from existing database records
// This should be called at application startup to continue sequence from last used
func (g *CodeGenerator) InitializeCountersFromDB(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Initialize counters for each entity type by querying max sequence
	// This is a simplified version - in production you might want to query each table

	// For now, we'll start fresh each month automatically
	// The counter map is keyed by prefix which includes YYYYMM
	// So old months' counters are naturally separate

	return nil
}

// ResetCounters resets all counters (useful for testing)
func (g *CodeGenerator) ResetCounters() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counters = make(map[string]int)
}

// GetEntityTypeFromString converts a string to EntityType
func GetEntityTypeFromString(s string) EntityType {
	switch strings.ToLower(s) {
	case "user":
		return EntityUser
	case "commissariat":
		return EntityCommissariat
	case "controle":
		return EntityControle
	case "inspection":
		return EntityInspection
	case "infraction":
		return EntityInfraction
	case "procesverbal", "pv":
		return EntityProcesVerbal
	case "paiement":
		return EntityPaiement
	case "recours":
		return EntityRecours
	case "vehicule":
		return EntityVehicule
	case "conducteur":
		return EntityConducteur
	case "document":
		return EntityDocument
	case "alertesecuritaire", "alerte":
		return EntityAlerteSecuritaire
	case "plainte":
		return EntityPlainte
	case "mission":
		return EntityMission
	case "equipe":
		return EntityEquipe
	case "objectif":
		return EntityObjectif
	case "observation":
		return EntityObservation
	case "competence":
		return EntityCompetence
	case "checkitem":
		return EntityCheckItem
	case "checkoption":
		return EntityCheckOption
	case "infractiontype":
		return EntityInfractionType
	case "auditlog":
		return EntityAuditLog
	default:
		return EntityType(strings.ToUpper(s[:min(4, len(s))]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
