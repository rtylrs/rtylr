import React from 'react';
import { ArrowUp } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { caretTokens } from '../../ui/tokens';

export interface SendButtonProps {
  disabled?: boolean;
  onClick: () => void;
  size?: 'md' | 'lg';
}

export const SendButton: React.FC<SendButtonProps> = ({ disabled, onClick, size = 'md' }) => {
  const dim = size === 'lg' ? 36 : 32;
  return (
    <Tooltip label="Send (Enter)">
      <IconButton
        size="md"
        className="caret-send-button"
        label="Send message"
        disabled={disabled}
        onClick={onClick}
        style={{
          width: dim,
          height: dim,
          borderRadius: 999,
          background: disabled ? caretTokens.surfaceHigh : caretTokens.accent,
          color: disabled ? caretTokens.textDim : caretTokens.accentInk,
        }}
      >
        <ArrowUp size={size === 'lg' ? 17 : 15} strokeWidth={2.3} />
      </IconButton>
    </Tooltip>
  );
};

export default SendButton;
