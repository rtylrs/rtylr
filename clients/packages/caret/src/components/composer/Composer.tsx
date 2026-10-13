import React, {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import { X } from "lucide-react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import ApprovalPicker from "./ApprovalPicker";
import MicButton, { type VoiceInputStatus } from "./MicButton";
import AttachButton from "./AttachButton";
import FileChip from "./FileChip";
import {
  IMAGE_ACCEPT,
  MAX_TEXT_FILES,
  TEXT_FILE_ACCEPT,
  isImageFile,
  isTextFile,
} from "../../lib/fileAttachments";
import SendButton from "./SendButton";
import StopButton from "./StopButton";
import { useAutosize } from "../../hooks/useAutosize";
import { CARET_FOCUS_COMPOSER_EVENT } from "../../lib/focusComposer";
import {
  filterCaretSlashCommands,
  type CaretSlashCommand,
} from "../../lib/slashCommands";
import { useCaretSettingsStore } from "../../stores/settingsStore";

export interface ComposerProps {
  isStreaming: boolean;
  onSubmit: (text: string, images: File[]) => void;
  onStop: () => void;
  onAttach?: (files: File[]) => void;
  placeholder?: string;
  autoFocus?: boolean;
  /** 'lg' renders a roomier console for full-screen/home contexts. Defaults to 'md'. */
  size?: "md" | "lg";
}

function voiceErrorMessage(error: string): string {
  switch (error) {
    case "not-allowed":
    case "service-not-allowed":
      return "Microphone access is blocked. Allow it in your browser, then try again.";
    case "audio-capture":
      return "No microphone was found. Check your input device and try again.";
    case "no-speech":
      return "I didn't catch that. Try dictating again.";
    case "network":
      return "Voice input lost its connection. Try again.";
    default:
      return "Voice input stopped unexpectedly. Try again.";
  }
}

const LARGE_COMPOSER_SHADOW =
  "0 1px 2px rgba(15, 15, 20, 0.05), 0 6px 16px -4px rgba(15, 15, 20, 0.10), 0 12px 24px -10px rgba(15, 15, 20, 0.14)";

export const Composer: React.FC<ComposerProps> = ({
  isStreaming,
  onSubmit,
  onStop,
  onAttach,
  placeholder = "Ask Caret anything…",
  autoFocus = true,
  size = "md",
}) => {
  const isLarge = size === "lg";
  const [value, setValue] = useState("");
  const [slashCommandIndex, setSlashCommandIndex] = useState(0);
  const [voiceStatus, setVoiceStatus] = useState<VoiceInputStatus | null>(null);
  const [attachments, setAttachments] = useState<File[]>([]);
  const [attachmentError, setAttachmentError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);
  const [previews, setPreviews] = useState<Array<{ file: File; url: string }>>(
    [],
  );
  const taRef = useRef<HTMLTextAreaElement>(null);
  const voiceFeedbackId = useId();
  const voiceEnabled = useCaretSettingsStore((s) => s.voiceEnabled);
  useAutosize(taRef, value);
  const slashCommands = useMemo(() => filterCaretSlashCommands(value), [value]);

  useEffect(() => {
    const next = attachments.map((file) => ({
      file,
      url: URL.createObjectURL(file),
    }));
    setPreviews(next);
    return () => next.forEach(({ url }) => URL.revokeObjectURL(url));
  }, [attachments]);

  const addAttachments = useCallback(
    (files: File[]): void => {
      const accepted = files.filter(
        (file) => isImageFile(file) || isTextFile(file),
      );
      if (accepted.length !== files.length)
        setAttachmentError(
          "Choose JPEG, PNG or WebP images, PDF, Word, Excel or PowerPoint files, or text files such as CSV.",
        );
      else setAttachmentError(null);
      if (accepted.length === 0) return;
      setAttachments((current) => {
        const next = [...current, ...accepted];
        const images = next.filter(isImageFile);
        const texts = next.filter((file) => !isImageFile(file));
        if (images.length > 4 || texts.length > MAX_TEXT_FILES) {
          setAttachmentError(
            `Attach up to four images and ${MAX_TEXT_FILES} files per message.`,
          );
          return [...images.slice(0, 4), ...texts.slice(0, MAX_TEXT_FILES)];
        }
        return next;
      });
      onAttach?.(accepted);
    },
    [onAttach],
  );

  useEffect(() => {
    setSlashCommandIndex((current) =>
      slashCommands.length === 0
        ? 0
        : Math.min(current, slashCommands.length - 1),
    );
  }, [slashCommands.length]);

  const focusComposer = useCallback((): void => {
    const el = taRef.current;
    if (!el) return;
    el.focus();
    const cursor = el.value.length;
    el.setSelectionRange(cursor, cursor);
  }, []);

  useEffect(() => {
    if (autoFocus) focusComposer();
  }, [autoFocus, focusComposer]);

  useEffect(() => {
    window.addEventListener(CARET_FOCUS_COMPOSER_EVENT, focusComposer);
    return (): void =>
      window.removeEventListener(CARET_FOCUS_COMPOSER_EVENT, focusComposer);
  }, [focusComposer]);

  const send = useCallback((): void => {
    const trimmed = value.trim();
    if ((!trimmed && attachments.length === 0) || isStreaming) return;
    onSubmit(trimmed, attachments);
    setValue("");
    setAttachments([]);
    setAttachmentError(null);
  }, [value, attachments, isStreaming, onSubmit]);

  const appendTranscript = useCallback((text: string): void => {
    const normalized = text.trim();
    if (!normalized) return;
    setValue(
      (prev) =>
        (prev.length && !prev.endsWith(" ") ? `${prev} ` : prev) + normalized,
    );
  }, []);

  const selectSlashCommand = useCallback(
    (command: CaretSlashCommand): void => {
      setValue(`/${command.name} `);
      setSlashCommandIndex(0);
      requestAnimationFrame(() => focusComposer());
    },
    [focusComposer],
  );

  const updateVoiceStatus = useCallback((status: VoiceInputStatus): void => {
    setVoiceStatus(status);
  }, []);

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (slashCommands.length > 0 && e.key === "ArrowDown") {
      e.preventDefault();
      setSlashCommandIndex((current) => (current + 1) % slashCommands.length);
      return;
    }
    if (slashCommands.length > 0 && e.key === "ArrowUp") {
      e.preventDefault();
      setSlashCommandIndex(
        (current) =>
          (current - 1 + slashCommands.length) % slashCommands.length,
      );
      return;
    }
    if (slashCommands.length > 0 && e.key === "Tab") {
      e.preventDefault();
      selectSlashCommand(slashCommands[slashCommandIndex]);
      return;
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  };

  return (
    <div
      className={isLarge ? "caret-composer is-large" : "caret-composer"}
      onDragEnter={(event) => {
        event.preventDefault();
        setDragging(true);
      }}
      onDragOver={(event) => event.preventDefault()}
      onDragLeave={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null))
          setDragging(false);
      }}
      onDrop={(event) => {
        event.preventDefault();
        setDragging(false);
        addAttachments(Array.from(event.dataTransfer.files));
      }}
      style={{
        margin: isLarge ? "0" : "10px 10px 11px",
        background: isLarge ? caretTokens.surface : caretTokens.surfaceRaised,
        border: `1px solid ${caretTokens.border}`,
        borderRadius: isLarge ? 28 : caretRadius.xl,
        padding: isLarge ? "14px 14px 12px 20px" : "8px 8px 8px 12px",
        display: "flex",
        flexDirection: "column",
        gap: isLarge ? 8 : 6,
        boxShadow: isLarge ? LARGE_COMPOSER_SHADOW : "none",
        outline: dragging ? `2px solid ${caretTokens.accent}` : undefined,
      }}
    >
      {previews.length > 0 && (
        <div
          aria-label="Attachments"
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: 7,
            padding: "4px 4px 0",
          }}
        >
          {previews.map(({ file, url }, index) => (
            <div
              key={`${file.name}-${file.lastModified}-${index}`}
              style={{ position: "relative" }}
            >
              {isImageFile(file) ? (
                <img
                  src={url}
                  alt={file.name}
                  style={{
                    width: 64,
                    height: 64,
                    objectFit: "cover",
                    borderRadius: 9,
                    border: `1px solid ${caretTokens.border}`,
                  }}
                />
              ) : (
                <FileChip name={file.name} />
              )}
              <button
                type="button"
                aria-label={`Remove ${file.name}`}
                onClick={() =>
                  setAttachments((items) => items.filter((_, i) => i !== index))
                }
                style={{
                  position: "absolute",
                  right: -5,
                  top: -5,
                  width: 20,
                  height: 20,
                  border: `1px solid ${caretTokens.border}`,
                  borderRadius: 10,
                  background: caretTokens.surface,
                  color: caretTokens.text,
                  display: "grid",
                  placeItems: "center",
                  cursor: "pointer",
                }}
              >
                <X size={11} />
              </button>
            </div>
          ))}
        </div>
      )}
      {attachmentError && (
        <div
          role="alert"
          style={{
            color: caretTokens.dangerText,
            fontSize: 12,
            padding: "0 4px",
          }}
        >
          {attachmentError}
        </div>
      )}
      {slashCommands.length > 0 && (
        <div
          role="listbox"
          aria-label="Caret commands"
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 2,
            margin: "0 0 2px",
            padding: 4,
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.lg,
            background: caretTokens.surfaceHigh,
          }}
        >
          {slashCommands.map((command, index) => (
            <button
              key={command.name}
              type="button"
              role="option"
              aria-selected={index === slashCommandIndex}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => selectSlashCommand(command)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 9,
                width: "100%",
                minWidth: 0,
                padding: "7px 8px",
                border: "none",
                borderRadius: caretRadius.md,
                background:
                  index === slashCommandIndex
                    ? caretTokens.hover
                    : "transparent",
                color: caretTokens.text,
                textAlign: "left",
                cursor: "pointer",
              }}
            >
              <code style={{ color: caretTokens.accentStrong, fontSize: 12 }}>
                /{command.name}
              </code>
              <span
                style={{
                  minWidth: 0,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  color: caretTokens.textMuted,
                  fontSize: 12,
                }}
              >
                {command.description}
              </span>
            </button>
          ))}
        </div>
      )}
      {(voiceStatus?.listening || voiceStatus?.error) && (
        <div
          id={voiceFeedbackId}
          role="status"
          aria-live="polite"
          className={
            voiceStatus.error
              ? "caret-voice-status is-error"
              : "caret-voice-status"
          }
          style={{
            display: "flex",
            alignItems: "center",
            gap: 7,
            minWidth: 0,
            margin: "0 4px 0 0",
            padding: "6px 8px",
            borderRadius: caretRadius.md,
            background: voiceStatus.error
              ? caretTokens.dangerSoft
              : caretTokens.accentSoft,
            color: voiceStatus.error
              ? caretTokens.dangerText
              : caretTokens.accentStrong,
            fontSize: 11.5,
            lineHeight: 1.35,
          }}
        >
          <span
            className={
              voiceStatus.listening ? "caret-listening-dot" : undefined
            }
            aria-hidden="true"
            style={{
              width: 7,
              height: 7,
              borderRadius: "50%",
              background: voiceStatus.error
                ? caretTokens.danger
                : caretTokens.accent,
              flexShrink: 0,
            }}
          />
          <span
            style={{
              minWidth: 0,
              overflow: "hidden",
              textOverflow: "ellipsis",
              flex: 1,
            }}
          >
            {voiceStatus.error
              ? voiceErrorMessage(voiceStatus.error)
              : voiceStatus.interimTranscript.trim()
                ? `Listening · ${voiceStatus.interimTranscript.trim()}`
                : "Listening… speak naturally, then tap the mic to stop."}
          </span>
          {voiceStatus.error && (
            <button
              type="button"
              aria-label="Dismiss voice message"
              onClick={() => setVoiceStatus(null)}
              style={{
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                width: 24,
                height: 24,
                margin: -4,
                border: "none",
                borderRadius: caretRadius.sm,
                background: "transparent",
                color: caretTokens.dangerText,
                cursor: "pointer",
                flexShrink: 0,
              }}
            >
              <X size={12} />
            </button>
          )}
        </div>
      )}
      <div
        className="caret-composer-row"
        style={{
          display: "flex",
          flexDirection: isLarge ? "column" : "row",
          alignItems: isLarge ? "stretch" : "flex-end",
          minWidth: 0,
          gap: isLarge ? 6 : 5,
        }}
      >
        <textarea
          className="caret-composer-input"
          ref={taRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onPaste={(event) => {
            const files = Array.from(event.clipboardData.items)
              .map((item) => (item.kind === "file" ? item.getAsFile() : null))
              .filter((file): file is File => file !== null);
            if (files.length > 0) addAttachments(files);
          }}
          onKeyDown={handleKey}
          placeholder={placeholder}
          rows={1}
          aria-label="Message Caret"
          aria-describedby={
            voiceStatus?.listening || voiceStatus?.error
              ? voiceFeedbackId
              : undefined
          }
          spellCheck
          style={{
            resize: "none",
            width: "100%",
            minWidth: 0,
            background: "transparent",
            border: "none",
            outline: "none",
            color: caretTokens.text,
            fontSize: isLarge ? 16 : 14,
            lineHeight: isLarge ? "24px" : "20px",
            padding: isLarge ? "4px 0" : "6px 3px",
            fontFamily: "inherit",
            height: isLarge ? 56 : 32,
            minHeight: isLarge ? 56 : 32,
            maxHeight: isLarge ? 200 : 140,
            overflowY: "auto",
            boxSizing: "border-box",
            display: "block",
            flex: 1,
            whiteSpace: "pre-wrap",
            overflowWrap: "anywhere",
          }}
        />
        <div
          className="caret-composer-actions"
          style={{
            display: "flex",
            alignItems: "center",
            gap: isLarge ? 6 : 3,
            flexShrink: 0,
            width: isLarge ? "100%" : undefined,
          }}
        >
          <ApprovalPicker />
          <span
            style={{
              display: "flex",
              order: isLarge ? -1 : 0,
              marginInlineEnd: isLarge ? "auto" : 0,
            }}
          >
            <AttachButton
              onFiles={addAttachments}
              accept={`${IMAGE_ACCEPT},${TEXT_FILE_ACCEPT}`}
            />
          </span>
          {voiceEnabled && (
            <MicButton
              onTranscript={appendTranscript}
              onStatusChange={updateVoiceStatus}
            />
          )}
          {isStreaming ? (
            <StopButton onStop={onStop} size={size} />
          ) : (
            <SendButton
              disabled={!value.trim() && attachments.length === 0}
              onClick={send}
              size={size}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default Composer;
