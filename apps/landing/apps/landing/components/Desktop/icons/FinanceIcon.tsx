import React from 'react';
import { IconBase } from './IconBase';

export const FinanceIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#4ade80" to="#15803d">
    <path d="M16 7h6v6" />
    <path d="m22 7-8.5 8.5-5-5L2 17" />
  </IconBase>
);
