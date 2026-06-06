package matrix

import (
	"testing"

	"github.com/google/uuid"
)

func Test_generateUUID(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "Returns valid UUID", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			got := generateUUID()
			if err = uuid.Validate(got); (err == nil) != tt.want {
				t.Errorf("generateUUID() = %v, want %v", err == nil, tt.want)
			}
		})
	}
}
