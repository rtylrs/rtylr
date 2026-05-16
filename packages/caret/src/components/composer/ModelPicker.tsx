import React from 'react';
import { Cpu } from 'lucide-react';
import Menu from '../../ui/Menu';
import Pill from '../../ui/Pill';
import { useCaretSettingsStore } from '../../stores/settingsStore';
import { CARET_MODELS } from '../../types/model';

export const ModelPicker: React.FC = () => {
  const model = useCaretSettingsStore((s) => s.model);
  const setModel = useCaretSettingsStore((s) => s.setModel);
  const def = CARET_MODELS[model];

  return (
    <Menu
      side="top"
      trigger={
        <Pill icon={<Cpu size={11} />} aria-haspopup="menu">
          {def.label}
        </Pill>
      }
      items={Object.values(CARET_MODELS).map((m) => ({
        id: m.id,
        label: m.label,
        description: m.description,
        onSelect: (): void => setModel(m.id),
      }))}
    />
  );
};

export default ModelPicker;
