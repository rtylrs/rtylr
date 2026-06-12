import React from 'react';
import { IconBase } from './IconBase';

export const StatusIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#10493a" to="#03251c">
    <path
      d="M2.4 12h3.3l2.3-6.5a1 1 0 0 1 1.9.02L13.5 18l2.2-5.3a1 1 0 0 1 .92-.62h4.98"
      stroke="#34d399"
      strokeWidth="2.3"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </IconBase>
);
