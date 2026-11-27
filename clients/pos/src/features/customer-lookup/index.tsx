import { createOneKitClient } from '@voxire/api';
import { CustomerServiceClient } from '@voxire/voxie/pos/customer/v1/client';
import React, { useState } from 'react';
import { AlertTriangle, Search, User, UserPlus, X } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useCustomerLookup } from './useCustomerLookup';
import { useCartStore } from '../../shared/stores/cartStore';
import { buildCustomerAttachment } from './customerLookupWorkflow';

const customerClient = createOneKitClient(CustomerServiceClient, 'pos');

const CustomerLookupScreen: React.FC = () => {
  const navigate = useNavigate();
  const { query, setQuery, results, loading, error } = useCustomerLookup();
  const setCustomer = useCartStore((s) => s.setCustomer);
  const currentCustomerId = useCartStore((s) => s.customerId);
  const currentCustomerName = useCartStore((s) => s.customerName);

  const [showCreate, setShowCreate] = useState(false);
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [creating, setCreating] = useState(false);

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

  const handleCreate = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    const trimmedFirst = firstName.trim();
    if (!trimmedFirst) {
      toast.error('First name is required');
      return;
    }
    setCreating(true);
    try {
      const res = await customerClient.createCustomer({
        firstName: trimmedFirst,
        ...(lastName.trim() ? { lastName: lastName.trim() } : {}),
        ...(phone.trim() ? { phone: phone.trim() } : {}),
        ...(email.trim() ? { email: email.trim() } : {}),
      });
      const created = res.customer;
      const fullName = `${created?.firstName ?? trimmedFirst} ${created?.lastName ?? lastName.trim()}`.trim();
      if (created?.id) {
        attach(String(created.id), fullName);
      } else {
        toast.success(`Created ${fullName}`);
        setShowCreate(false);
      }
    } catch (err) {
      const ex = err as { response?: { data?: { message?: string } }; message?: string };
      toast.error(ex?.response?.data?.message ?? ex?.message ?? 'Failed to create customer');
    } finally {
      setCreating(false);
    }
  };

  return (
    <ScreenLayout
      title="Customer Lookup"
      actions={
        <button
          type="button"
          onClick={() => setShowCreate((open) => !open)}
          className={`h-10 px-3.5 rounded-lg text-xs font-semibold inline-flex items-center gap-1.5 transition-colors ${
            showCreate
              ? 'bg-pos-text text-pos-bg'
              : 'bg-white/10 text-pos-text hover:bg-white/20'
          }`}
        >
          <UserPlus className="h-4 w-4" aria-hidden="true" />
          {showCreate ? 'Close Form' : 'New Customer'}
        </button>
      }
    >
      <div className="p-5 space-y-4 max-w-3xl">
        {currentCustomerId && currentCustomerName && (
          <div className="rounded-xl bg-emerald-500/10 border border-emerald-500/30 px-4 py-3 flex items-center justify-between gap-3">
            <div className="min-w-0">
              <p className="text-[10px] font-semibold uppercase tracking-wide text-emerald-300">
                Attached to current order
              </p>
              <p className="text-sm font-semibold text-pos-text truncate">{currentCustomerName}</p>
            </div>
            <button
              type="button"
              onClick={() => {
                setCustomer(null, null);
                toast.success('Detached customer from order');
              }}
              className="h-8 px-3 rounded-lg bg-white/10 hover:bg-white/20 text-xs font-semibold text-pos-text inline-flex items-center gap-1"
            >
              <X className="h-3.5 w-3.5" aria-hidden="true" />
              Remove
            </button>
          </div>
        )}

        {showCreate && (
          <form
            onSubmit={(e) => void handleCreate(e)}
            className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-3"
          >
            <h2 className="text-sm font-bold text-pos-text">Create &amp; Attach Customer</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                placeholder="First name *"
                aria-label="First name"
                className="h-11 px-3.5 rounded-lg bg-white/8 border border-pos-border text-sm text-pos-text outline-none focus:border-white/25"
              />
              <input
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                placeholder="Last name"
                aria-label="Last name"
                className="h-11 px-3.5 rounded-lg bg-white/8 border border-pos-border text-sm text-pos-text outline-none focus:border-white/25"
              />
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="Phone number"
                aria-label="Phone number"
                className="h-11 px-3.5 rounded-lg bg-white/8 border border-pos-border text-sm text-pos-text outline-none focus:border-white/25"
              />
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="Email address"
                aria-label="Email address"
                className="h-11 px-3.5 rounded-lg bg-white/8 border border-pos-border text-sm text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <button
              type="submit"
              disabled={creating || !firstName.trim()}
              className="h-11 px-5 rounded-lg bg-pos-text text-pos-bg text-sm font-bold disabled:opacity-50"
            >
              {creating ? 'Creating…' : 'Create & Attach'}
            </button>
          </form>
        )}

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
          <EmptyState
            icon={Search}
            title="No matches"
            message="No existing customer matched your search. You can create a new customer above."
          />
        ) : (
          <ul className="space-y-2">
            {results.map((c) => (
              <li
                key={c.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between gap-3"
              >
                <div className="min-w-0">
                  <div className="text-sm font-semibold text-pos-text truncate">{c.name}</div>
                  <div className="text-xs text-pos-muted truncate">
                    {c.phone && <span>{c.phone}</span>}
                    {c.phone && c.email && <span> · </span>}
                    {c.email && <span>{c.email}</span>}
                  </div>
                  {c.loyaltyPoints != null && (
                    <div className="text-[11px] font-medium text-amber-300 mt-0.5">
                      {c.loyaltyPoints} loyalty points
                    </div>
                  )}
                </div>
                <button
                  type="button"
                  onClick={() => attach(c.id, c.name)}
                  className="h-11 px-4 rounded-lg bg-pos-text text-pos-bg text-sm font-semibold shrink-0"
                >
                  {String(currentCustomerId ?? '') === c.id ? 'Attached' : 'Attach'}
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
