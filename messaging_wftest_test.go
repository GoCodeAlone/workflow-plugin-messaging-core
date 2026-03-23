package messaging_test

import (
	"testing"
	"time"

	messaging "github.com/GoCodeAlone/workflow-plugin-messaging-core"
	"github.com/GoCodeAlone/workflow/wftest"
)

func TestMessagingCore_SendMessagePipeline(t *testing.T) {
	sendRec := wftest.RecordStep("step.messaging_send")
	sendRec.WithOutput(map[string]any{"message_id": "msg-123", "sent": true})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  send-message:
    trigger:
      type: manual
    steps:
      - name: send
        type: step.messaging_send
        config:
          channel: general
          text: "Hello, world!"
`), sendRec)

	result := h.ExecutePipeline("send-message", map[string]any{
		"channel": "general",
		"text":    "Hello",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if sendRec.CallCount() != 1 {
		t.Errorf("expected 1 call, got %d", sendRec.CallCount())
	}

	// Verify the config passed to the step
	calls := sendRec.Calls()
	if calls[0].Config["channel"] != "general" {
		t.Errorf("expected channel=general, got %v", calls[0].Config["channel"])
	}

	// Verify messaging types compile correctly in this context
	var _ messaging.Provider = nil
	var _ messaging.EventListener = nil
	_ = messaging.MessageOpts{}
}

func TestMessagingCore_ReplyPipeline(t *testing.T) {
	replyRec := wftest.RecordStep("step.messaging_reply")
	replyRec.WithOutput(map[string]any{"message_id": "reply-456", "sent": true})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  reply-message:
    trigger:
      type: manual
    steps:
      - name: reply
        type: step.messaging_reply
        config:
          channel: support
          parent_id: "msg-123"
          text: "Thanks for your message!"
`), replyRec)

	result := h.ExecutePipeline("reply-message", map[string]any{
		"channel":   "support",
		"parent_id": "msg-123",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if replyRec.CallCount() != 1 {
		t.Errorf("expected 1 call, got %d", replyRec.CallCount())
	}

	// Verify messaging struct types are usable
	embed := messaging.Embed{
		Title:       "Test",
		Description: "A test embed",
		Color:       0x00ff00,
		Timestamp:   time.Now(),
		Fields: []messaging.EmbedField{
			{Name: "field1", Value: "value1", Inline: true},
		},
	}
	_ = embed

	msg := messaging.Message{
		ID:        "msg-123",
		ChannelID: "ch-456",
		Content:   "Hello",
	}
	_ = msg
}

func TestMessagingCore_EventPipeline(t *testing.T) {
	receiveRec := wftest.RecordStep("step.messaging_receive")
	receiveRec.WithOutput(map[string]any{
		"event_type": "message_create",
		"channel_id": "ch-general",
		"user_id":    "user-789",
		"content":    "Hello bot!",
	})

	reactRec := wftest.RecordStep("step.messaging_react")
	reactRec.WithOutput(map[string]any{"reacted": true})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  handle-event:
    trigger:
      type: manual
    steps:
      - name: receive
        type: step.messaging_receive
        config:
          channel: general
      - name: react
        type: step.messaging_react
        config:
          emoji: "👍"
`), receiveRec, reactRec)

	result := h.ExecutePipeline("handle-event", map[string]any{
		"channel": "general",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if receiveRec.CallCount() != 1 {
		t.Errorf("expected 1 receive call, got %d", receiveRec.CallCount())
	}
	if reactRec.CallCount() != 1 {
		t.Errorf("expected 1 react call, got %d", reactRec.CallCount())
	}

	// Verify Event and Component types compile correctly
	event := messaging.Event{
		Type:      "message_create",
		ChannelID: "ch-general",
		UserID:    "user-789",
		Content:   "Hello bot!",
		Data:      map[string]any{"platform": "slack"},
		Timestamp: time.Now(),
	}
	_ = event

	component := messaging.Component{
		Type: "button",
		Data: map[string]any{"label": "Click me", "style": "primary"},
	}
	_ = component
}
