import React, { useMemo, useState } from 'react';
import { Calculator, Pointer } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import ReceiptPreview from '../../shared/components/ReceiptPreview';
import PrintButton from '../../shared/components/PrintButton';
import { useReceipts } from './useReceipts';
import { formatCurrency, formatDate } from '../../shared/utils/format';
import type { Receipt } from '../../shared/types/pos';

const ReceiptsScreen: React.FC = () => {
  const { receipts, loading, email } = useReceipts();
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState<Receipt | null>(null);
  const [emailInput, setEmailInput] = useState('');

  const searchTerm = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      receipts.filter(
        (r) =>
          !searchTerm ||
          r.orderId.toLowerCase().includes(searchTerm) ||
          (r.customerName ?? '').toLowerCase().includes(searchTerm)
      ),
    [receipts, searchTerm]
  );

  const send = async (id: string) => {
    const ok = await email(id, emailInput);
    if (ok) toast.success('Emailed receipt');
    else toast.error('Failed to email receipt');
    setEmailInput('');
  };

  return (
    <ScreenLayout title="Receipt History">
      <div className="p-5 grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="space-y-3">
          <SearchBar value={search} onChange={setSearch} placeholder="Search receipts..." />
          {loading ? (
            <Skeleton rows={8} />
          ) : filtered.length === 0 ? (
            <EmptyState icon={Calculator} title="No receipts" />
          ) : (
            <ul className="space-y-2">
              {filtered.map((r) => (
                <li
                  key={r.id}
                  onClick={() => setSelected(r)}
                  className={`cursor-pointer p-3 rounded-lg border transition ${
                    selected?.id === r.id
                      ? 'border-white/30 bg-white/10'
                      : 'bg-pos-card border-pos-border hover:border-white/20'
                  }`}
                >
                  <div className="flex justify-between">
                    <span className="text-sm font-semibold text-pos-text">Order #{r.orderId}</span>
                    <span className="text-sm text-pos-text">{formatCurrency(r.total)}</span>
                  </div>
                  <div className="text-[10px] text-pos-muted">{formatDate(r.printedAt)}</div>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="space-y-3">
          {selected ? (
            <>
              <div id="receipt-target">
                <ReceiptPreview receipt={selected} />
              </div>
              <div className="flex flex-wrap gap-2">
                <PrintButton targetId="receipt-target" />
                <input
                  value={emailInput}
                  onChange={(e) => setEmailInput(e.target.value)}
                  placeholder="email@example.com"
                  className="flex-1 h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
                />
                <button
                  onClick={() => send(selected.id)}
                  disabled={!emailInput}
                  className="h-12 px-5 rounded-lg bg-white/10 text-pos-text font-semibold disabled:opacity-50"
                >
                  Email
                </button>
              </div>
            </>
          ) : (
            <EmptyState icon={Pointer} title="Select a receipt" />
          )}
        </div>
      </div>
    </ScreenLayout>
  );
};

export default ReceiptsScreen;
