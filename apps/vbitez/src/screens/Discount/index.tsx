import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import PinPad from '../../components/PinPad';
import { posService } from '../../services/posService';
import { useOverride } from '../../hooks/useOverride';
import { formatCurrency } from '../../utils/format';

const DiscountScreen: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const [mode, setMode] = useState<'percent' | 'flat'>('percent');
  const [value, setValue] = useState('');
  const [verified, setVerified] = useState(false);
  const { verify, error } = useOverride();

  const submit = async () => {
    try {
      await posService.applyDiscount(orderId!, {
        mode,
        value: Number(value),
      });
      toast.success('Discount applied');
      navigate(-1);
    } catch {
      toast.error('Could not apply discount');
    }
  };

  return (
    <ScreenLayout title="Apply Discount" subtitle={`Order #${orderId ?? 'current'}`}>
      <div className="p-5 max-w-md space-y-5">
        {!verified ? (
          <div className="space-y-3">
            <p className="text-sm text-pos-muted">Manager PIN required to apply discount.</p>
            <PinPad
              label="Enter manager PIN"
              error={error}
              onComplete={async (pin) => {
                const ok = await verify(pin);
                if (ok) setVerified(true);
              }}
            />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => setMode('percent')}
                className={`h-14 rounded-lg font-semibold ${
                  mode === 'percent' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
                }`}
              >
                Percent
              </button>
              <button
                onClick={() => setMode('flat')}
                className={`h-14 rounded-lg font-semibold ${
                  mode === 'flat' ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
                }`}
              >
                Flat amount
              </button>
            </div>
            <div>
              <label className="text-xs text-pos-muted">
                {mode === 'percent' ? 'Discount %' : 'Amount'}
              </label>
              <input
                type="number"
                inputMode="decimal"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={mode === 'percent' ? '10' : '5.00'}
                className="w-full h-12 mt-1 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <button
              onClick={submit}
              disabled={!Number(value)}
              className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
            >
              Apply{' '}
              {mode === 'percent'
                ? `${Number(value || 0)}%`
                : formatCurrency(Number(value || 0))}
            </button>
          </>
        )}
      </div>
    </ScreenLayout>
  );
};

export default DiscountScreen;
