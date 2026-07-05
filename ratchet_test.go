package messaging

import (
	"strings"
	"testing"
)

func TestParseRatchetNotificationEventProjectsMessagingSend(t *testing.T) {
	event, err := ParseRatchetNotificationEvent([]byte(`{
		"section":"coordination",
		"key":"status",
		"value":"ready",
		"author":"agent",
		"revision":3,
		"timestamp":"2026-07-05T06:00:00Z",
		"messaging":{"text":"[coordination/status] ready"}
	}`))
	if err != nil {
		t.Fatalf("parse event: %v", err)
	}
	if event.Section != "coordination" || event.Messaging.Text != "[coordination/status] ready" {
		t.Fatalf("event = %#v", event)
	}

	input, err := ProjectRatchetNotificationToMessagingSend(event, "ops")
	if err != nil {
		t.Fatalf("project event: %v", err)
	}
	if input.Channel != "ops" || input.Text != "[coordination/status] ready" {
		t.Fatalf("input = %#v", input)
	}
}

func TestParseRatchetNotificationEventsJSONL(t *testing.T) {
	events, err := ParseRatchetNotificationEvents(strings.NewReader(`
{"section":"coordination","key":"status","messaging":{"text":"[coordination/status] ready"}}

{"section":"release","key":"gate","messaging":{"text":"[release/gate] green"}}
`))
	if err != nil {
		t.Fatalf("parse events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d", len(events))
	}
	if events[1].Messaging.Text != "[release/gate] green" {
		t.Fatalf("events[1] = %#v", events[1])
	}
}

func TestParseRatchetNotificationEventsJSONArray(t *testing.T) {
	events, err := ParseRatchetNotificationEvents(strings.NewReader(`[
		{"section":"coordination","key":"status","messaging":{"text":"[coordination/status] ready"}},
		{"section":"release","key":"gate","messaging":{"text":"[release/gate] green"}}
	]`))
	if err != nil {
		t.Fatalf("parse events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d", len(events))
	}
}

func TestProjectRatchetNotificationUsesWorkflowProjection(t *testing.T) {
	event, err := ParseRatchetNotificationEvent([]byte(`{
		"section":"coordination",
		"key":"handoff",
		"messaging":{"text":"fallback"},
		"workflow":{
			"stepType":"step.messaging_send",
			"pluginFamily":"workflow-plugin-messaging-core",
			"input":{"text":"workflow text"},
			"requiredConfig":["channel"],
			"metadata":{"section":"coordination","key":"handoff"}
		}
	}`))
	if err != nil {
		t.Fatalf("parse event: %v", err)
	}
	input, err := ProjectRatchetNotificationToMessagingSend(event, "triage")
	if err != nil {
		t.Fatalf("project event: %v", err)
	}
	if input.Text != "workflow text" || input.Channel != "triage" {
		t.Fatalf("input = %#v", input)
	}
}

func TestProjectRatchetNotificationRejectsMissingChannel(t *testing.T) {
	event := RatchetNotificationEvent{Messaging: RatchetMessagingRecord{Text: "ready"}}
	if _, err := ProjectRatchetNotificationToMessagingSend(event, " "); err == nil {
		t.Fatal("expected missing channel error")
	}
}

func TestProjectRatchetNotificationIgnoresForeignWorkflowProjection(t *testing.T) {
	event := RatchetNotificationEvent{
		Messaging: RatchetMessagingRecord{Text: "fallback"},
		Workflow: &RatchetWorkflowProjection{
			StepType:     StepTypeMessagingSend,
			PluginFamily: "other-plugin",
			Input:        RatchetWorkflowInput{Text: "foreign text"},
		},
	}
	input, err := ProjectRatchetNotificationToMessagingSend(event, "ops")
	if err != nil {
		t.Fatalf("project event: %v", err)
	}
	if input.Text != "fallback" {
		t.Fatalf("input = %#v", input)
	}
}

func TestProjectRatchetNotificationRejectsMissingText(t *testing.T) {
	_, err := ProjectRatchetNotificationToMessagingSend(RatchetNotificationEvent{}, "ops")
	if err == nil {
		t.Fatal("expected missing messaging text error")
	}
}
