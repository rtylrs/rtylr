import React from 'react';
import { IconBase } from './IconBase';

export const TerminalIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#3c3c43" to="#0c0c0f">
    <path
      d="M4.3 5.4a1.4 1.4 0 0 1 1.98-1.98l7.3 7.3a1.4 1.4 0 0 1 0 1.98l-7.3 7.3A1.4 1.4 0 0 1 4.3 18l6.3-6.3Z"
      fill="#4ade80"
    />
    <rect x="13.2" y="18.1" width="7.2" height="2.6" rx="1.3" fill="#4ade80" opacity="0.85" />
  </IconBase>
);
