import React from 'react';
import { IconBase } from './IconBase';

export const DoomIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#591111" to="#180404">
    <path
      d="M12 1.6c.3 2.8 1.9 5.3 3.9 6.9 2.1 1.7 3.4 3.9 3.4 6.1a7.3 7.3 0 1 1-14.6 0c0-1.2.45-2.4 1.05-3.15a2.6 2.6 0 0 0 5.05-.85c0-1.44-.52-2.08-1.04-3.12C8.67 5.2 9.55 3.6 12 1.6Z"
      fill="#fb923c"
    />
    <path
      d="M12 21.9a4.6 4.6 0 0 1-4.6-4.6c0-1.6.9-2.9 1.9-3.9.3 1 1.2 1.8 2.3 1.8 1.3 0 2.3-.9 2.5-2.1 1.5 1.1 2.5 2.6 2.5 4.2a4.6 4.6 0 0 1-4.6 4.6Z"
      fill="#fde047"
      opacity="0.9"
    />
  </IconBase>
);
