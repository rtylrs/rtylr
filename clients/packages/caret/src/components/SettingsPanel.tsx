import React, { useId } from "react";
import { caretTokens, caretRadius } from "../ui/tokens";
import { useCaretSettingsStore } from "../stores/settingsStore";
import MemorySettings from "./MemorySettings";

const sectionTitle: React.CSSProperties = {
  fontSize: 10.5,
  fontWeight: 700,
  color: caretTokens.textDim,
  textTransform: "uppercase",
  letterSpacing: 0.7,
  margin: "0 0 7px 2px",
};

const description: React.CSSProperties = {
  color: caretTokens.textMuted,
  fontSize: 11.5,
  lineHeight: 1.45,
  marginTop: 2,
};

const label: React.CSSProperties = {
  color: caretTokens.text,
  fontSize: 13,
  fontWeight: 550,
};

interface SettingsToggleProps {
  checked: boolean;
  label: string;
  description: string;
  onChange: (checked: boolean) => void;
  divided?: boolean;
}

const SettingsToggle: React.FC<SettingsToggleProps> = ({
  checked,
  label: title,
  description: detail,
  onChange,
  divided = false,
}) => (
  <button
    type="button"
    role="switch"
    aria-checked={checked}
    className="caret-settings-row caret-switch"
    onClick={() => onChange(!checked)}
    style={{
      width: "100%",
      minHeight: 58,
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      gap: 14,
      padding: "10px 12px",
      background: "transparent",
      border: "none",
      borderTop: divided ? `1px solid ${caretTokens.border}` : "none",
      color: caretTokens.text,
      textAlign: "left",
      cursor: "pointer",
    }}
  >
    <span style={{ minWidth: 0 }}>
      <span style={{ ...label, display: "block" }}>{title}</span>
      <span style={{ ...description, display: "block" }}>{detail}</span>
    </span>
    <span
      aria-hidden="true"
      style={{
        width: 34,
        height: 20,
        padding: 2,
        borderRadius: 3,
        background: checked ? caretTokens.accent : caretTokens.active,
        boxSizing: "border-box",
        flexShrink: 0,
        transition: "background 0.16s ease",
      }}
    >
      <span
        className="caret-switch-thumb"
        style={{
          display: "block",
          width: 16,
          height: 16,
          borderRadius: 2,
          background: checked ? caretTokens.accentInk : caretTokens.surface,
          boxShadow: "none",
        }}
      />
    </span>
  </button>
);

export interface SettingsPanelProps {
  /** Enables the Memory section, which reads and edits saved memories. */
  userId?: number;
  apiBaseUrl?: string;
}

export const SettingsPanel: React.FC<SettingsPanelProps> = ({
  userId = 0,
  apiBaseUrl,
}) => {
  const s = useCaretSettingsStore();

  return (
    <div
      className="caret-settings-panel"
      style={{ padding: "18px 14px 20px", overflowY: "auto", flex: 1 }}
    >
      <div style={{ marginBottom: 20 }}>
        <h2
          style={{
            margin: 0,
            color: caretTokens.text,
            fontSize: 15,
            fontWeight: 650,
            lineHeight: 1.4,
          }}
        >
          Settings
        </h2>
        <p style={{ ...description, margin: "3px 0 0" }}>
          Tune how Caret responds and works.
        </p>
      </div>

      <section>
        <h3 style={sectionTitle}>Behavior</h3>
        <div
          style={{
            overflow: "hidden",
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.lg,
          }}
        >
          <SettingsToggle
            checked={s.autoApproveReadOnly}
            label="Auto-approve safe tools"
            description="Skip confirmation for actions that only read data."
            onChange={s.setAutoApproveReadOnly}
          />
          <SettingsToggle
            divided
            checked={s.showThinking}
            label="Show working"
            description="Display progress details while Caret is responding."
            onChange={s.setShowThinking}
          />
          <SettingsToggle
            divided
            checked={s.voiceEnabled}
            label="Voice input"
            description="Show dictation controls in the message composer."
            onChange={s.setVoiceEnabled}
          />
        </div>
      </section>

      {userId > 0 && apiBaseUrl !== undefined && (
        <section style={{ marginTop: 22 }}>
          <h3 style={sectionTitle}>Memory</h3>
          <p style={{ ...description, margin: "0 2px 8px" }}>
            Caret uses these in every chat. Correct or remove anything that is
            wrong.
          </p>
          <MemorySettings userId={userId} apiBaseUrl={apiBaseUrl} />
        </section>
      )}
    </div>
  );
};

export default SettingsPanel;
