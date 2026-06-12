import React from 'react';
import { IconBase } from './IconBase';

export const LandingIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#f5f2ec" to="#c4beb2">
    <g stroke="#17171b" strokeWidth="1.8" strokeLinecap="round" fill="none">
      <circle cx="12" cy="12" r="8.8" />
      <ellipse cx="12" cy="12" rx="4" ry="8.8" />
      <path d="M3.8 9.2h16.4" />
      <path d="M3.8 14.8h16.4" />
    </g>
  </IconBase>
);
