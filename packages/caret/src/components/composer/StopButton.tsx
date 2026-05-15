import React from 'react';
import { Square } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';

export interface StopButtonProps {
  onStop: () => void;
}

/** Visible while a response is streaming. Aborts the in-flight request. */
export const StopButton: React.FC<StopButtonProps> = ({ onStop }) => (
  <Tooltip label="Stop generating">
    <IconButton size="md" variant="solid" label="Stop generating" onClick={onStop}>
      <Square size={11} fill="currentColor" />
    </IconButton>
  </Tooltip>
);

export default StopButton;
