import { createOneKitClient } from '@voxire/api';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { Order } from '@voxire/voxie/pos/orders/v1/types';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Gift, RefreshCw, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import SearchBar from '../../shared/components/SearchBar';
import { useGiftCards } from './useGiftCards';
import { formatCurrency, formatShortDate } from '../../shared/utils/format';
import {
  defaultGiftCardExpiryDate,
  generateGiftCardCode,
  issueGiftCard,
} from './giftCardWorkflow';
import type { GiftCard } from '../../shared/types/pos';

const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const GiftCardsScreen: React.FC = () => {
  const { cards, issue, redeem, checkBalance, refresh } = useGiftCards();
  const [mode, setMode] = useState<'list' | 'sell' | 'redeem' | 'balance'>('list');
  const [code, setCode] = useState('');
  const [amount, setAmount] = useState('');
  const [paidOrderId, setPaidOrderId] = useState('');
  const [redeemOrderId, setRedeemOrderId] = useState('');
  const [customerName, setCustomerName] = useState('');
  const [customerEmail, setCustomerEmail] = useState('');
  const [expiresAt, setExpiresAt] = useState(() => defaultGiftCardExpiryDate());
  const [orders, setOrders] = useState<Order[]>([]);
  const [loadingOrders, setLoadingOrders] = useState(false);
  const [balanceCard, setBalanceCard] = useState<GiftCard | null>(null);
  const [balanceNotFound, setBalanceNotFound] = useState(false);
  const [redeemPreview, setRedeemPreview] = useState<GiftCard | null>(null);
  const [search, setSearch] = useState('');
  const [issuing, setIssuing] = useState(false);
  const [redeeming, setRedeeming] = useState(false);

  const loadOrders = useCallback(async () => {
    setLoadingOrders(true);
    try {
      const res = await orderClient.listOrders({
        status: '',
        dateFrom: '',
        dateTo: '',
        search: '',
        page: 1,
        pageSize: 25,
      });
      const fetched = res.orders ?? [];
      setOrders(fetched);
      const firstPaid = fetched.find((o) => o.status === 'paid' || o.paymentStatus === 'paid');
      if (firstPaid?.id) {
        setPaidOrderId((prev) => prev || String(firstPaid.id));
      }
      const firstRedeemable = fetched.find(
        (o) => o.status !== 'cancelled' && o.status !== 'voided' && o.status !== 'refunded'
      );
      if (firstRedeemable?.id) {
        setRedeemOrderId((prev) => prev || String(firstRedeemable.id));
      }
    } catch {
      // Non-fatal: users can still inspect existing cards or check balance.
    } finally {
      setLoadingOrders(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
    void loadOrders();
  }, [refresh, loadOrders]);

  const paidOrders = useMemo(
    () => orders.filter((o) => o.status === 'paid' || o.paymentStatus === 'paid'),
    [orders]
  );

  const redeemableOrders = useMemo(
    () =>
      orders.filter(
        (o) => o.status !== 'cancelled' && o.status !== 'voided' && o.status !== 'refunded'
      ),
    [orders]
  );

  const searchTerm = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      cards.filter(
        (c) =>
          !searchTerm ||
          c.code.toLowerCase().includes(searchTerm) ||
          (c.customerName ?? '').toLowerCase().includes(searchTerm)
      ),
    [cards, searchTerm]
  );

  const handleSelectPaidOrder = (nextOrderId: string) => {
    setPaidOrderId(nextOrderId);
    const chosen = paidOrders.find((o) => String(o.id) === nextOrderId);
    if (chosen?.customerName && !customerName.trim()) {
      setCustomerName(chosen.customerName);
    }
    if (chosen?.total && !amount.trim()) {
      setAmount(chosen.total.toFixed(2));
    }
  };

  const handleSell = async () => {
    if (issuing) return;
    if (!paidOrderId) {
      toast.error('Select a paid POS order before issuing a gift card');
      return;
    }
    setIssuing(true);
    try {
      const result = await issueGiftCard({
        codeInput: code,
        amountInput: amount,
        paidOrderId,
        customerName,
        customerEmail,
        expiresAt,
        issue,
      });
      if (!result.ok) {
        toast.error(result.error);
        return;
      }
      const { card } = result;
      toast.success(`Gift card ${card.code} issued (${formatCurrency(card.balance)})`);
      setCode('');
      setAmount('');
      setCustomerName('');
      setCustomerEmail('');
      setExpiresAt(defaultGiftCardExpiryDate());
      void refresh();
      setMode('list');
    } finally {
      setIssuing(false);
    }
  };

  const handleRedeem = async () => {
    if (redeeming) return;
    setRedeeming(true);
    try {
      const result = await redeem(code, amount, redeemOrderId);
      if (!result.ok) {
        toast.error(result.error ?? 'Redemption failed');
        return;
      }
      toast.success(`Redeemed. Remaining: ${formatCurrency(result.remaining ?? 0)}`);
      setCode('');
      setAmount('');
      setRedeemPreview(null);
      void refresh();
    } finally {
      setRedeeming(false);
    }
  };

  const handleCheck = async () => {
    setBalanceNotFound(false);
    const card = await checkBalance(code.trim());
    if (!card) {
      setBalanceCard(null);
      setBalanceNotFound(true);
      return;
    }
    setBalanceCard(card);
  };

  const handlePreviewRedeemCard = async () => {
    const card = await checkBalance(code.trim());
    if (!card) {
      setRedeemPreview(null);
      toast.error('Gift card not found');
      return;
    }
    setRedeemPreview(card);
  };

  return (
    <ScreenLayout
      title="Gift Cards"
      subtitle="Sell, redeem, and check balance"
      actions={
        <button
          type="button"
          onClick={() => {
            void refresh();
            void loadOrders();
          }}
          className="h-9 px-3 rounded-lg bg-white/10 hover:bg-white/20 text-pos-text text-xs font-semibold inline-flex items-center gap-1.5"
        >
          <RefreshCw className="h-3.5 w-3.5" aria-hidden="true" />
          Refresh
        </button>
      }
    >
      <div className="p-5 space-y-4 max-w-4xl">
        <div className="flex flex-wrap gap-2">
          {(['list', 'sell', 'redeem', 'balance'] as const).map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setMode(m)}
              className={`h-11 px-5 rounded-lg font-semibold text-sm capitalize ${
                mode === m ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              {m}
            </button>
          ))}
        </div>

        {mode === 'list' && (
          <div className="space-y-3">
            <SearchBar
              value={search}
              onChange={setSearch}
              placeholder="Filter by code or customer"
              enableVoice={false}
            />
            {filtered.length === 0 ? (
              <EmptyState icon={Gift} title="No gift cards yet" />
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {filtered.map((c) => (
                  <div key={c.id} className="p-4 rounded-xl bg-pos-card border border-pos-border">
                    <div className="flex justify-between items-baseline">
                      <span className="font-mono text-sm text-pos-text">{c.code}</span>
                      <span
                        className={`text-[10px] font-semibold ${
                          c.active ? 'text-emerald-300' : 'text-destructive'
                        }`}
                      >
                        {c.active ? 'ACTIVE' : 'INACTIVE'}
                      </span>
                    </div>
                    <div className="mt-2 text-2xl font-bold text-pos-text">
                      {formatCurrency(c.balance)}
                    </div>
                    {c.customerName && (
                      <div className="text-xs text-pos-muted mt-1 truncate">{c.customerName}</div>
                    )}
                    <div className="text-[10px] text-pos-muted mt-1">
                      Issued {formatShortDate(c.issuedAt)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {mode === 'sell' && (
          <div className="rounded-xl bg-pos-card border border-pos-border p-5 space-y-3">
            <div className="space-y-1">
              <label htmlFor="gc-paid-order" className="text-xs font-medium text-pos-muted">
                Paid POS Order (tender reference) *
              </label>
              <select
                id="gc-paid-order"
                value={paidOrderId}
                onChange={(e) => handleSelectPaidOrder(e.target.value)}
                disabled={loadingOrders}
                className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              >
                <option value="">
                  {loadingOrders
                    ? 'Loading paid orders…'
                    : paidOrders.length === 0
                      ? 'No paid orders found — complete a paid sale first'
                      : 'Select a paid order'}
                </option>
                {paidOrders.map((o) => (
                  <option key={o.id} value={String(o.id)}>
                    #{o.orderNumber || o.id} · {formatCurrency(o.total ?? 0)}
                    {o.customerName ? ` · ${o.customerName}` : ''}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2">
              <input
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="Gift card code"
                className="flex-1 h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text font-mono outline-none focus:border-white/25"
              />
              <button
                type="button"
                onClick={() => setCode(generateGiftCardCode())}
                className="h-12 px-4 rounded-lg bg-white/10 hover:bg-white/20 text-pos-text text-xs font-semibold inline-flex items-center gap-1.5 shrink-0"
              >
                <Sparkles className="h-3.5 w-3.5" aria-hidden="true" />
                Generate
              </button>
            </div>
            <input
              type="number"
              inputMode="decimal"
              step="0.01"
              min="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="Amount"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input
                value={customerName}
                onChange={(e) => setCustomerName(e.target.value)}
                placeholder="Recipient name (optional)"
                className="h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
              <input
                type="email"
                value={customerEmail}
                onChange={(e) => setCustomerEmail(e.target.value)}
                placeholder="Recipient email (optional)"
                className="h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <div className="space-y-1">
              <label htmlFor="gc-expiry" className="text-xs font-medium text-pos-muted">
                Expiry date
              </label>
              <input
                id="gc-expiry"
                type="date"
                value={expiresAt}
                onChange={(e) => setExpiresAt(e.target.value)}
                className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <button
              type="button"
              onClick={() => void handleSell()}
              disabled={issuing || !paidOrderId || !code.trim() || !amount}
              className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              {issuing ? 'Issuing…' : 'Issue Card'}
            </button>
          </div>
        )}

        {mode === 'redeem' && (
          <div className="rounded-xl bg-pos-card border border-pos-border p-5 space-y-3">
            <div className="space-y-1">
              <label htmlFor="gc-redeem-order" className="text-xs font-medium text-pos-muted">
                Order to apply gift card to *
              </label>
              <select
                id="gc-redeem-order"
                value={redeemOrderId}
                onChange={(e) => setRedeemOrderId(e.target.value)}
                disabled={loadingOrders}
                className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              >
                <option value="">
                  {loadingOrders
                    ? 'Loading orders…'
                    : redeemableOrders.length === 0
                      ? 'No active orders found'
                      : 'Select an order'}
                </option>
                {redeemableOrders.map((o) => (
                  <option key={o.id} value={String(o.id)}>
                    #{o.orderNumber || o.id} · {formatCurrency(o.total ?? 0)} · {o.status}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2">
              <input
                value={code}
                onChange={(e) => {
                  setCode(e.target.value);
                  setRedeemPreview(null);
                }}
                placeholder="Gift card code"
                className="flex-1 h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text font-mono outline-none focus:border-white/25"
              />
              <button
                type="button"
                onClick={() => void handlePreviewRedeemCard()}
                disabled={!code.trim()}
                className="h-12 px-4 rounded-lg bg-white/10 hover:bg-white/20 text-pos-text text-xs font-semibold disabled:opacity-50 shrink-0"
              >
                Check
              </button>
            </div>
            {redeemPreview && (
              <div className="rounded-lg bg-white/6 border border-pos-border px-3.5 py-2.5 flex items-center justify-between text-xs">
                <span className="font-mono text-pos-text">{redeemPreview.code}</span>
                <span className={redeemPreview.active ? 'text-emerald-300 font-semibold' : 'text-destructive font-semibold'}>
                  {redeemPreview.active ? `Available: ${formatCurrency(redeemPreview.balance)}` : 'INACTIVE'}
                </span>
              </div>
            )}
            <input
              type="number"
              inputMode="decimal"
              step="0.01"
              min="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="Amount"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <button
              type="button"
              onClick={() => void handleRedeem()}
              disabled={redeeming || !redeemOrderId || !code.trim() || !amount}
              className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              {redeeming ? 'Redeeming…' : 'Redeem'}
            </button>
          </div>
        )}

        {mode === 'balance' && (
          <div className="rounded-xl bg-pos-card border border-pos-border p-5 space-y-3">
            <input
              value={code}
              onChange={(e) => {
                setCode(e.target.value);
                setBalanceNotFound(false);
              }}
              placeholder="Gift card code"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text font-mono outline-none focus:border-white/25"
            />
            <button
              type="button"
              onClick={() => void handleCheck()}
              disabled={!code.trim()}
              className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              Check Balance
            </button>
            {balanceNotFound && (
              <div className="text-center text-sm text-destructive">Card not found</div>
            )}
            {balanceCard && (
              <div className="rounded-xl bg-white/6 border border-pos-border p-4 flex items-center justify-between gap-3">
                <div>
                  <div className="font-mono text-sm font-semibold text-pos-text">{balanceCard.code}</div>
                  {balanceCard.customerName && (
                    <div className="text-xs text-pos-muted">{balanceCard.customerName}</div>
                  )}
                  <div className="text-[11px] text-pos-muted mt-0.5">
                    {balanceCard.active ? 'Active' : 'Inactive'} · Issued {formatShortDate(balanceCard.issuedAt)}
                  </div>
                </div>
                <div className="text-2xl font-bold text-pos-text">
                  {formatCurrency(balanceCard.balance)}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default GiftCardsScreen;
