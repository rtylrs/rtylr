import React from 'react';
import { IconBase } from './IconBase';

export const ChangelogIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#2dd4ee" to="#0e7490">
    <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
    <path d="M3 3v5h5" />
    <path d="M12 7v5l4 2" />
  </IconBase>
);
