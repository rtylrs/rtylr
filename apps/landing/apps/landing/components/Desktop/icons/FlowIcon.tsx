import React from 'react';
import { IconBase } from './IconBase';

export const FlowIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#a98bfb" to="#5118b8">
    <rect x="3.8" y="3.6" width="6.8" height="10" rx="1.8" />
    <rect x="3.8" y="15.6" width="6.8" height="4.8" rx="1.8" opacity="0.55" />
    <rect x="13.4" y="3.6" width="6.8" height="4.8" rx="1.8" opacity="0.55" />
    <rect x="13.4" y="10.4" width="6.8" height="10" rx="1.8" fill="#fde047" />
    <rect x="15.1" y="12.6" width="3.4" height="1.5" rx="0.75" fill="rgba(0,0,0,0.22)" />
    <rect x="15.1" y="15.4" width="2.4" height="1.5" rx="0.75" fill="rgba(0,0,0,0.22)" />
  </IconBase>
);
