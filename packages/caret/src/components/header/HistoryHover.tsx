import React from 'react';
import { History } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useCaretUIStore } from '../../stores/uiStore';

/** Opens the dedicated history view without covering the active conversation. */
export const HistoryHover: React.FC = () => {
  const setView = useCaretUIStore((state) => state.setView);
  const view = useCaretUIStore((state) => state.view);

  return (
    <Tooltip label="History">
      <IconButton
        className="caret-chrome-button"
        size="md"
        label="History"
        active={view === 'history'}
        onClick={() => setView('history')}
      >
        <History size={14} />
      </IconButton>
    </Tooltip>
  );
};

export default HistoryHover;
