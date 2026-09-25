import { createOneKitClient } from '@voxire/api';
import { TableServiceClient } from '@voxire/voxie/pos/table/v1/client';
import React, { useCallback, useEffect, useState } from 'react';
import { AlertTriangle, Armchair } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import ScreenLayout from '../../layouts/ScreenLayout';
import TableCard from '../../shared/components/TableCard';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { mapRemoteTables, type TableRow } from './tablesWorkflow';

const tableClient = createOneKitClient(TableServiceClient, 'pos');

const TablesScreen: React.FC = () => {
  const [tables, setTables] = useState<TableRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const res = await tableClient.listTables({});
        setTables(mapRemoteTables(res.tables ?? []));
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        setError(err?.response?.data?.message ?? 'Failed to load tables');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const openTable = useCallback(
    (tableNumber: string) => navigate(`/?table=${tableNumber}`),
    [navigate]
  );

  return (
    <ScreenLayout title="Tables">
      <div className="p-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : tables.length === 0 ? (
          <EmptyState icon={Armchair} title="No tables configured" />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {tables.map((t) => (
              <TableCard key={t.id} {...t} onClick={() => openTable(t.number)} />
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default TablesScreen;
