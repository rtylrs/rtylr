import React from 'react';
import Skeleton from './Skeleton';

/** Placeholder shown while history loads from the backend. */
export const MessageSkeleton: React.FC = () => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: 7, margin: '8px 0 12px' }}>
    <Skeleton height={10} width="62%" />
    <Skeleton height={10} width="88%" />
    <Skeleton height={10} width="44%" />
  </div>
);

export default MessageSkeleton;
