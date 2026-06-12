import React from 'react';
import { IconBase } from './IconBase';

export const CRMIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#c585fb" to="#7619c4">
    <circle cx="16.6" cy="7.2" r="2.8" opacity="0.5" />
    <path
      d="M15.3 13.5c3.3.4 5.9 2.6 5.9 5.7v.2a1 1 0 0 1-1 1h-3.3c.1-.4.2-.8.2-1.2 0-2.2-.65-4.1-1.8-5.7Z"
      opacity="0.5"
    />
    <circle cx="9" cy="7.8" r="3.6" />
    <path d="M2.8 19.2c0-3.5 2.8-5.9 6.2-5.9s6.2 2.4 6.2 5.9v.2a1 1 0 0 1-1 1H3.8a1 1 0 0 1-1-1Z" />
    <path
      d="M18.6 2.6c.9-.95 2.4-.95 3.3 0 .9.94.9 2.47 0 3.41l-2.95 3.09a.5.5 0 0 1-.72 0L15.3 6c-.9-.94-.9-2.47 0-3.41.9-.95 2.4-.95 3.3 0Z"
      fill="#ff8fab"
      transform="translate(-1.2 0.4) scale(0.78)"
    />
  </IconBase>
);
