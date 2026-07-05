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
	StepTypeMessagingSend        = "step.messaging_send"
	PluginFamilyMessagingCore    = "workflow-plugin-messaging-core"
	RatchetRequiredConfigChannel = "channel"
)

type MessagingSendInput struct {
	Channel  string `json:"channel,omitempty"`
	Text     string `json:"text"`
	Provider string `json:"provider,omitempty"`
}

type RatchetNotificationEvent struct {
	Section   string                     `json:"section"`
	Key       string                     `json:"key"`
	Value     string                     `json:"value,omitempty"`
	Author    string                     `json:"author,omitempty"`
	Revision  int64                      `json:"revision,omitempty"`
	Timestamp string                     `json:"timestamp,omitempty"`
	Messaging RatchetMessagingRecord     `json:"messaging"`
	Workflow  *RatchetWorkflowProjection `json:"workflow,omitempty"`
}

type RatchetMessagingRecord struct {
	Text string `json:"text"`
}

type RatchetWorkflowProjection struct {
	StepType       string                    `json:"stepType"`
	PluginFamily   string                    `json:"pluginFamily"`
	Input          RatchetWorkflowInput      `json:"input"`
	RequiredConfig []string                  `json:"requiredConfig,omitempty"`
	Metadata       RatchetWorkflowSourceMeta `json:"metadata,omitempty"`
}

type RatchetWorkflowInput struct {
	Text string `json:"text"`
}

type RatchetWorkflowSourceMeta struct {
	Section string `json:"section,omitempty"`
	Key     string `json:"key,omitempty"`
}

func ParseRatchetNotificationEvent(data []byte) (RatchetNotificationEvent, error) {
	var event RatchetNotificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return event, fmt.Errorf("parse ratchet notification-event: %w", err)
	}
	return event, nil
}

func ParseRatchetNotificationEvents(r io.Reader) ([]RatchetNotificationEvent, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read ratchet notification-events: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}
	if data[0] == '[' {
		var events []RatchetNotificationEvent
		if err := json.Unmarshal(data, &events); err != nil {
			return nil, fmt.Errorf("parse ratchet notification-event array: %w", err)
		}
		return events, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	events := make([]RatchetNotificationEvent, 0)
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		event, err := ParseRatchetNotificationEvent([]byte(raw))
		if err != nil {
			return nil, fmt.Errorf("parse ratchet notification-event line %d: %w", line, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan ratchet notification-events: %w", err)
	}
	return events, nil
}

func ProjectRatchetNotificationToMessagingSend(event RatchetNotificationEvent, channel string) (MessagingSendInput, error) {
	text := strings.TrimSpace(event.Messaging.Text)
	if event.Workflow != nil && event.Workflow.StepType == StepTypeMessagingSend {
		if workflowText := strings.TrimSpace(event.Workflow.Input.Text); workflowText != "" {
			text = workflowText
		}
	}
	if text == "" {
		return MessagingSendInput{}, fmt.Errorf("ratchet notification-event missing messaging.text")
	}
	return MessagingSendInput{
		Channel: strings.TrimSpace(channel),
		Text:    text,
	}, nil
}
