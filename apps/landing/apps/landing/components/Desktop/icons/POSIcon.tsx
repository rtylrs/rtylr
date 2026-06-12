import React from 'react';
import { IconBase } from './IconBase';

export const POSIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#5a9cf8" to="#1d4ed8">
    <path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z" />
    <path d="M14 8H8" />
    <path d="M16 12H8" />
    <path d="M13 16H8" />
  </IconBase>
);
