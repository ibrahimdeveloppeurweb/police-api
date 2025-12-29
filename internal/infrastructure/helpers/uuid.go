package helpers

import (
	"github.com/google/uuid"
)

// ParseUUID converts a string to uuid.UUID
// Returns uuid.Nil if the string is empty or invalid
func ParseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// ParseUUIDPtr converts a *string to uuid.UUID
// Returns uuid.Nil if the pointer is nil or the string is empty/invalid
func ParseUUIDPtr(s *string) uuid.UUID {
	if s == nil || *s == "" {
		return uuid.Nil
	}
	return ParseUUID(*s)
}

// MustParseUUID converts a string to uuid.UUID
// Panics if the string is invalid (use for known-good UUIDs)
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// UUIDToString converts uuid.UUID to string
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
