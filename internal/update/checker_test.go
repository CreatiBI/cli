package update

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{
			name:    "latest greater",
			latest:  "0.1.18",
			current: "0.1.17",
			want:    true,
		},
		{
			name:    "latest equal",
			latest:  "0.1.17",
			current: "0.1.17",
			want:    false,
		},
		{
			name:    "latest lower",
			latest:  "0.1.16",
			current: "0.1.17",
			want:    false,
		},
		{
			name:    "with v prefix",
			latest:  "v0.1.18",
			current: "v0.1.17",
			want:    true,
		},
		{
			name:    "mixed prefix",
			latest:  "v0.1.18",
			current: "0.1.17",
			want:    true,
		},
		{
			name:    "invalid version",
			latest:  "invalid",
			current: "0.1.17",
			want:    false,
		},
		{
			name:    "major version update",
			latest:  "1.0.0",
			current: "0.1.17",
			want:    true,
		},
		{
			name:    "patch version update",
			latest:  "0.1.17",
			current: "0.1.16",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("compareVersions(%s, %s) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}
