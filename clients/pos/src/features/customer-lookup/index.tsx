import React from 'react';
import { AlertTriangle, Search, User } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useCustomerLookup } from './useCustomerLookup';
import { useCartStore } from '../../shared/stores/cartStore';
import { buildCustomerAttachment } from './customerLookupWorkflow';

const CustomerLookupScreen: React.FC = () => {
  const navigate = useNavigate();
  const { query, setQuery, results, loading, error } = useCustomerLookup();
  const setCustomer = useCartStore((s) => s.setCustomer);

  const attach = (id: string, name: string): void => {
    try {
      const attachment = buildCustomerAttachment({ id, name });
      setCustomer(attachment.customerId, attachment.customerName);
      toast.success(`Attached ${attachment.customerName} to order`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to attach customer');
      return;
    }
    navigate(-1);
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
          <EmptyState icon={AlertTriangle} title="Search failed" message={error} />
        ) : !query ? (
          <EmptyState icon={User} title="Start typing to search" />
        ) : results.length === 0 ? (
          <EmptyState icon={Search} title="No matches" />
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
                  onClick={() => attach(c.id, c.name)}
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
