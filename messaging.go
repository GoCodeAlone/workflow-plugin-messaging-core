// Package messaging defines shared interfaces for workflow platform messaging plugins.
// Each platform plugin (Discord, Slack, Teams) implements these interfaces.
package messaging

import (
	"context"
	"io"
	"time"
)

// MessageOpts configures optional message parameters.
type MessageOpts struct {
	Embeds     []Embed      // Rich embeds (Discord), blocks (Slack), cards (Teams)
	Files      []FileAttachment // Inline file attachments
	ThreadID   string       // Reply to thread/parent message
	Ephemeral  bool         // Only visible to one user (where supported)
	Components []Component  // Interactive components (buttons, menus)
}

// Embed represents a rich message attachment (platform-specific rendering).
type Embed struct {
	Title       string
	Description string
	URL         string
	Color       int
	Fields      []EmbedField
	ImageURL    string
	FooterText  string
	Timestamp   time.Time
}

// EmbedField is a named field within an Embed.
type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

// FileAttachment is a file to attach to a message.
type FileAttachment struct {
	Name   string
	Reader io.Reader
}

// Component is an interactive UI element (button, select menu, etc.).
type Component struct {
	Type string         // "button", "select", "action_row"
	Data map[string]any // Platform-specific component data
}

// Message represents a sent or received message.
type Message struct {
	ID        string
	ChannelID string
	AuthorID  string
	Content   string
	Timestamp time.Time
	ThreadID  string
	Embeds    []Embed
}

// Event represents a real-time platform event.
type Event struct {
	Type      string         // "message_create", "message_update", "reaction_add", "member_join", etc.
	ChannelID string
	UserID    string
	MessageID string
	Content   string
	Data      map[string]any // Platform-specific event data
	Timestamp time.Time
}

// Provider is the common messaging interface implemented by each platform plugin.
type Provider interface {
	// Name returns the platform identifier ("discord", "slack", "teams").
	Name() string

	// SendMessage sends a message to a channel and returns the message ID.
	SendMessage(ctx context.Context, channelID, content string, opts *MessageOpts) (string, error)

	// EditMessage updates an existing message.
	EditMessage(ctx context.Context, channelID, messageID, content string) error

	// DeleteMessage removes a message.
	DeleteMessage(ctx context.Context, channelID, messageID string) error

	// SendReply sends a threaded reply and returns the message ID.
	SendReply(ctx context.Context, channelID, parentID, content string, opts *MessageOpts) (string, error)

	// React adds a reaction to a message.
	React(ctx context.Context, channelID, messageID, emoji string) error

	// UploadFile sends a file to a channel and returns the message/file ID.
	UploadFile(ctx context.Context, channelID string, file io.Reader, filename string) (string, error)
}

// EventListener receives real-time events from the platform.
type EventListener interface {
	// Listen starts receiving events. The returned channel is closed when
	// the context is cancelled or Close is called.
	Listen(ctx context.Context) (<-chan Event, error)

	// Close stops the event listener and releases resources.
	Close() error
}

// VoiceProvider is optionally implemented by platforms with voice support (Discord).
type VoiceProvider interface {
	JoinVoice(ctx context.Context, guildID, channelID string) error
	LeaveVoice(ctx context.Context, guildID string) error
	PlayAudio(ctx context.Context, guildID string, audio io.Reader) error
}
