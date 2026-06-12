import React from 'react';
import { IconBase } from './IconBase';

export const POSIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#5b9df9" to="#1e44c8">
    <path
      d="M5 2.4a1.2 1.2 0 0 1 1.2-1.2h11.6A1.2 1.2 0 0 1 19 2.4v18.5c0 .86-1 1.33-1.66.78l-1.02-.85a1.2 1.2 0 0 0-1.53 0l-1.18.98a1.2 1.2 0 0 1-1.53 0l-1.17-.98a1.2 1.2 0 0 0-1.53 0l-1.02.85C6 22.28 5 21.8 5 20.95Z"
      fill="#fff"
    />
    <rect x="7.6" y="5.6" width="8.8" height="1.8" rx="0.9" fill="rgba(0,0,0,0.24)" />
    <rect x="7.6" y="9.2" width="8.8" height="1.8" rx="0.9" fill="rgba(0,0,0,0.24)" />
    <rect x="7.6" y="12.8" width="5.4" height="1.8" rx="0.9" fill="rgba(0,0,0,0.24)" />
  </IconBase>
);
