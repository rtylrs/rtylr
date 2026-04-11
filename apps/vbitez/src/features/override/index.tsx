import React, { useState } from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import PinPad from '../../shared/components/PinPad';
import { useOverride } from './useOverride';

const ACTIONS = [
  { id: 'void', label: 'Void order' },
  { id: 'refund', label: 'Process refund' },
  { id: 'discount', label: 'Apply discount' },
  { id: 'price', label: 'Override price' },
  { id: 'reopen', label: 'Reopen paid order' },
];

const OverrideScreen: React.FC = () => {
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
            }
          }}
        />
        {verified && (
          <div className="p-4 rounded-xl border border-emerald-500/30 bg-emerald-500/10 text-emerald-300 text-sm">
            {ACTIONS.find((a) => a.id === action)?.label} approved.
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default OverrideScreen;
