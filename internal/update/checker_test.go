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
		// git describe 偏移版本号场景
		{
			name:    "git describe equal base",
			latest:  "0.2.1",
			current: "0.2.1-3-gc255d1d",
			want:    false, // 基础版本号相同，不应报更新
		},
		{
			name:    "git describe lower base",
			latest:  "0.2.1",
			current: "0.2.0-4-g0288843",
			want:    true, // 基础版本号 0.2.0 < 0.2.1
		},
		{
			name:    "git describe dirty suffix",
			latest:  "0.2.1",
			current: "0.2.1-dirty",
			want:    false, // 基础版本号相同
		},
		{
			name:    "git describe full dirty",
			latest:  "0.2.1",
			current: "0.2.1-3-gc255d1d-dirty",
			want:    false, // 基础版本号相同
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

func TestStripGitDescribeSuffix(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:  "plain version",
			input: "0.2.1",
			want:  "0.2.1",
		},
		{
			name:  "git describe offset",
			input: "0.2.1-3-gc255d1d",
			want:  "0.2.1",
		},
		{
			name:  "git describe dirty",
			input: "0.2.1-dirty",
			want:  "0.2.1",
		},
		{
			name:  "git describe offset dirty",
			input: "0.2.1-3-gc255d1d-dirty",
			want:  "0.2.1",
		},
		{
			name:  "v prefix with offset",
			input: "v0.2.1-3-gc255d1d",
			want:  "v0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripGitDescribeSuffix(tt.input)
			if got != tt.want {
				t.Errorf("stripGitDescribeSuffix(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
