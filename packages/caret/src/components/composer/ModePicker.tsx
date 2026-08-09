import React from 'react';
import { Compass, Wand2, BarChart3, LifeBuoy } from 'lucide-react';
import Menu from '../../ui/Menu';
import Pill from '../../ui/Pill';
import IconButton from '../../ui/IconButton';
import { useCaretSettingsStore } from '../../stores/settingsStore';
import { CARET_MODES, type CaretMode } from '../../types/mode';

const MODE_ICON: Record<CaretMode, React.ReactNode> = {
  ask: <Compass size={11} />,
  agent: <Wand2 size={11} />,
  analyst: <BarChart3 size={11} />,
  support: <LifeBuoy size={11} />,
};

export interface ModePickerProps {
  compact?: boolean;
}

export const ModePicker: React.FC<ModePickerProps> = ({ compact = false }) => {
  const mode = useCaretSettingsStore((s) => s.mode);
  const setMode = useCaretSettingsStore((s) => s.setMode);
  const def = CARET_MODES[mode];

  return (
    <Menu
      side="top"
      align="end"
      minWidth={168}
      trigger={
        compact ? (
          <IconButton
            className="caret-composer-icon"
            size="md"
            label={`Mode: ${def.label}`}
            aria-haspopup="menu"
          >
            {MODE_ICON[mode]}
          </IconButton>
        ) : (
          <Pill icon={MODE_ICON[mode]} aria-haspopup="menu">
            {def.label}
          </Pill>
        )
      }
      items={Object.values(CARET_MODES).map((m) => ({
        id: m.id,
        label: m.label,
        description: compact ? undefined : m.description,
        icon: MODE_ICON[m.id],
        selected: m.id === mode,
        onSelect: (): void => setMode(m.id),
      }))}
    />
  );
};

export default ModePicker;
