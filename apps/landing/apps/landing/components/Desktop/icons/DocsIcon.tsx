import React from 'react';
import { IconBase } from './IconBase';

export const DocsIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#46b8f5" to="#0369a1">
    <path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" />
    <path d="M14 2v4a2 2 0 0 0 2 2h4" />
    <path d="M10 9H8" />
    <path d="M16 13H8" />
    <path d="M16 17H8" />
  </IconBase>
);
