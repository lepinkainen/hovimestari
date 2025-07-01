package weather

import (
	"testing"

	"github.com/lepinkainen/hovimestari/internal/store"
)

// TestNewImporter tests the NewImporter function
func TestNewImporter(t *testing.T) {
	// Create a mock store
	mockStore := &store.Store{}

	// Test parameters
	latitude := 60.1699
	longitude := 24.9384
	location := "Helsinki"

	// Create a new importer
	importer := NewImporter(mockStore, latitude, longitude, location)

	// Verify the importer was created correctly
	if importer.store != mockStore {
		t.Error("Store not properly set in importer")
	}

	if importer.latitude != latitude {
		t.Errorf("Expected latitude %f, got %f", latitude, importer.latitude)
	}

	if importer.longitude != longitude {
		t.Errorf("Expected longitude %f, got %f", longitude, importer.longitude)
	}

	if importer.location != location {
		t.Errorf("Expected location %q, got %q", location, importer.location)
	}
}

// TestSourcePrefix tests that the SourcePrefix constant is set correctly
func TestSourcePrefix(t *testing.T) {
	expected := "weather-metno"
	if SourcePrefix != expected {
		t.Errorf("Expected SourcePrefix to be %q, got %q", expected, SourcePrefix)
	}
}
