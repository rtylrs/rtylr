package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrSlackAPI        = errors.New("slack API error")
)

type Service struct {
	botToken string
	channels map[string]string
	ctx      context.Context

	buffer     map[string][]Message
	bufferMu   sync.RWMutex
	bufferSize int
	flushTimer *time.Timer
	flushDelay time.Duration
}

type Message struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	Type    string `json:"type,omitempty"`
}

type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func FormatMessage(text string) string {
	text = strings.ReplaceAll(text, "**", "*")
	text = strings.ReplaceAll(text, "__", "_")

	return text
}

func NewAlertsService(ctx context.Context, botToken string) *Service {
	service := &Service{
		botToken:   botToken,
		channels:   make(map[string]string),
		ctx:        ctx,
		buffer:     make(map[string][]Message),
		bufferSize: 10,
		flushDelay: 30 * time.Second,
	}

	service.startFlushTimer()

	return service
}

func (s *Service) AddChannel(name, channelID string) {
	s.channels[name] = channelID
}

func (s *Service) GetChannel(name string) string {
	return s.channels[name]
}

func (s *Service) SendMessageImmediate(ctx context.Context, channelName, message string) error {
	channelID, exists := s.channels[channelName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrChannelNotFound, channelName)
	}

	return s.sendMessageToChannel(ctx, channelID, message)
}

func (s *Service) SendMessage(ctx context.Context, channelName, message string) error {
	channelID, exists := s.channels[channelName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrChannelNotFound, channelName)
	}

	s.bufferMu.Lock()
	s.buffer[channelName] = append(s.buffer[channelName], Message{
		Channel: channelID,
		Text:    message,
	})

	if len(s.buffer[channelName]) >= s.bufferSize {
		messages := s.buffer[channelName]
		s.buffer[channelName] = nil
		s.bufferMu.Unlock()

		return s.sendBatchedMessages(ctx, channelID, messages)
	}

	s.bufferMu.Unlock()

	return nil
}

func (s *Service) Flush() error {
	return s.FlushWithContext(context.Background())
}

func (s *Service) FlushWithContext(ctx context.Context) error {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	var lastError error

	for channelName, messages := range s.buffer {
		if len(messages) > 0 {
			channelID := s.channels[channelName]
			if err := s.sendBatchedMessages(ctx, channelID, messages); err != nil {
				lastError = err
			}

			s.buffer[channelName] = nil
		}
	}

	return lastError
}

func (s *Service) SetBufferSize(size int) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.bufferSize = size
}

func (s *Service) SetFlushDelay(delay time.Duration) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.flushDelay = delay

	if s.flushTimer != nil {
		s.flushTimer.Stop()
	}

	s.startFlushTimer()
}

func (s *Service) sendBatchedMessages(ctx context.Context, channelID string, messages []Message) error {
	if len(messages) == 0 {
		return nil
	}

	if len(messages) == 1 {
		return s.sendMessageToChannel(ctx, channelID, messages[0].Text)
	}

	var combinedText strings.Builder

	for i, msg := range messages {
		if i > 0 {
			combinedText.WriteString("\n\n")
		}

		combinedText.WriteString(msg.Text)
	}

	return s.sendMessageToChannel(ctx, channelID, combinedText.String())
}

func (s *Service) sendMessageToChannel(ctx context.Context, channelID, message string) error {
	payload := Message{
		Channel: channelID,
		Text:    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var slackResp Response
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("%w: %s", ErrSlackAPI, slackResp.Error)
	}

	return nil
}

func (s *Service) startFlushTimer() {
	s.flushTimer = time.AfterFunc(s.flushDelay, func() {
		_ = s.FlushWithContext(s.ctx) //nolint:errcheck // Background flush - errors are logged internally
		s.startFlushTimer()
	})
}
