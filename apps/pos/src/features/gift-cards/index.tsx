import React, { useEffect, useMemo, useState } from 'react';
import { Gift } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import SearchBar from '../../shared/components/SearchBar';
import { useGiftCards } from './useGiftCards';
import { formatCurrency, formatShortDate } from '../../shared/utils/format';
import { issueGiftCard } from './giftCardWorkflow';

const GiftCardsScreen: React.FC = () => {
  const { cards, issue, redeem, checkBalance, refresh } = useGiftCards();
  const [mode, setMode] = useState<'list' | 'sell' | 'redeem' | 'balance'>('list');
  const [code, setCode] = useState('');
  const [amount, setAmount] = useState('');
  const [balanceResult, setBalanceResult] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [redeeming, setRedeeming] = useState(false);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const searchTerm = search.trim().toLowerCase();
  const filtered = useMemo(
    () => cards.filter((c) => !searchTerm || c.code.toLowerCase().includes(searchTerm)),
    [cards, searchTerm]
  );

  const handleSell = async () => {
    const result = await issueGiftCard({ codeInput: code, amountInput: amount, issue });
    if (!result.ok) {
      toast.error(result.error);
      return;
    }
    const { card } = result;
    toast.success(`Gift card ${card.code} issued (${formatCurrency(card.balance)})`);
    setCode('');
    setAmount('');
    setMode('list');
  };

  const handleRedeem = async () => {
    if (redeeming) return;
    setRedeeming(true);
    try {
      const result = await redeem(code, amount);
      if (!result.ok) {
        toast.error(result.error ?? 'Redemption failed');
        return;
      }
      toast.success(`Redeemed. Remaining: ${formatCurrency(result.remaining ?? 0)}`);
      setCode('');
      setAmount('');
    } finally {
      setRedeeming(false);
    }
  };

  const handleCheck = async () => {
    const card = await checkBalance(code.trim());
    if (!card) {
      setBalanceResult('Card not found');
      return;
    }
    setBalanceResult(`${card.code}: ${formatCurrency(card.balance)}`);
  };

  return (
    <ScreenLayout title="Gift Cards" subtitle="Sell, redeem, and check balance">
      <div className="p-5 space-y-4 max-w-4xl">
        <div className="flex flex-wrap gap-2">
          {(['list', 'sell', 'redeem', 'balance'] as const).map((m) => (
            <button
              key={m}
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
              placeholder="Filter by code"
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
                    <div className="text-[10px] text-pos-muted mt-1">
                      Issued {formatShortDate(c.issuedAt)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {(mode === 'sell' || mode === 'redeem') && (
          <div className="rounded-xl bg-pos-card border border-pos-border p-5 space-y-3">
            <input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="Gift card code"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <input
              type="number"
              inputMode="decimal"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="Amount"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <button
              onClick={mode === 'sell' ? () => void handleSell() : () => void handleRedeem()}
              disabled={mode === 'redeem' && redeeming}
              className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              {mode === 'sell' ? 'Issue Card' : redeeming ? 'Redeeming…' : 'Redeem'}
            </button>
          </div>
        )}

        {mode === 'balance' && (
          <div className="rounded-xl bg-pos-card border border-pos-border p-5 space-y-3">
            <input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="Gift card code"
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            />
            <button
              onClick={handleCheck}
              className="w-full h-12 rounded-lg bg-pos-text text-pos-bg font-bold"
            >
              Check Balance
            </button>
            {balanceResult && (
              <div className="text-center text-sm text-pos-text">{balanceResult}</div>
            )}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default GiftCardsScreen;
