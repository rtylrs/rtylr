import React from 'react';
import { useNavigate } from 'react-router-dom';
import { ChevronRight, ShieldCheck, ShieldAlert } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useCapabilities } from '../../shared/hooks/useCapabilities';

// Each action is gated by a capability flag. There is no separate manager PIN
// in the unified identity model: the signed-in operator either holds the
// capability or a manager signs in via the roster. The destination screens
// self-gate too (defence in depth), so this menu is purely navigational.
const ACTIONS: { id: string; label: string; route: string; cap: keyof ReturnType<typeof useCapabilities> }[] = [
  { id: 'void', label: 'Void order', route: '/orders?reopen=1', cap: 'canVoid' },
  { id: 'refund', label: 'Process refund', route: '/orders', cap: 'canRefund' },
  { id: 'discount', label: 'Apply discount', route: '/orders', cap: 'canApplyDiscount' },
  { id: 'reopen', label: 'Reopen paid order', route: '/orders?reopen=1', cap: 'canVoid' },
];

const OverrideScreen: React.FC = () => {
  const navigate = useNavigate();
  const caps = useCapabilities();

  return (
    <ScreenLayout title="Manager Actions">
      <div className="p-5 max-w-xl space-y-3">
        <p className="text-sm text-pos-muted">
          Actions available to{' '}
          <span className="text-pos-text font-medium capitalize">{caps.role || 'this operator'}</span>.
          Locked actions require a manager to sign in.
        </p>
        <div className="grid grid-cols-1 gap-2">
          {ACTIONS.map((a) => {
            const allowed = Boolean(caps[a.cap]);
            return (
              <button
                key={a.id}
                disabled={!allowed}
                onClick={() => allowed && navigate(a.route)}
                className={`flex items-center justify-between h-14 px-4 rounded-xl text-sm font-semibold transition-colors ${
                  allowed
                    ? 'bg-white/8 text-pos-text hover:bg-white/14'
                    : 'bg-white/4 text-pos-muted cursor-not-allowed'
                }`}
              >
                <span className="flex items-center gap-2">
                  {allowed ? (
                    <ShieldCheck className="w-4 h-4 text-emerald-500" />
                  ) : (
                    <ShieldAlert className="w-4 h-4 text-amber-500" />
                  )}
                  {a.label}
                </span>
                {allowed && <ChevronRight className="w-4 h-4" />}
              </button>
            );
          })}
        </div>
      </div>
    </ScreenLayout>
  );
};

export default OverrideScreen;
