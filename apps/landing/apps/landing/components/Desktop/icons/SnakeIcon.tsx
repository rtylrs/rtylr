import React from 'react';
import { IconBase } from './IconBase';

export const SnakeIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#1d6b3c" to="#04240f">
    <path
      d="M4 19.6a1.6 1.6 0 0 1 0-3.2h6.2a.6.6 0 0 0 .6-.6V9.4A3.4 3.4 0 0 1 14.2 6h2.6a1.6 1.6 0 1 1 0 3.2h-2.6a.2.2 0 0 0-.2.2v6.4a3.8 3.8 0 0 1-3.8 3.8Z"
      fill="#4ade80"
    />
    <circle cx="19" cy="17.4" r="2" fill="#f87171" />
  </IconBase>
);
