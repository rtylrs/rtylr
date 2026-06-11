import React from 'react';
import { IconBase } from './IconBase';

export const SnakeIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#14532d" to="#052e16" stroke="#4ade80">
    <path d="M5 18h6a2 2 0 0 0 2-2v-6a2 2 0 0 1 2-2h2" strokeWidth="3" />
    <circle cx="19" cy="17" r="1.6" fill="#f87171" stroke="none" />
  </IconBase>
);
