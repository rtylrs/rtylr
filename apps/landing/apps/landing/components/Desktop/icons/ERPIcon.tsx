import React from 'react';
import { IconBase } from './IconBase';

export const ERPIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#8d8df9" to="#3d2fa8">
    <path d="M12 1.9 20.4 6.6 12 11.3 3.6 6.6Z" fill="#fff" />
    <path d="M3.6 6.6 12 11.3v9.8l-8.4-4.7Z" fill="#fff" opacity="0.78" />
    <path d="M20.4 6.6 12 11.3v9.8l8.4-4.7Z" fill="#fff" opacity="0.55" />
  </IconBase>
);
