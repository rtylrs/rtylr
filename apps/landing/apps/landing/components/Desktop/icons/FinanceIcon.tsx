import React from 'react';
import { IconBase } from './IconBase';

export const FinanceIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#4cdd85" to="#0f7a38">
    <rect x="3.6" y="12.6" width="4.2" height="8" rx="1.3" opacity="0.55" />
    <rect x="9.9" y="8.4" width="4.2" height="12.2" rx="1.3" opacity="0.78" />
    <rect x="16.2" y="3.4" width="4.2" height="17.2" rx="1.3" />
    <path
      d="M4.4 9.6 9.4 5.9l3.4 2.1 4.6-4.1"
      stroke="#fde047"
      strokeWidth="1.7"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
    <path d="M17.6 2.9h1.9a.5.5 0 0 1 .5.5v1.9Z" fill="#fde047" />
  </IconBase>
);
