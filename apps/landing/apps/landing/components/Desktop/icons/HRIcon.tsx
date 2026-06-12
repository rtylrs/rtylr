import React from 'react';
import { IconBase } from './IconBase';

export const HRIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#36d6bd" to="#0c6b62">
    <rect x="3.4" y="4.4" width="17.2" height="15.2" rx="2.4" />
    <rect x="10.4" y="2.8" width="3.2" height="3.6" rx="1.4" fill="rgba(0,0,0,0.28)" />
    <circle cx="8.7" cy="11" r="2" fill="rgba(0,0,0,0.26)" />
    <path d="M5.6 16.6c.55-1.7 1.75-2.6 3.1-2.6s2.55.9 3.1 2.6Z" fill="rgba(0,0,0,0.26)" />
    <rect x="14" y="9.6" width="4.4" height="1.7" rx="0.85" fill="rgba(0,0,0,0.26)" />
    <rect x="14" y="13" width="4.4" height="1.7" rx="0.85" fill="rgba(0,0,0,0.26)" />
    <circle cx="17.9" cy="17.2" r="1.15" fill="#34d399" />
  </IconBase>
);
