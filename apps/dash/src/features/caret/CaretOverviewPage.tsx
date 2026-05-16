import React from 'react';
import { useCaretChats } from '../../shared/hooks/queries/useCaretChats';
import { useCaretTools } from '../../shared/hooks/queries/useCaretTools';
import { PageHeader, Tile } from './components';
import { commaNumber } from '../../shared/lib/caret';

const CaretOverviewPage: React.FC = () => {
  const { data: chats = [] } = useCaretChats();
  const { data: tools = [] } = useCaretTools();

  const totalChats = chats.length;
  const pinned = chats.filter((c) => c.pinned).length;
  const writeTools = tools.filter((t) => t.scope === 'write').length;

  return (
    <div className="space-y-4 max-w-4xl">
      <PageHeader title="Caret overview" subtitle="A quick read on how Caret is being used." />
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
          gap: 12,
        }}
      >
        <Tile title="Total chats" value={commaNumber(totalChats)} />
        <Tile title="Pinned" value={commaNumber(pinned)} accent="#f59e0b" />
        <Tile title="Available tools" value={commaNumber(tools.length)} accent="#10b981" />
        <Tile title="Write-scoped tools" value={commaNumber(writeTools)} accent="#f43f5e" />
      </div>
    </div>
  );
};

export default CaretOverviewPage;
