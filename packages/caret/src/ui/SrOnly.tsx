import React from 'react';

const srOnly: React.CSSProperties = {
  position: 'absolute',
  width: 1,
  height: 1,
  padding: 0,
  margin: -1,
  overflow: 'hidden',
  clip: 'rect(0,0,0,0)',
  whiteSpace: 'nowrap',
  border: 0,
};

export const SrOnly: React.FC<React.PropsWithChildren> = ({ children }) => (
  <span style={srOnly}>{children}</span>
);

export default SrOnly;
