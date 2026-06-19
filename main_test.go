package main

import "testing"

// Testet die Add-Funktion mit verschiedenen Eingaben.
func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"Positive Zahlen", 2, 3, 5},
		{"Negative Zahlen", -1, -1, -2},
		{"Gemischte Zahlen", -1, 1, 0},
		{"Null", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
