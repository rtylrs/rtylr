import React from 'react';
import Skeleton from './Skeleton';
import Avatar from '../ui/Avatar';

/** Placeholder shown while history loads from the backend. */
export const MessageSkeleton: React.FC = () => (
  <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
    <Avatar role="assistant" initials="C" />
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
      <Skeleton height={11} width="60%" />
      <Skeleton height={11} width="85%" />
      <Skeleton height={11} width="40%" />
    </div>
  </div>
);

export default MessageSkeleton;
