package main

import "testing"

// Testet die IstPrimzahl-Funktion mit verschiedenen Eingaben.
func TestIstPrimzahl(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected bool
	}{
		{"Primzahl 2", 2, true},
		{"Primzahl 3", 3, true},
		{"Primzahl 5", 5, true},
		{"Keine Primzahl 4", 4, false},
		{"Keine Primzahl 6", 6, false},
		{"Keine Primzahl 1", 1, false},
		{"Keine Primzahl 0", 0, false},
		{"Negative Zahl", -1, false},
		{"Große Primzahl", 7919, true},
		{"Große keine Primzahl", 7920, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IstPrimzahl(tt.n)
			if result != tt.expected {
				t.Errorf("IstPrimzahl(%d) = %t; want %t", tt.n, result, tt.expected)
			}
		})
	}
}
