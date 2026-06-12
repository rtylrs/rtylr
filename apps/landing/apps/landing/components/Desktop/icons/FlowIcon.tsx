import React from 'react';
import { IconBase } from './IconBase';

export const FlowIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#a98bfb" to="#5118b8">
    <rect x="3.8" y="3.6" width="6.8" height="10" rx="1.8" fill="#fff" />
    <rect x="3.8" y="15.6" width="6.8" height="4.8" rx="1.8" fill="#fff" opacity="0.6" />
    <rect x="13.4" y="3.6" width="6.8" height="4.8" rx="1.8" fill="#fff" opacity="0.6" />
    <rect x="13.4" y="10.4" width="6.8" height="10" rx="1.8" fill="#fff" />
  </IconBase>
);
