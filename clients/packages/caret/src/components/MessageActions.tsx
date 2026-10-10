import React, { useEffect, useRef, useState } from "react";
import {
  Copy,
  RefreshCcw,
  Check,
  Volume2,
  VolumeX,
  ThumbsUp,
  ThumbsDown,
} from "lucide-react";
import { caretTokens } from "../ui/tokens";
import IconButton from "../ui/IconButton";
import Tooltip from "../ui/Tooltip";
import { useClipboard } from "../hooks/useClipboard";

export interface MessageActionsProps {
  /** Plain-text representation of the message for copy. */
  text: string;
  onRetry?: () => void;
  /** How long the reply took; shown only for replies timed in this session. */
  responseMs?: number;
  /** Rates the reply; Caret learns lessons from these verdicts. */
  onRate?: (rating: ReplyRating, reason?: string) => void;
}

export type ReplyRating = "up" | "down" | "clear";

/** "0.8s", "4.2s", "1m 05s". */
export function formatResponseTime(ms: number): string {
  const seconds = ms / 1000;
  if (seconds < 60)
    return `${seconds < 10 ? seconds.toFixed(1) : Math.round(seconds)}s`;
  const whole = Math.round(seconds);
  return `${Math.floor(whole / 60)}m ${String(whole % 60).padStart(2, "0")}s`;
}

/** Hover-revealed action row attached to assistant messages. */
export const MessageActions: React.FC<MessageActionsProps> = ({
  text,
  onRetry,
  responseMs,
  onRate,
}) => {
  const { copy, copied } = useClipboard();
  const [rating, setRating] = useState<"up" | "down" | null>(null);
  const [askReason, setAskReason] = useState(false);
  const [reason, setReason] = useState("");
  const rate = (next: "up" | "down"): void => {
    if (!onRate) return;
    if (rating === next) {
      setRating(null);
      setAskReason(false);
      onRate("clear");
      return;
    }
    setRating(next);
    setAskReason(next === "down");
    onRate(next);
  };
  const cancelled = useRef(false);
  const sendReason = (): void => {
    if (cancelled.current) {
      cancelled.current = false;
      return;
    }
    const text = reason.trim();
    setAskReason(false);
    if (text && onRate) onRate("down", text);
    setReason("");
  };
  const utterance = useRef<SpeechSynthesisUtterance | null>(null);
  const [reading, setReading] = useState(false);
  const canRead =
    typeof window !== "undefined" &&
    "speechSynthesis" in window &&
    typeof SpeechSynthesisUtterance !== "undefined";

  useEffect(
    () => () => {
      if (utterance.current) window.speechSynthesis.cancel();
    },
    [],
  );

  const toggleReadAloud = (): void => {
    if (!canRead) return;
    if (reading) {
      window.speechSynthesis.cancel();
      utterance.current = null;
      setReading(false);
      return;
    }
    const next = new SpeechSynthesisUtterance(text);
    next.onend = () => setReading(false);
    next.onerror = () => setReading(false);
    utterance.current = next;
    setReading(true);
    window.speechSynthesis.speak(next);
  };

  return (
    <div
      className="caret-message-actions"
      style={{
        display: "flex",
        alignItems: "center",
        gap: 2,
        marginTop: 4,
        color: caretTokens.textDim,
      }}
    >
      <Tooltip label={copied ? "Copied" : "Copy"}>
        <IconButton size="sm" label="Copy message" onClick={() => copy(text)}>
          {copied ? <Check size={11} /> : <Copy size={11} />}
        </IconButton>
      </Tooltip>
      <Tooltip
        label={
          !canRead
            ? "Read aloud is not supported in this browser"
            : reading
              ? "Stop reading"
              : "Read aloud"
        }
      >
        <IconButton
          size="sm"
          label={reading ? "Stop reading aloud" : "Read message aloud"}
          disabled={!canRead}
          onClick={toggleReadAloud}
        >
          {reading ? <VolumeX size={11} /> : <Volume2 size={11} />}
        </IconButton>
      </Tooltip>
      {onRate && (
        <>
          <Tooltip label={rating === "up" ? "Remove rating" : "Good reply"}>
            <IconButton
              size="sm"
              label="Good reply"
              aria-pressed={rating === "up"}
              onClick={() => rate("up")}
            >
              <ThumbsUp
                size={11}
                fill={rating === "up" ? "currentColor" : "none"}
              />
            </IconButton>
          </Tooltip>
          <Tooltip label={rating === "down" ? "Remove rating" : "Bad reply"}>
            <IconButton
              size="sm"
              label="Bad reply"
              aria-pressed={rating === "down"}
              onClick={() => rate("down")}
            >
              <ThumbsDown
                size={11}
                fill={rating === "down" ? "currentColor" : "none"}
              />
            </IconButton>
          </Tooltip>
        </>
      )}
      {onRetry && (
        <Tooltip label="Retry">
          <IconButton size="sm" label="Retry message" onClick={onRetry}>
            <RefreshCcw size={11} />
          </IconButton>
        </Tooltip>
      )}
      {responseMs !== undefined && (
        <Tooltip label="Time to reply">
          <span
            aria-label={`Replied in ${formatResponseTime(responseMs)}`}
            style={{
              fontSize: 11,
              fontVariantNumeric: "tabular-nums",
              padding: "0 6px",
              color: caretTokens.textDim,
            }}
          >
            {formatResponseTime(responseMs)}
          </span>
        </Tooltip>
      )}
      {askReason && (
        <input
          autoFocus
          aria-label="What was wrong with this reply?"
          placeholder="What was wrong? Caret will learn from it"
          value={reason}
          maxLength={1000}
          onChange={(event) => setReason(event.target.value)}
          onKeyDown={(event) => {
            // Blurring sends the reason, so Enter and leaving the field
            // both submit it exactly once.
            if (event.key === "Enter") event.currentTarget.blur();
            if (event.key === "Escape") {
              cancelled.current = true;
              setReason("");
              setAskReason(false);
            }
          }}
          onBlur={sendReason}
          style={{
            flex: "1 1 180px",
            minWidth: 0,
            marginLeft: 6,
            padding: "3px 8px",
            fontSize: 12,
            color: caretTokens.text,
            background: "transparent",
            border: `1px solid ${caretTokens.border}`,
            borderRadius: 6,
            outline: "none",
          }}
        />
      )}
    </div>
  );
};

export default MessageActions;
