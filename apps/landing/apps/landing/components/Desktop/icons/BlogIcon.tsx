import React from 'react';
import { IconBase } from './IconBase';

export const BlogIcon = ({ size = 44 }: { size?: number }) => (
  <IconBase size={size} from="#fb7b94" to="#ab0a32">
    <path
      d="M16.7 3.2a2.5 2.5 0 0 1 3.54 3.54l-1.2 1.2-3.54-3.54Z"
      fill="#fff"
    />
    <path
      d="M14.06 5.84l3.54 3.54-8.53 8.53a2.2 2.2 0 0 1-1 .57l-3.62 1.05a.65.65 0 0 1-.8-.8l1.05-3.62a2.2 2.2 0 0 1 .57-1Z"
      fill="#fff"
    />
    <path d="M4.45 16.4l2.6 2.6-2.5.72a.65.65 0 0 1-.8-.8Z" fill="rgba(0,0,0,0.24)" />
  </IconBase>
);
