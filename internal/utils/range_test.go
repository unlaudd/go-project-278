package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{"valid no spaces", "[0,10]", 0, 10, false},
		{"valid with spaces", "[5, 15]", 5, 15, false},
		{"single element", "[3,3]", 3, 3, false},
		{"large numbers", "[100,200]", 100, 200, false},
		{"empty input", "", 0, 0, true},
		{"no brackets", "0,10", 0, 0, true},
		{"single bracket", "[0,10", 0, 0, true},
		{"non-numeric start", "[a,10]", 0, 0, true},
		{"non-numeric end", "[0,b]", 0, 0, true},
		{"negative start", "[-1,10]", 0, 0, true},
		{"negative end", "[0,-5]", 0, 0, true},
		{"end before start", "[10,5]", 0, 0, true},
		{"extra commas", "[0,10,20]", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseRange(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStart, start)
				assert.Equal(t, tt.wantEnd, end)
			}
		})
	}
}
