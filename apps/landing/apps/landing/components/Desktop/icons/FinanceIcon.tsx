import React from 'react';
import { IconBase } from './IconBase';

export const FinanceIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#4cdd85" to="#0f7a38">
    <rect x="3.6" y="12.6" width="4.2" height="8" rx="1.3" fill="#fff" opacity="0.6" />
    <rect x="9.9" y="8.4" width="4.2" height="12.2" rx="1.3" fill="#fff" opacity="0.8" />
    <rect x="16.2" y="3.4" width="4.2" height="17.2" rx="1.3" fill="#fff" />
  </IconBase>
);
