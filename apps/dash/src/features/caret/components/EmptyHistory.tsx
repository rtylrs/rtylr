import React from 'react';
import { MessageSquareDashed } from 'lucide-react';

/** Friendly empty state for the Caret history page. */
export const EmptyHistory: React.FC = () => (
  <div
    style={{
      padding: '40px 12px',
      textAlign: 'center',
      color: 'var(--dash-text-muted)',
    }}
  >
    <MessageSquareDashed size={28} style={{ color: 'var(--dash-text-dim)', marginBottom: 6 }} />
    <p style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--dash-text)' }}>
      No chats yet
    </p>
    <p style={{ margin: '4px 0 0', fontSize: 12.5 }}>Ask Caret something and it will land here.</p>
  </div>
);

export default EmptyHistory;
