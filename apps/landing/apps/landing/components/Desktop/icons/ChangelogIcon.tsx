import React from 'react';
import { IconBase } from './IconBase';

export const ChangelogIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#34d8f0" to="#085e78">
    <circle cx="12" cy="12" r="9" fill="#fff" />
    <path
      d="M12 6.6a1.1 1.1 0 0 1 1.1 1.1v3.74l2.86 1.65a1.1 1.1 0 1 1-1.1 1.9l-3.41-1.97a1.1 1.1 0 0 1-.55-.95V7.7A1.1 1.1 0 0 1 12 6.6Z"
      fill="rgba(0,0,0,0.34)"
    />
  </IconBase>
);
