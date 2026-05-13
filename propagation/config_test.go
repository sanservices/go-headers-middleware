package propagation

import (
	"reflect"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{"empty", "", nil},
		{"single", "X-Deployment-Color", []string{"X-Deployment-Color"}},
		{"multiple", "X-Deployment-Color,X-Canary", []string{"X-Deployment-Color", "X-Canary"}},
		{"whitespace trimmed", "  X-A  ,  X-B  ", []string{"X-A", "X-B"}},
		{"empty entries skipped", "X-A,,X-B,", []string{"X-A", "X-B"}},
		{"duplicates case-insensitive", "X-A,x-a,X-A", []string{"X-A"}},
		{"preserves first casing on dup", "X-Foo,x-foo", []string{"X-Foo"}},
	}

	const envVar = "TEST_FORWARD_HEADERS"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envVar, tt.value)

			got := LoadConfigFromEnv(envVar)

			if len(tt.want) == 0 && len(got.Headers) == 0 {
				return
			}
			if !reflect.DeepEqual(got.Headers, tt.want) {
				t.Errorf("Headers = %#v, want %#v", got.Headers, tt.want)
			}
		})
	}
}

func TestConfigEnabled(t *testing.T) {
	if (Config{}).Enabled() {
		t.Error("empty Config should not be Enabled")
	}
	if !(Config{Headers: []string{"X-A"}}).Enabled() {
		t.Error("non-empty Config should be Enabled")
	}
}

func TestLoadConfigDefaultEnvVar(t *testing.T) {
	t.Setenv(DefaultEnvVar, "X-Foo")
	cfg := LoadConfig()
	if !cfg.Enabled() || cfg.Headers[0] != "X-Foo" {
		t.Errorf("LoadConfig() = %#v, want [X-Foo]", cfg.Headers)
	}
}
