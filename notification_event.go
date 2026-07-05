package messaging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	StepTypeMessagingSend     = "step.messaging_send"
	PluginFamilyMessagingCore = "workflow-plugin-messaging-core"
	RequiredConfigChannel     = "channel"
)

type MessagingSendInput struct {
	Channel  string `json:"channel,omitempty"`
	Text     string `json:"text"`
	Provider string `json:"provider,omitempty"`
}

type NotificationEvent struct {
	Section   string                `json:"section"`
	Key       string                `json:"key"`
	Value     string                `json:"value,omitempty"`
	Author    string                `json:"author,omitempty"`
	Revision  int64                 `json:"revision,omitempty"`
	Timestamp string                `json:"timestamp,omitempty"`
	Messaging NotificationMessaging `json:"messaging"`
	Workflow  *NotificationWorkflow `json:"workflow,omitempty"`
}

type NotificationMessaging struct {
	Text string `json:"text"`
}

type NotificationWorkflow struct {
	StepType       string                     `json:"stepType"`
	PluginFamily   string                     `json:"pluginFamily"`
	Input          NotificationWorkflowInput  `json:"input"`
	RequiredConfig []string                   `json:"requiredConfig,omitempty"`
	Metadata       NotificationWorkflowSource `json:"metadata,omitempty"`
}

type NotificationWorkflowInput struct {
	Text string `json:"text"`
}

type NotificationWorkflowSource struct {
	Section string `json:"section,omitempty"`
	Key     string `json:"key,omitempty"`
}

func ParseNotificationEvent(data []byte) (NotificationEvent, error) {
	var event NotificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return event, fmt.Errorf("parse notification-event: %w", err)
	}
	return event, nil
}

func ParseNotificationEvents(r io.Reader) ([]NotificationEvent, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read notification-events: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}
	if data[0] == '[' {
		var events []NotificationEvent
		if err := json.Unmarshal(data, &events); err != nil {
			return nil, fmt.Errorf("parse notification-event array: %w", err)
		}
		return events, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	events := make([]NotificationEvent, 0)
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		event, err := ParseNotificationEvent([]byte(raw))
		if err != nil {
			return nil, fmt.Errorf("parse notification-event line %d: %w", line, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan notification-events: %w", err)
	}
	return events, nil
}

func ProjectNotificationEventToMessagingSend(event NotificationEvent, channel string) (MessagingSendInput, error) {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return MessagingSendInput{}, fmt.Errorf("messaging send channel is required")
	}
	text := strings.TrimSpace(event.Messaging.Text)
	if event.Workflow != nil &&
		event.Workflow.StepType == StepTypeMessagingSend &&
		event.Workflow.PluginFamily == PluginFamilyMessagingCore {
		if workflowText := strings.TrimSpace(event.Workflow.Input.Text); workflowText != "" {
			text = workflowText
		}
	}
	if text == "" {
		return MessagingSendInput{}, fmt.Errorf("notification-event missing messaging.text")
	}
	return MessagingSendInput{
		Channel: channel,
		Text:    text,
	}, nil
}
