//! Slack batching, mirroring `internal/platform/slack/slack.go`.
//!
//! Buffered per channel and flushed on an interval, combined into as few
//! `chat.postMessage` calls as possible. This crate only provides the
//! batching primitives; the composition root owns the flush loop.

pub const FLUSH_INTERVAL_SECS: u64 = 10;
pub const MAX_MESSAGE_CHARS: usize = 38000;
pub const MAX_SLACK_RESPONSE_BODY_BYTES: usize = 64 << 10;
pub const SEPARATOR: &str = "\n──────────\n";

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Channel {
    Alerts,
    Requests,
    Dashboard,
}

/// Join lines with the separator into as few strings as possible, each
/// under `limit`. A single oversized line is hard-truncated.
pub fn chunk(lines: &[String], limit: usize) -> Vec<String> {
    let mut out = Vec::new();
    let mut current = String::new();
    for line in lines {
        let mut text = line.clone();
        if text.len() > limit {
            let mut cut = limit.saturating_sub(3);
            while !text.is_char_boundary(cut) && cut > 0 {
                cut -= 1;
            }
            text = format!("{}...", &text[..cut]);
        }
        if !current.is_empty() && current.len() + SEPARATOR.len() + text.len() > limit {
            out.push(current.clone());
            current.clear();
        }
        if !current.is_empty() {
            current.push_str(SEPARATOR);
        }
        current.push_str(&text);
    }
    if !current.is_empty() {
        out.push(current);
    }
    out
}

#[derive(Debug, Clone, Default)]
pub struct Batcher {
    buffers: std::collections::HashMap<Channel, Vec<String>>,
}

impl Batcher {
    pub fn new() -> Self {
        Self {
            buffers: std::collections::HashMap::new(),
        }
    }

    pub fn queue(&mut self, channel: Channel, text: &str) {
        if text.is_empty() {
            return;
        }
        self.buffers
            .entry(channel)
            .or_default()
            .push(text.to_string());
    }

    /// Drain one channel into flushed chunks (bounded by the limit).
    pub fn drain(&mut self, channel: Channel) -> Vec<String> {
        let lines = self.buffers.remove(&channel).unwrap_or_default();
        if lines.is_empty() {
            Vec::new()
        } else {
            chunk(&lines, MAX_MESSAGE_CHARS)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::Batcher;
    use crate::slack::Channel;

    #[test]
    fn chunking_and_drain() {
        let mut batcher = Batcher::new();
        batcher.queue(Channel::Alerts, "first");
        batcher.queue(Channel::Alerts, "second");
        let chunks = batcher.drain(Channel::Alerts);
        assert_eq!(chunks.len(), 1);
        assert!(chunks[0].contains("first") && chunks[0].contains("second"));
        assert!(batcher.drain(Channel::Alerts).is_empty());
        let long = "x".repeat(40_000);
        let single = super::chunk(std::slice::from_ref(&long), super::MAX_MESSAGE_CHARS);
        assert_eq!(single.len(), 1);
        assert!(single[0].ends_with("..."));
    }
}
