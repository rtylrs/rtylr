import { unwrap } from '@voxire/api';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenLayout from '../../layouts/ScreenLayout';
import TableCard from '../../shared/components/TableCard';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { posService } from '../../shared/services/api/posService';
import { normalizeTablesResponse, type TableRow } from './tablesWorkflow';

const TablesScreen: React.FC = () => {
  const [tables, setTables] = useState<TableRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const res = await unwrap(posService.getTables());
        setTables(normalizeTablesResponse(res));
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        setError(err?.response?.data?.message ?? 'Failed to load tables');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <ScreenLayout title="Tables">
      <div className="p-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon="⚠" title="Failed to load" message={error} />
        ) : tables.length === 0 ? (
          <EmptyState icon="🪑" title="No tables configured" />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {tables.map((t) => (
              <TableCard key={t.id} {...t} onClick={() => navigate(`/?table=${t.number}`)} />
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default TablesScreen;
