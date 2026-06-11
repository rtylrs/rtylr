import React from 'react';
import { IconBase } from './IconBase';

export const POSIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#5a9cf8" to="#1d4ed8">
    <rect x="2" y="5" width="20" height="14" rx="2" />
    <line x1="2" y1="10" x2="22" y2="10" />
  </IconBase>
);
