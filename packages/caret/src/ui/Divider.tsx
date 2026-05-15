import React from 'react';
import { caretTokens } from './tokens';

export interface DividerProps {
  orientation?: 'horizontal' | 'vertical';
}

export const Divider: React.FC<DividerProps> = ({ orientation = 'horizontal' }) => (
  <span
    aria-hidden="true"
    style={
      orientation === 'horizontal'
        ? { display: 'block', height: 1, background: caretTokens.border, width: '100%' }
        : { display: 'inline-block', width: 1, background: caretTokens.border, height: '100%' }
    }
  />
);

export default Divider;
