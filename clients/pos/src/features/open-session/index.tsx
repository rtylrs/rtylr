import { createOneKitClient } from '@voxire/api';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/client';
import React, { useRef, useState } from 'react';
import { toast } from 'sonner';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useRegisterSessionStore } from '../../shared/stores/registerSessionStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { useUserStore } from '../../shared/stores/userStore';
import { formatCurrency } from '../../shared/utils/format';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';
import { openRegisterErrorMessage, openRegisterSession } from './openSessionWorkflow';

const sessionClient = createOneKitClient(SessionServiceClient, 'pos');

const OpenRegisterScreen: React.FC = () => {
  const locationId = useTerminalContextStore((s) => s.locationId);
  const staffId = useUserStore((s) => s.userId);
  const openDrawer = useCashDrawerStore((s) => s.openDrawer);
  const setStatus = useRegisterSessionStore((s) => s.setStatus);
  const endSession = useEmployeeSessionStore((s) => s.endSession);
  const [openingFloat, setOpeningFloat] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Two submits inside one tick would race to open two sessions.
  const openingRef = useRef(false);

  const open = async (event: React.FormEvent) => {
    event.preventDefault();
    if (openingRef.current) return;
    openingRef.current = true;
    setBusy(true);
    setError(null);
    try {
      const result = await openRegisterSession({
        floatInput: openingFloat,
        locationId,
        staffId,
        openSession: (payload) => sessionClient.openSession(payload),
        resolveSessionId: resolveOpenSessionId,
        openDrawer,
      });
      if (result.status === 'opened') {
        toast.success(`Register opened with ${formatCurrency(result.openingFloat)} float`);
      } else {
        toast.success('Register is already open');
      }
      setStatus('open');
    } catch (err) {
      setError(openRegisterErrorMessage(err));
      openingRef.current = false;
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-pos-bg p-6">
      <form
        onSubmit={(event) => void open(event)}
        className="w-full max-w-sm bg-pos-card border border-pos-border rounded-2xl p-8 flex flex-col gap-5"
      >
        <div className="flex flex-col gap-1 text-center">
          <h1 className="text-lg font-bold text-pos-text">Open register</h1>
          <p className="text-xs text-pos-muted">
            Count the cash in the drawer and enter it as the opening float to start selling.
          </p>
        </div>
        <div>
          <label htmlFor="opening-float" className="text-xs text-pos-muted">
            Opening float
          </label>
          <input
            id="opening-float"
            type="number"
            inputMode="decimal"
            min="0"
            step="any"
            autoFocus
            value={openingFloat}
            onChange={(e) => setOpeningFloat(e.target.value)}
            placeholder="0.00"
            disabled={busy}
            className="w-full h-14 mt-1 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text text-xl outline-none focus:border-white/25"
          />
        </div>
        {error && (
          <p role="alert" className="text-xs text-destructive text-center">
            {error}
          </p>
        )}
        <button
          type="submit"
          disabled={busy}
          className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold disabled:opacity-50"
        >
          {busy ? 'Opening…' : 'Open register'}
        </button>
        <button
          type="button"
          onClick={endSession}
          disabled={busy}
          className="text-xs text-pos-muted hover:text-pos-text underline disabled:opacity-40"
        >
          Switch employee
        </button>
      </form>
    </div>
  );
};

export default OpenRegisterScreen;
