import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import PinPad from '../../shared/components/PinPad';
import { useOverride } from './useOverride';

const ACTIONS = [
  { id: 'void', label: 'Void order', route: '/void' },
  { id: 'refund', label: 'Process refund', route: '/refund' },
  { id: 'discount', label: 'Apply discount', route: '/discount' },
  { id: 'price', label: 'Override price', route: '/products/quick' },
  { id: 'reopen', label: 'Reopen paid order', route: '/orders?reopen=1' },
];

const OverrideScreen: React.FC = () => {
  const navigate = useNavigate();
  const { verify, error, clearError } = useOverride();
  const [action, setAction] = useState(ACTIONS[0].id);
  const [verified, setVerified] = useState(false);

  return (
    <ScreenLayout title="Manager Override">
      <div className="p-5 max-w-xl space-y-4">
        <p className="text-sm text-pos-muted">
          Select the action and enter your manager PIN to approve.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          {ACTIONS.map((a) => (
            <button
              key={a.id}
              onClick={() => {
                setAction(a.id);
                setVerified(false);
                clearError();
              }}
              className={`h-12 rounded-lg text-sm font-semibold ${
                action === a.id ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              {a.label}
            </button>
          ))}
        </div>
        <PinPad
          label="Manager PIN"
          error={error}
          onComplete={async (pin) => {
            const ok = await verify(pin);
            if (ok) {
              setVerified(true);
              toast.success('Approved');
              const actionDef = ACTIONS.find((a) => a.id === action);
              if (actionDef?.route) {
                setTimeout(() => navigate(actionDef.route!), 600);
              }
            }
          }}
        />
        {verified && (
          <div className="p-4 rounded-xl border border-emerald-500/30 bg-emerald-500/10 text-emerald-300 text-sm space-y-1">
            <div className="font-semibold">
              {ACTIONS.find((a) => a.id === action)?.label} — approved
            </div>
            {action === 'price' && (
              <div className="text-xs text-emerald-400/70">
                Redirecting to quick edit — you can now modify item prices on the current order.
              </div>
            )}
            {action === 'reopen' && (
              <div className="text-xs text-emerald-400/70">
                Redirecting to order history — select the order and tap Reopen.
              </div>
            )}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default OverrideScreen;
