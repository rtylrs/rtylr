import React from 'react';
import { IconBase } from './IconBase';

export const TerminalIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#3f3f46" to="#131316" stroke="#4ade80">
    <path d="m4 17 6-6-6-6" />
    <path d="M12 19h8" />
  </IconBase>
);
