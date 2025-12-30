package schoollunch

import (
	"testing"
	"time"

	"github.com/lepinkainen/hovimestari/internal/store"
	lunch "github.com/lepinkainen/palmia-lunch/lunch"
)

func TestNewImporter(t *testing.T) {
	mockStore := &store.Store{}
	url := "https://example.com/lunch"
	schoolName := "Test School"

	importer := NewImporter(mockStore, url, schoolName)

	if importer.store != mockStore {
		t.Error("Store not properly set in importer")
	}

	if importer.url != url {
		t.Errorf("URL not properly set: expected %s, got %s", url, importer.url)
	}

	if importer.schoolName != schoolName {
		t.Errorf("SchoolName not properly set: expected %s, got %s", schoolName, importer.schoolName)
	}
}

func TestSourcePrefix(t *testing.T) {
	expected := "schoollunch"
	if SourcePrefix != expected {
		t.Errorf("SourcePrefix incorrect: expected %s, got %s", expected, SourcePrefix)
	}
}

func TestFormatMealContent(t *testing.T) {
	tests := []struct {
		name     string
		day      *lunch.Day
		expected string
	}{
		{
			name: "full meal with allergens and components",
			day: &lunch.Day{
				Date:    time.Date(2025, 12, 18, 0, 0, 0, 0, time.UTC),
				Weekday: "Keskiviikko",
				Lunch: lunch.Meal{
					Name:       "Lihapiirakat",
					Allergens:  []string{"K", "G"},
					Components: []string{"Perunamuusi", "Suolakurkku"},
				},
				Vegetarian: lunch.Meal{
					Name:       "Juustopiirakat",
					Allergens:  []string{"K", "G", "M"},
					Components: []string{"Punainen porkkana", "Leivän viipale"},
				},
			},
			expected: "Lounas: Lihapiirakat (K,G)\nOsat: Perunamuusi, Suolakurkku\nKasvislounas: Juustopiirakat (K,G,M)\nOsat: Punainen porkkana, Leivän viipale",
		},
		{
			name: "meal without allergens",
			day: &lunch.Day{
				Date:    time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC),
				Weekday: "Torstai",
				Lunch: lunch.Meal{
					Name:       "Lohikeitto",
					Components: []string{"Ruisleipä"},
				},
				Vegetarian: lunch.Meal{
					Name:       "Kasvispihvit",
					Components: []string{"Perunat"},
				},
			},
			expected: "Lounas: Lohikeitto\nOsat: Ruisleipä\nKasvislounas: Kasvispihvit\nOsat: Perunat",
		},
		{
			name: "meal without components",
			day: &lunch.Day{
				Date:    time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC),
				Weekday: "Perjantai",
				Lunch: lunch.Meal{
					Name:      "Kalakeitto",
					Allergens: []string{"L"},
				},
				Vegetarian: lunch.Meal{
					Name:      "Salaatti",
					Allergens: []string{"VEG"},
				},
			},
			expected: "Lounas: Kalakeitto (L)\nKasvislounas: Salaatti (VEG)",
		},
		{
			name: "only lunch, no vegetarian",
			day: &lunch.Day{
				Date:    time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC),
				Weekday: "Maanantai",
				Lunch: lunch.Meal{
					Name:       "Jauhelihakastike",
					Allergens:  []string{"G"},
					Components: []string{"Makaroni"},
				},
			},
			expected: "Lounas: Jauhelihakastike (G)\nOsat: Makaroni",
		},
		{
			name: "empty meals",
			day: &lunch.Day{
				Date:    time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC),
				Weekday: "Lauantai",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMealContent(tt.day)
			if result != tt.expected {
				t.Errorf("formatMealContent() failed\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
