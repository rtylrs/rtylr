import React, { useEffect, useRef, useState } from "react";
import { Copy, RefreshCcw, Check, Volume2, VolumeX } from "lucide-react";
import { caretTokens } from "../ui/tokens";
import IconButton from "../ui/IconButton";
import Tooltip from "../ui/Tooltip";
import { useClipboard } from "../hooks/useClipboard";

export interface MessageActionsProps {
  /** Plain-text representation of the message for copy. */
  text: string;
  onRetry?: () => void;
}

/** Hover-revealed action row attached to assistant messages. */
export const MessageActions: React.FC<MessageActionsProps> = ({
  text,
  onRetry,
}) => {
  const { copy, copied } = useClipboard();
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
      {onRetry && (
        <Tooltip label="Retry">
          <IconButton size="sm" label="Retry message" onClick={onRetry}>
            <RefreshCcw size={11} />
          </IconButton>
        </Tooltip>
      )}
    </div>
  );
};

export default MessageActions;
