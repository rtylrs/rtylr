import React from 'react';
import { Copy, RefreshCcw, Check } from 'lucide-react';
import { caretTokens } from '../ui/tokens';
import IconButton from '../ui/IconButton';
import Tooltip from '../ui/Tooltip';
import { useClipboard } from '../hooks/useClipboard';

export interface MessageActionsProps {
  /** Plain-text representation of the message for copy. */
  text: string;
  onRetry?: () => void;
}

/** Hover-revealed action row attached to assistant messages. */
export const MessageActions: React.FC<MessageActionsProps> = ({ text, onRetry }) => {
  const { copy, copied } = useClipboard();
  return (
    <div
      className="caret-message-actions"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 2,
        marginTop: 4,
        color: caretTokens.textDim,
      }}
    >
      <Tooltip label={copied ? 'Copied' : 'Copy'}>
        <IconButton size="sm" label="Copy message" onClick={() => copy(text)}>
          {copied ? <Check size={11} /> : <Copy size={11} />}
        </IconButton>
      </Tooltip>
      {onRetry && (
        <Tooltip label="Retry">
          <IconButton size="sm" label="Retry message" onClick={onRetry}>
            <RefreshCcw size={11} />
          </IconButton>
        </Tooltip>
      )}
    </div>
  );
};

export default MessageActions;
