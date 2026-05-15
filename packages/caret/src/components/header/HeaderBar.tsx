import React from 'react';
import { Sparkles, History, Settings, X, Plus } from 'lucide-react';
import { caretTokens } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useCaretStore } from '../../stores/caretStore';
import { useCaretUIStore, type CaretView } from '../../stores/uiStore';

export interface HeaderBarProps {
  onNewChat?: () => void;
}

const TABS: Array<{ id: CaretView; icon: React.ReactNode; label: string }> = [
  { id: 'chat', icon: <Sparkles size={13} />, label: 'Chat' },
  { id: 'history', icon: <History size={13} />, label: 'History' },
  { id: 'settings', icon: <Settings size={13} />, label: 'Settings' },
];

export const HeaderBar: React.FC<HeaderBarProps> = ({ onNewChat }) => {
  const close = useCaretStore((s) => s.close);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);

  return (
    <div
      style={{
        height: 44,
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        padding: '0 8px',
        borderBottom: `1px solid ${caretTokens.border}`,
        flexShrink: 0,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '0 4px' }}>
        <Sparkles size={14} style={{ color: caretTokens.accent }} />
        <span style={{ color: caretTokens.text, fontSize: 13, fontWeight: 600 }}>Caret</span>
      </div>

      <span style={{ flex: 1 }} />

      {TABS.map((t) => (
        <Tooltip key={t.id} label={t.label}>
          <IconButton
            size="md"
            label={t.label}
            active={view === t.id}
            onClick={() => setView(t.id)}
          >
            {t.icon}
          </IconButton>
        </Tooltip>
      ))}

      {onNewChat && (
        <Tooltip label="New chat">
          <IconButton size="md" label="New chat" onClick={onNewChat}>
            <Plus size={13} />
          </IconButton>
        </Tooltip>
      )}

      <Tooltip label="Close">
        <IconButton size="md" label="Close Caret" onClick={close}>
          <X size={13} />
        </IconButton>
      </Tooltip>
    </div>
  );
};

export default HeaderBar;
