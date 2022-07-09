package model_test

import (
	"github.com/alexadhy/shortener/model"
	"testing"
)

func TestGenFake(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		hasError bool
	}{
		{
			name:     "should be able to generate a single fake data",
			input:    1,
			hasError: false,
		},
		{
			name:     "should be able to generate large number of fake data",
			input:    200,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := model.GenFake(tt.input)
			if tt.hasError && err == nil {
				t.Fatalf("should have error, instead got nil")
			}
			if !tt.hasError {
				if err != nil {
					t.Fatalf("shouldn't return any error, got: %v", err)
				}
				t.Logf("generated fake data: %v", res)
			}
		})
	}
}
