import React, { useId } from 'react';
import { ChevronDown } from 'lucide-react';
import { caretTokens, caretRadius } from '../ui/tokens';
import { useCaretSettingsStore } from '../stores/settingsStore';
import { CARET_MODES, type CaretMode } from '../types/mode';

const sectionTitle: React.CSSProperties = {
  fontSize: 10.5,
  fontWeight: 700,
  color: caretTokens.textDim,
  textTransform: 'uppercase',
  letterSpacing: 0.7,
  margin: '0 0 7px 2px',
};

const label: React.CSSProperties = {
  color: caretTokens.text,
  fontSize: 13,
  fontWeight: 550,
};

const description: React.CSSProperties = {
  color: caretTokens.textMuted,
  fontSize: 11.5,
  lineHeight: 1.45,
  marginTop: 2,
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
      width: '100%',
      minHeight: 58,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: 14,
      padding: '10px 12px',
      background: 'transparent',
      border: 'none',
      borderTop: divided ? `1px solid ${caretTokens.border}` : 'none',
      color: caretTokens.text,
      textAlign: 'left',
      cursor: 'pointer',
    }}
  >
    <span style={{ minWidth: 0 }}>
      <span style={{ ...label, display: 'block' }}>{title}</span>
      <span style={{ ...description, display: 'block' }}>{detail}</span>
    </span>
    <span
      aria-hidden="true"
      style={{
        width: 34,
        height: 20,
        padding: 2,
        borderRadius: 3,
        background: checked ? caretTokens.accent : caretTokens.active,
        boxSizing: 'border-box',
        flexShrink: 0,
        transition: 'background 0.16s ease',
      }}
    >
      <span
        className="caret-switch-thumb"
        style={{
          display: 'block',
          width: 16,
          height: 16,
          borderRadius: 2,
          background: checked ? caretTokens.accentInk : caretTokens.surface,
          boxShadow: 'none',
        }}
      />
    </span>
  </button>
);

export const SettingsPanel: React.FC = () => {
  const s = useCaretSettingsStore();
  const modeSelectId = useId();
  const selectedMode = CARET_MODES[s.mode];

  return (
    <div
      className="caret-settings-panel"
      style={{ padding: '18px 14px 20px', overflowY: 'auto', flex: 1 }}
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
        <p style={{ ...description, margin: '3px 0 0' }}>Tune how Caret responds and works.</p>
      </div>

      <section style={{ marginBottom: 18 }}>
        <h3 style={sectionTitle}>Default mode</h3>
        <div
          style={{
            padding: '2px 0 0',
          }}
        >
          <label htmlFor={modeSelectId} style={{ ...label, display: 'block', marginBottom: 7 }}>
            Response style
          </label>
          <div
            className="caret-settings-select-wrap"
            style={{
              position: 'relative',
              display: 'flex',
              alignItems: 'center',
              background: caretTokens.inputBg,
              border: `1px solid ${caretTokens.border}`,
              borderRadius: caretRadius.md,
              transition: 'border-color 0.14s ease, box-shadow 0.14s ease',
            }}
          >
            <select
              id={modeSelectId}
              value={s.mode}
              onChange={(event) => s.setMode(event.target.value as CaretMode)}
              style={{
                width: '100%',
                appearance: 'none',
                WebkitAppearance: 'none',
                background: 'transparent',
                border: 'none',
                outline: 'none',
                color: caretTokens.text,
                fontSize: 12.5,
                fontWeight: 550,
                padding: '8px 34px 8px 10px',
                cursor: 'pointer',
              }}
            >
              {Object.values(CARET_MODES).map((mode) => (
                <option key={mode.id} value={mode.id}>
                  {mode.label}
                </option>
              ))}
            </select>
            <ChevronDown
              size={13}
              aria-hidden="true"
              style={{
                position: 'absolute',
                right: 10,
                color: caretTokens.textMuted,
                pointerEvents: 'none',
              }}
            />
          </div>
          <p style={{ ...description, margin: '7px 1px 0' }}>{selectedMode.description}</p>
        </div>
      </section>

      <section>
        <h3 style={sectionTitle}>Behavior</h3>
        <div
          style={{
            overflow: 'hidden',
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
    </div>
  );
};

export default SettingsPanel;
