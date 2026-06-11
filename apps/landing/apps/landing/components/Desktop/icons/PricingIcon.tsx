import React from 'react';
import { IconBase } from './IconBase';

export const PricingIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#fbbf24" to="#ea580c">
    <path d="M12.586 2.586A2 2 0 0 0 11.172 2H4a2 2 0 0 0-2 2v7.172a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42z" />
    <circle cx="7.5" cy="7.5" r="1" fill="#fff" stroke="none" />
  </IconBase>
);
