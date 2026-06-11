import React from 'react';
import { IconBase } from './IconBase';

export const HRIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#34d6c0" to="#0f766e">
    <rect x="2" y="5" width="20" height="14" rx="2" />
    <circle cx="9" cy="10" r="2" />
    <path d="M6.2 15a3 3 0 0 1 5.6 0" />
    <path d="M15 9h4" />
    <path d="M15 13h4" />
  </IconBase>
);
