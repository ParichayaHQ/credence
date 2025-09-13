package store

import (
	"github.com/ParichayaHQ/credence/internal/events"
)

// generateCID generates a simple CID-like identifier from data
func generateCID(data []byte) (string, error) {
	return events.GenerateCID(data)
}