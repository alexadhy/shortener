package hash

import "testing"

func TestHash(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedShort string
		expectedSum   string
	}{
		{
			name:          "should be able to hash url successfully",
			input:         "https://qa.omh.life/console/ohmyhome/listings?accountUid=200b8b81656f7dca38c57ab87bfaac89",
			expectedShort: "67UOYTH8",
			expectedSum:   "f97bfc3e7a0825cbafb7b5851753d02106f66ab13c3ebcf71e72934386f6f629",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum, short := Hash(tt.input)
			if short == "" || sum == "" {
				t.Fatal("empty hash")
			}
			if sum != tt.expectedSum {
				t.Fatalf("expecting %s, got: %s", tt.expectedSum, sum)
			}

			if short != tt.expectedShort {
				t.Fatalf("expecting %s, got: %s", tt.expectedShort, short)
			}
		})
	}
}
