import React from 'react';
import { IconBase } from './IconBase';

export const PricingIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#fcc434" to="#cc4a0a">
    <path
      d="M12.7 2.3h6.7a1.7 1.7 0 0 1 1.7 1.7v6.7a2.1 2.1 0 0 1-.62 1.5l-8.5 8.5a2.1 2.1 0 0 1-2.97 0L3.4 15.1a2.1 2.1 0 0 1 0-2.97l8.5-8.5a2.1 2.1 0 0 1 1.5-.62Z"
      fill="#fff"
    />
    <circle cx="16.4" cy="7.6" r="1.8" fill="rgba(0,0,0,0.26)" />
  </IconBase>
);
