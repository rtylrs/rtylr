import React from 'react';
import { caretTokens, caretRadius } from '../ui/tokens';
import Divider from '../ui/Divider';
import { useCaretSettingsStore } from '../stores/settingsStore';
import { CARET_MODES, type CaretMode } from '../types/mode';
import { CARET_MODELS, type CaretModel } from '../types/model';

const sectionTitle: React.CSSProperties = {
  fontSize: 11,
  fontWeight: 600,
  color: caretTokens.textDim,
  textTransform: 'uppercase',
  letterSpacing: 0.4,
  margin: '0 0 8px',
};

const row: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  padding: '8px 0',
};

const label: React.CSSProperties = {
  color: caretTokens.text,
  fontSize: 13,
};

const description: React.CSSProperties = {
  color: caretTokens.textMuted,
  fontSize: 12,
  marginTop: 2,
};

export const SettingsPanel: React.FC = () => {
  const s = useCaretSettingsStore();

  return (
    <div style={{ padding: 16, overflowY: 'auto', flex: 1 }}>
      <div>
        <h3 style={sectionTitle}>Mode</h3>
        <select
          value={s.mode}
          onChange={(e) => s.setMode(e.target.value as CaretMode)}
          style={{
            width: '100%',
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            color: caretTokens.text,
            fontSize: 13,
            padding: '8px 10px',
            borderRadius: caretRadius.md,
          }}
        >
          {Object.values(CARET_MODES).map((m) => (
            <option key={m.id} value={m.id}>
              {m.label} — {m.description}
            </option>
          ))}
        </select>
      </div>

      <Divider />

      <div style={{ paddingTop: 12 }}>
        <h3 style={sectionTitle}>Model</h3>
        <select
          value={s.model}
          onChange={(e) => s.setModel(e.target.value as CaretModel)}
          style={{
            width: '100%',
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            color: caretTokens.text,
            fontSize: 13,
            padding: '8px 10px',
            borderRadius: caretRadius.md,
          }}
        >
          {Object.values(CARET_MODELS).map((m) => (
            <option key={m.id} value={m.id}>
              {m.label}
            </option>
          ))}
        </select>
      </div>

      <Divider />

      <div style={{ paddingTop: 12 }}>
        <h3 style={sectionTitle}>Behavior</h3>
        <div style={row}>
          <div>
            <div style={label}>Auto-approve read-only tools</div>
            <div style={description}>Skip confirmation for tools that don't modify data.</div>
          </div>
          <input
            type="checkbox"
            checked={s.autoApproveReadOnly}
            onChange={(e) => s.setAutoApproveReadOnly(e.target.checked)}
          />
        </div>
        <div style={row}>
          <div>
            <div style={label}>Show thinking</div>
            <div style={description}>Reveal the model's reasoning blocks while it works.</div>
          </div>
          <input
            type="checkbox"
            checked={s.showThinking}
            onChange={(e) => s.setShowThinking(e.target.checked)}
          />
        </div>
        <div style={row}>
          <div>
            <div style={label}>Voice input</div>
            <div style={description}>Enable the dictation mic in the composer.</div>
          </div>
          <input
            type="checkbox"
            checked={s.voiceEnabled}
            onChange={(e) => s.setVoiceEnabled(e.target.checked)}
          />
        </div>
      </div>
    </div>
  );
};

export default SettingsPanel;
