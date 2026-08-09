import React from 'react';
import { ArrowUp } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { caretTokens } from '../../ui/tokens';

export interface SendButtonProps {
  disabled?: boolean;
  onClick: () => void;
}

export const SendButton: React.FC<SendButtonProps> = ({ disabled, onClick }) => (
  <Tooltip label="Send (Enter)">
    <IconButton
      size="md"
      className="caret-send-button"
      label="Send message"
      disabled={disabled}
      onClick={onClick}
      style={{
        width: 32,
        height: 32,
        borderRadius: 4,
        background: disabled ? caretTokens.surfaceHigh : caretTokens.accent,
        color: disabled ? caretTokens.textDim : caretTokens.accentInk,
      }}
    >
      <ArrowUp size={15} strokeWidth={2.3} />
    </IconButton>
  </Tooltip>
);

export default SendButton;
