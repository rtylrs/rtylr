import React from 'react';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../components/SearchBar';
import EmptyState from '../../components/EmptyState';
import Skeleton from '../../components/Skeleton';
import { useCustomerLookup } from '../../hooks/useCustomerLookup';
import { toast } from 'sonner';

const CustomerLookupScreen: React.FC = () => {
  const { query, setQuery, results, loading, error } = useCustomerLookup();

  const attach = (name: string) => {
    toast.success(`Attached ${name} to current order`);
  };

  return (
    <ScreenLayout title="Customer Lookup">
      <div className="p-5 space-y-4 max-w-3xl">
        <SearchBar
          value={query}
          onChange={setQuery}
          placeholder="Search by name, phone, or email..."
        />
        {loading ? (
          <Skeleton rows={5} />
        ) : error ? (
          <EmptyState icon="⚠" title="Search failed" message={error} />
        ) : !query ? (
          <EmptyState icon="👤" title="Start typing to search" />
        ) : results.length === 0 ? (
          <EmptyState icon="🔍" title="No matches" />
        ) : (
          <ul className="space-y-2">
            {results.map((c) => (
              <li
                key={c.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
              >
                <div>
                  <div className="text-sm font-semibold text-pos-text">{c.name}</div>
                  <div className="text-xs text-pos-muted">
                    {c.phone && <span>{c.phone}</span>}
                    {c.phone && c.email && <span> · </span>}
                    {c.email && <span>{c.email}</span>}
                  </div>
                  {c.loyaltyPoints != null && (
                    <div className="text-[10px] text-amber-300">
                      {c.loyaltyPoints} loyalty points
                    </div>
                  )}
                </div>
                <button
                  onClick={() => attach(c.name)}
                  className="h-11 px-4 rounded-lg bg-pos-text text-pos-bg text-sm font-semibold"
                >
                  Attach
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </ScreenLayout>
  );
};

export default CustomerLookupScreen;
