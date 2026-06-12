import React from 'react';
import { IconBase } from './IconBase';

export const CRMIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#c585fb" to="#7619c4">
    <circle cx="9" cy="7.8" r="3.6" fill="#fff" />
    <path
      d="M2.8 19.2c0-3.5 2.8-5.9 6.2-5.9s6.2 2.4 6.2 5.9v.2a1 1 0 0 1-1 1H3.8a1 1 0 0 1-1-1Z"
      fill="#fff"
    />
    <circle cx="16.6" cy="7.2" r="2.8" fill="#fff" opacity="0.55" />
    <path
      d="M15.3 13.5c3.3.4 5.9 2.6 5.9 5.7v.2a1 1 0 0 1-1 1h-3.3c.1-.4.2-.8.2-1.2 0-2.2-.65-4.1-1.8-5.7Z"
      fill="#fff"
      opacity="0.55"
    />
  </IconBase>
);
