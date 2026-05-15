import React from 'react';
import { Send } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';

export interface SendButtonProps {
  disabled?: boolean;
  onClick: () => void;
}

export const SendButton: React.FC<SendButtonProps> = ({ disabled, onClick }) => (
  <Tooltip label="Send (Enter)">
    <IconButton
      size="md"
      variant={disabled ? 'soft' : 'solid'}
      label="Send message"
      disabled={disabled}
      onClick={onClick}
    >
      <Send size={13} />
    </IconButton>
  </Tooltip>
);

export default SendButton;
