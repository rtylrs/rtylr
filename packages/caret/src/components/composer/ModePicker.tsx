import React from 'react';
import Menu from '../../ui/Menu';
import Pill from '../../ui/Pill';
import IconButton from '../../ui/IconButton';
import { useCaretSettingsStore } from '../../stores/settingsStore';
import { CARET_MODES } from '../../types/mode';
import { CARET_MODE_ICONS as MODE_ICON } from '../../lib/modeIcons';

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
      align={compact ? 'end' : 'start'}
      minWidth={240}
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
