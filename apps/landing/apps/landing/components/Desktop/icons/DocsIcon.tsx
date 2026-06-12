import React from 'react';
import { IconBase } from './IconBase';

export const DocsIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#48bbf7" to="#0359a8">
    <path d="M6.2 2.2h7.1L19 7.9v12.2a1.7 1.7 0 0 1-1.7 1.7H6.2a1.7 1.7 0 0 1-1.7-1.7V3.9a1.7 1.7 0 0 1 1.7-1.7Z" />
    <path d="M13.3 2.2 19 7.9h-4.3a1.4 1.4 0 0 1-1.4-1.4Z" fill="rgba(0,0,0,0.2)" />
    <rect x="7.4" y="11.2" width="9.2" height="1.7" rx="0.85" fill="rgba(0,0,0,0.22)" />
    <rect x="7.4" y="14.6" width="9.2" height="1.7" rx="0.85" fill="rgba(0,0,0,0.22)" />
    <rect x="7.4" y="18" width="5.6" height="1.7" rx="0.85" fill="rgba(0,0,0,0.22)" />
    <path d="M7.2 2.2h2.6v4.4l-1.3-1-1.3 1Z" fill="#fde047" />
  </IconBase>
);
