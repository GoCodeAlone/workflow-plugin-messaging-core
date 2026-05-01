package messaging_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/GoCodeAlone/workflow/plugin"
)

// TestPluginManifest validates that plugin.json parses correctly and satisfies
// the PluginManifest contract required by the Workflow engine.
func TestPluginManifest(t *testing.T) {
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
	data, err := os.ReadFile("plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}

	var m plugin.PluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}

	total := len(m.StepTypes)
	if total == 0 {
		t.Fatal("no step types declared in plugin.json")
	}

	strict := len(m.StepSchemas)
	if strict != total {
		t.Errorf("strict contract coverage: %d/%d step types have schemas (want 100%%)", strict, total)
	}
}
