// Package slack posts batched log messages to Slack channels. It is wired into
// the global logger (error-level records -> alerts channel) and called directly
// by the HTTP middlewares (request logs, blocked/suspicious requests).
//
// Messages are buffered per channel and flushed on an interval, combined into as
// few chat.postMessage calls as possible, so high log volume never trips Slack's
// rate limits.
//
// IMPORTANT: this package must never call platform/logger — the logger forwards
// errors here, so logging a Slack failure would loop. Internal errors go to
// stderr only.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

const (
	flushInterval = 10 * time.Second
	// Slack's per-message text limit is ~40k chars; leave headroom.
	maxMessageChars = 38000
	separator       = "\n──────────\n"
)

// Channel identifies a logical destination, mapped to a Slack channel ID at init.
type Channel int

const (
	Alerts Channel = iota
	Requests
	Suspicious
	DashboardCh
)

type notifier struct {
	token    string
	channels map[Channel]string
	client   *http.Client

	mu   sync.Mutex
	bufs map[Channel][]string
}

var (
	global *notifier
	mu     sync.RWMutex
)

// Init configures the notifier from config and starts the flush loop. It is a
// no-op (Slack stays disabled) unless SLACK_LOGGING_ENABLED is true and a bot
// token is set. Safe to call from each binary's startup.
func Init() {
	cfg := config.GlobalConfig
	if cfg == nil || !cfg.SlackLoggingEnabled || cfg.SlackBotToken == "" {
		return
	}
	n := &notifier{
		token: cfg.SlackBotToken,
		channels: map[Channel]string{
			Alerts:      cfg.SlackAlertsChannel,
			Requests:    cfg.SlackRequestLogsChannel,
			Suspicious:  cfg.SlackSuspiciousChannel,
			DashboardCh: cfg.SlackDashboardChannel,
		},
		client: &http.Client{Timeout: 10 * time.Second},
		bufs:   make(map[Channel][]string),
	}
	mu.Lock()
	global = n
	mu.Unlock()
	go n.loop()
}

// Enabled reports whether Slack logging is active.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return global != nil
}

// Post queues a message for the given channel. It returns immediately; delivery
// happens on the next flush. No-op when Slack is disabled or the channel is unset.
func Post(ch Channel, text string) {
	if text == "" {
		return
	}
	mu.RLock()
	n := global
	mu.RUnlock()
	if n == nil || n.channels[ch] == "" {
		return
	}
	n.mu.Lock()
	n.bufs[ch] = append(n.bufs[ch], text)
	n.mu.Unlock()
}

func (n *notifier) loop() {
	t := time.NewTicker(flushInterval)
	defer t.Stop()
	for range t.C {
		n.flush()
	}
}

func (n *notifier) flush() {
	n.mu.Lock()
	pending := n.bufs
	n.bufs = make(map[Channel][]string)
	n.mu.Unlock()

	for ch, lines := range pending {
		channelID := n.channels[ch]
		if channelID == "" || len(lines) == 0 {
			continue
		}
		for _, chunk := range chunk(lines, maxMessageChars) {
			n.post(channelID, chunk)
		}
	}
}

// chunk joins lines with a separator into as few strings as possible, each under
// limit. A single oversized line is hard-truncated.
func chunk(lines []string, limit int) []string {
	var out []string
	var cur bytes.Buffer
	for _, line := range lines {
		if len(line) > limit {
			line = line[:limit-3] + "..."
		}
		if cur.Len() > 0 && cur.Len()+len(separator)+len(line) > limit {
			out = append(out, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteString(separator)
		}
		cur.WriteString(line)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func (n *notifier) post(channelID, text string) {
	body, err := json.Marshal(map[string]any{"channel": channelID, "text": text})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+n.token)

	resp, err := n.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "slack: post failed: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Slack returns 200 with {"ok":false,"error":"..."} on logical failures.
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if jsonErr := json.NewDecoder(resp.Body).Decode(&result); jsonErr == nil && !result.OK {
		fmt.Fprintf(os.Stderr, "slack: chat.postMessage error: %s\n", result.Error)
	}
}
