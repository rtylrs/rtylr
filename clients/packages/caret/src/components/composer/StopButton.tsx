import React from 'react';
import { Square } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { caretTokens } from '../../ui/tokens';

export interface StopButtonProps {
  onStop: () => void;
  size?: 'md' | 'lg';
}

/** Visible while a response is streaming. Aborts the in-flight request. */
export const StopButton: React.FC<StopButtonProps> = ({ onStop, size = 'md' }) => {
  const dim = size === 'lg' ? 36 : 32;
  return (
    <Tooltip label="Stop generating">
      <IconButton
        className="caret-send-button"
        size="md"
        label="Stop generating"
        onClick={onStop}
        style={{
          width: dim,
          height: dim,
          borderRadius: 999,
          background: caretTokens.surfaceHigh,
          color: caretTokens.text,
          border: `1px solid ${caretTokens.border}`,
        }}
      >
        <Square size={11} fill="currentColor" />
      </IconButton>
    </Tooltip>
  );
};

export default StopButton;
