import React from 'react';
import { IconBase } from './IconBase';

export const ERPIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#8d8df9" to="#3d2fa8">
    <path d="M12 1.9 20.4 6.6 12 11.3 3.6 6.6Z" />
    <path d="M3.6 6.6 12 11.3v9.8l-8.4-4.7Z" opacity="0.75" />
    <path d="M20.4 6.6 12 11.3v9.8l8.4-4.7Z" opacity="0.5" />
    <path d="M12 1.9 20.4 6.6l-2.1 1.18L9.9 3.08Z" fill="#fff" opacity="0.6" />
  </IconBase>
);
