import React from 'react';
import { IconBase } from './IconBase';

export const FlowIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#a78bfa" to="#6d28d9">
    <rect x="3" y="3" width="8" height="8" rx="2" />
    <path d="M7 11v4a2 2 0 0 0 2 2h4" />
    <rect x="13" y="13" width="8" height="8" rx="2" />
  </IconBase>
);
