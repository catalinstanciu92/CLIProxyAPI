package test

import (
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

func TestConfigModelFallbacks(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []config.ModelFallback
		wantErr bool
	}{
		{
			name: "basic fallback",
			yaml: `model-fallbacks:
  - from: "gemini-2.5-pro"
    to: "gemini-2.5-flash"`,
			want:    []config.ModelFallback{{From: "gemini-2.5-pro", To: "gemini-2.5-flash"}},
			wantErr: false,
		},
		{
			name: "chain of fallbacks",
			yaml: `model-fallbacks:
  - from: "model-a"
    to: "model-b"
  - from: "model-b"
    to: "model-c"
  - from: "model-c"
    to: "model-d"`,
			want: []config.ModelFallback{
				{From: "model-a", To: "model-b"},
				{From: "model-b", To: "model-c"},
				{From: "model-c", To: "model-d"},
			},
			wantErr: false,
		},
		{
			name:    "empty fallbacks",
			yaml:    `model-fallbacks: []`,
			want:    []config.ModelFallback{},
			wantErr: false,
		},
		{
			name: "case insensitive matching",
			yaml: `model-fallbacks:
  - from: "Model-A"
    to: "Model-B"`,
			want:    []config.ModelFallback{{From: "Model-A", To: "Model-B"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.yaml)
			cfg, err := config.LoadConfig(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(cfg.ModelFallbacks) != len(tt.want) {
				t.Fatalf("ModelFallbacks length = %d, want %d", len(cfg.ModelFallbacks), len(tt.want))
			}
			for i, fb := range cfg.ModelFallbacks {
				if fb.From != tt.want[i].From || fb.To != tt.want[i].To {
					t.Errorf("ModelFallbacks[%d] = {%q, %q}, want {%q, %q}",
						i, fb.From, fb.To, tt.want[i].From, tt.want[i].To)
				}
			}
		})
	}
}

func TestModelFallbackCycleDetection(t *testing.T) {
	tests := []struct {
		name           string
		yaml           string
		wantCount      int
		wantFromModels []string
		wantToModels   []string
	}{
		{
			name: "direct cycle a->b->a",
			yaml: `model-fallbacks:
  - from: "a"
    to: "b"
  - from: "b"
    to: "a"`,
			wantCount:      1,
			wantFromModels: []string{"a"},
			wantToModels:   []string{"b"},
		},
		{
			name: "transitive cycle a->b->c->a",
			yaml: `model-fallbacks:
  - from: "a"
    to: "b"
  - from: "b"
    to: "c"
  - from: "c"
    to: "a"`,
			wantCount:      2,
			wantFromModels: []string{"a", "b"},
			wantToModels:   []string{"b", "c"},
		},
		{
			name: "self reference",
			yaml: `model-fallbacks:
  - from: "a"
    to: "a"`,
			wantCount: 0,
		},
		{
			name: "cycle in middle of chain",
			yaml: `model-fallbacks:
  - from: "a"
    to: "b"
  - from: "b"
    to: "c"
  - from: "c"
    to: "b"
  - from: "d"
    to: "e"`,
			wantCount:      3,
			wantFromModels: []string{"a", "b", "d"},
			wantToModels:   []string{"b", "c", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.yaml)
			cfg, err := config.LoadConfig(path)
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if len(cfg.ModelFallbacks) != tt.wantCount {
				t.Errorf("ModelFallbacks count = %d, want %d", len(cfg.ModelFallbacks), tt.wantCount)
			}
			for i, fb := range cfg.ModelFallbacks {
				if i < len(tt.wantFromModels) && (fb.From != tt.wantFromModels[i] || fb.To != tt.wantToModels[i]) {
					t.Errorf("ModelFallbacks[%d] = {%q, %q}, want {%q, %q}",
						i, fb.From, fb.To, tt.wantFromModels[i], tt.wantToModels[i])
				}
			}
		})
	}
}

func TestModelFallbackDepth(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedDepth int
	}{
		{
			name:          "default depth 3",
			yaml:          `model-fallbacks: []`,
			expectedDepth: 3,
		},
		{
			name: "custom depth",
			yaml: `model-fallbacks: []
model-fallback-depth: 5`,
			expectedDepth: 5,
		},
		{
			name: "zero depth disables fallbacks",
			yaml: `model-fallbacks: []
model-fallback-depth: 0`,
			expectedDepth: 0,
		},
		{
			name: "negative depth defaults to 3",
			yaml: `model-fallbacks: []
model-fallback-depth: -1`,
			expectedDepth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.yaml)
			cfg, err := config.LoadConfig(path)
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if cfg.ModelFallbackDepth == nil || *cfg.ModelFallbackDepth != tt.expectedDepth {
				var actual int
				if cfg.ModelFallbackDepth != nil {
					actual = *cfg.ModelFallbackDepth
				}
				t.Errorf("ModelFallbackDepth = %d, want %d", actual, tt.expectedDepth)
			}
		})
	}
}

