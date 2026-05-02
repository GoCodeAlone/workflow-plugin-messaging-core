package messaging_test

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow/plugin"
	"github.com/GoCodeAlone/workflow/schema"
)

func loadPluginManifest(t *testing.T) plugin.PluginManifest {
	t.Helper()

	data, err := os.ReadFile("plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}

	var m plugin.PluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}

	if err := m.Validate(); err != nil {
		t.Fatalf("plugin.json validation failed: %v", err)
	}

	return m
}

// TestPluginManifest validates that plugin.json parses correctly and satisfies
// the PluginManifest contract required by the Workflow engine.
func TestPluginManifest(t *testing.T) {
	m := loadPluginManifest(t)

	// Verify expected step types are declared.
	expectedSteps := []string{
		"step.messaging_send",
		"step.messaging_reply",
		"step.messaging_receive",
		"step.messaging_react",
	}
	stepSet := make(map[string]bool, len(m.StepTypes))
	for _, s := range m.StepTypes {
		stepSet[s] = true
	}
	for _, want := range expectedSteps {
		if !stepSet[want] {
			t.Errorf("stepTypes: missing %q", want)
		}
	}

	// Verify each declared step type has a corresponding step schema (strict contract).
	schemaSet := make(map[string]bool, len(m.StepSchemas))
	for _, s := range m.StepSchemas {
		schemaSet[s.Type] = true
	}
	for _, stepType := range m.StepTypes {
		if !schemaSet[stepType] {
			t.Errorf("stepSchemas: missing strict schema for step type %q", stepType)
		}
	}

	// Verify each step schema has required fields.
	for _, s := range m.StepSchemas {
		if s.Type == "" {
			t.Errorf("stepSchema has empty type")
		}
		if s.Description == "" {
			t.Errorf("stepSchema %q has empty description", s.Type)
		}
		// All messaging step types must have at least one config field (e.g. channel)
		// because they all require a destination or context to operate on.
		if len(s.ConfigFields) == 0 {
			t.Errorf("stepSchema %q has no configFields", s.Type)
		}
		// All messaging step types produce at least one output (e.g. message_id, sent,
		// reacted) to allow downstream steps to react to the result.
		if len(s.Outputs) == 0 {
			t.Errorf("stepSchema %q has no outputs", s.Type)
		}
		for _, f := range s.ConfigFields {
			if f.Key == "" {
				t.Errorf("stepSchema %q: configField has empty key", s.Type)
			}
			if f.Type == "" {
				t.Errorf("stepSchema %q: configField %q has empty type", s.Type, f.Key)
			}
		}
		for _, o := range s.Outputs {
			if o.Key == "" {
				t.Errorf("stepSchema %q: output has empty key", s.Type)
			}
			if o.Type == "" {
				t.Errorf("stepSchema %q: output %q has empty type", s.Type, o.Key)
			}
		}
	}
}

// TestPluginManifest_StrictCoverage verifies that the strict-contract coverage
// ratio for step types is 100% (every declared step has a full schema).
func TestPluginManifest_StrictCoverage(t *testing.T) {
	m := loadPluginManifest(t)

	assertStrictStepSchemaCoverage(t, m)
}

func TestStrictStepSchemaCoverageReportsContractDrift(t *testing.T) {
	tests := []struct {
		name string
		m    plugin.PluginManifest
		want []string
	}{
		{
			name: "valid coverage",
			m: plugin.PluginManifest{
				StepTypes: []string{"step.messaging_send", "step.messaging_reply"},
				StepSchemas: []*schema.StepSchema{
					{Type: "step.messaging_send"},
					{Type: "step.messaging_reply"},
				},
			},
			want: nil,
		},
		{
			name: "missing schema",
			m: plugin.PluginManifest{
				StepTypes: []string{"step.messaging_send", "step.messaging_reply"},
				StepSchemas: []*schema.StepSchema{
					{Type: "step.messaging_send"},
				},
			},
			want: []string{`missing strict schema for step type "step.messaging_reply"`},
		},
		{
			name: "duplicate schema",
			m: plugin.PluginManifest{
				StepTypes: []string{"step.messaging_send"},
				StepSchemas: []*schema.StepSchema{
					{Type: "step.messaging_send"},
					{Type: "step.messaging_send"},
				},
			},
			want: []string{`duplicate strict schema for step type "step.messaging_send"`},
		},
		{
			name: "undeclared schema",
			m: plugin.PluginManifest{
				StepTypes: []string{"step.messaging_send"},
				StepSchemas: []*schema.StepSchema{
					{Type: "step.messaging_send"},
					{Type: "step.messaging_reply"},
				},
			},
			want: []string{`strict schema "step.messaging_reply" is not declared in stepTypes`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strictStepSchemaCoverageErrors(tt.m)
			if len(tt.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("strictStepSchemaCoverageErrors() = %v, want no errors", got)
				}
				return
			}
			for _, want := range tt.want {
				if !containsError(got, want) {
					t.Fatalf("strictStepSchemaCoverageErrors() = %v, want error containing %q", got, want)
				}
			}
		})
	}
}

func assertStrictStepSchemaCoverage(t *testing.T, m plugin.PluginManifest) {
	t.Helper()

	for _, errMsg := range strictStepSchemaCoverageErrors(m) {
		t.Error(errMsg)
	}
}

func strictStepSchemaCoverageErrors(m plugin.PluginManifest) []string {
	var errs []string
	if len(m.StepTypes) == 0 {
		return []string{"no step types declared in plugin.json"}
	}

	stepSet := make(map[string]bool, len(m.StepTypes))
	for _, stepType := range m.StepTypes {
		if stepSet[stepType] {
			errs = append(errs, "duplicate step type "+strconv.Quote(stepType))
		}
		stepSet[stepType] = true
	}

	schemaCounts := make(map[string]int, len(m.StepSchemas))
	for _, stepSchema := range m.StepSchemas {
		if stepSchema == nil {
			errs = append(errs, "nil step schema")
			continue
		}
		schemaCounts[stepSchema.Type]++
	}

	covered := 0
	for _, stepType := range m.StepTypes {
		count := schemaCounts[stepType]
		switch {
		case count == 0:
			errs = append(errs, "missing strict schema for step type "+strconv.Quote(stepType))
		case count > 1:
			errs = append(errs, "duplicate strict schema for step type "+strconv.Quote(stepType))
			covered++
		default:
			covered++
		}
	}

	for schemaType := range schemaCounts {
		if !stepSet[schemaType] {
			errs = append(errs, "strict schema "+strconv.Quote(schemaType)+" is not declared in stepTypes")
		}
	}

	if covered != len(m.StepTypes) {
		errs = append(errs, "strict contract coverage: "+strconv.Itoa(covered)+"/"+strconv.Itoa(len(m.StepTypes))+" step types have schemas (want 100%)")
	}

	return errs
}

func containsError(errs []string, want string) bool {
	for _, errMsg := range errs {
		if strings.Contains(errMsg, want) {
			return true
		}
	}
	return false
}