func TestSanitizeModelFallbacks(t *testing.T) {
	tests := []struct {
		name      string
		fallbacks []config.ModelFallback
		expected  []config.ModelFallback
	}{
		{
			name: "removes empty from entries",
			fallbacks: []config.ModelFallback{
				{From: "", To: "b"},
				{From: "a", To: "b"},
			},
			expected: []config.ModelFallback{
				{From: "a", To: "b"},
			},
		},
		{
			name: "removes empty to entries",
			fallbacks: []config.ModelFallback{
				{From: "a", To: ""},
				{From: "a", To: "b"},
			},
			expected: []config.ModelFallback{
				{From: "a", To: "b"},
			},
		},
		{
			name: "removes self-references",
			fallbacks: []config.ModelFallback{
				{From: "a", To: "a"},
				{From: "b", To: "c"},
			},
			expected: []config.ModelFallback{
				{From: "b", To: "c"},
			},
		},
		{
			name: "case insensitive self-references",
			fallbacks: []config.ModelFallback{
				{From: "Model-A", To: "model-a"},
			},
			expected: []config.ModelFallback{},
		},
		{
			name: "deduplicates exact entries",
			fallbacks: []config.ModelFallback{
				{From: "a", To: "b"},
				{From: "a", To: "b"},
				{From: "a", To: "b"},
			},
			expected: []config.ModelFallback{
				{From: "a", To: "b"},
			},
		},
		{
			name: "deduplicates case insensitive entries",
			fallbacks: []config.ModelFallback{
				{From: "a", To: "b"},
				{From: "A", To: "B"},
			},
			expected: []config.ModelFallback{
				{From: "a", To: "b"},
			},
		},
		{
			name: "preserves different entries with same from but different to",
			fallbacks: []config.ModelFallback{
				{From: "a", To: "b"},
				{From: "a", To: "c"},
			},
			expected: []config.ModelFallback{
				{From: "a", To: "b"},
				{From: "a", To: "c"},
			},
		},
		{
			name: "trims whitespace",
			fallbacks: []config.ModelFallback{
				{From: "  model-a  ", To: "  model-b  "},
			},
			expected: []config.ModelFallback{
				{From: "model-a", To: "model-b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ModelFallbacks: tt.fallbacks,
			}
			cfg.SanitizeModelFallbacks()
			if len(cfg.ModelFallbacks) != len(tt.expected) {
				t.Errorf("ModelFallbacks length = %d, want %d", len(cfg.ModelFallbacks), len(tt.expected))
				return
			}
			for i, fb := range cfg.ModelFallbacks {
				if fb.From != tt.expected[i].From || fb.To != tt.expected[i].To {
					t.Errorf("ModelFallbacks[%d] = {%q, %q}, want {%q, %q}",
						i, fb.From, fb.To, tt.expected[i].From, tt.expected[i].To)
				}
			}
		})
	}
}

func TestFallbackLookup(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		model        string
		wantFallback string
	}{
		{
			name: "finds direct fallback",
			yaml: `model-fallbacks:
  - from: "gemini-2.5-pro"
    to: "gemini-2.5-flash"`,
			model:        "gemini-2.5-pro",
			wantFallback: "gemini-2.5-flash",
		},
		{
			name: "model without fallback returns empty",
			yaml: `model-fallbacks:
  - from: "model-a"
    to: "model-b"`,
			model:        "model-c",
			wantFallback: "",
		},
		{
			name: "case insensitive lookup",
			yaml: `model-fallbacks:
  - from: "Model-A"
    to: "Model-B"`,
			model:        "model-a",
			wantFallback: "Model-B",
		},
		{
			name: "case insensitive from matching",
			yaml: `model-fallbacks:
  - from: "model-a"
    to: "model-b"`,
			model:        "MODEL-A",
			wantFallback: "model-b",
		},
		{
			name: "empty model returns empty",
			yaml: `model-fallbacks:
  - from: "a"
    to: "b"`,
			model:        "",
			wantFallback: "",
		},
		{
			name: "multiple fallbacks chain",
			yaml: `model-fallbacks:
  - from: "a"
    to: "b"
  - from: "b"
    to: "c"
  - from: "c"
    to: "d"`,
			model:        "a",
			wantFallback: "b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.yaml)
			cfg, err := config.LoadConfig(path)
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			fallback := lookupFallback(cfg, tt.model)
			if fallback != tt.wantFallback {
				t.Errorf("lookupFallback(%q) = %q, want %q", tt.model, fallback, tt.wantFallback)
			}
		})
	}
}

func lookupFallback(cfg *config.Config, model string) string {
	if cfg == nil || model == "" {
		return ""
	}
	for _, fb := range cfg.ModelFallbacks {
		if strings.EqualFold(fb.From, model) {
			return fb.To
		}
	}
	return ""
}
