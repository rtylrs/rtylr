import React, { useState, useEffect, useCallback } from 'react';
import { createOneKitClient } from '@voxire/api';
import { toast } from 'sonner';
import { Loader2, ChevronLeft, ShieldCheck } from 'lucide-react';
import { AccessServiceClient } from '@voxire/voxie/pos/access/v1/client';
import { UserServiceClient } from '@voxire/voxie/pos/user/v1/client';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';
import { normalizeDigits } from '../../shared/utils/digits';
import { PIN_MAX_LENGTH, PIN_MIN_LENGTH, isNewPin, isVerifiablePin } from './pinPolicy';
import type { RosterMember } from './rosterWorkflow';

const accessClient = createOneKitClient(AccessServiceClient, 'pos');
const userClient = createOneKitClient(UserServiceClient, 'pos');

const KEYS = [
  ['1', '2', '3'],
  ['4', '5', '6'],
  ['7', '8', '9'],
  ['', '0', '⌫'],
];

type Step = 'roster' | 'enter' | 'setup';

const PinLogin: React.FC = () => {
  const locationId = useTerminalContextStore((s) => s.locationId);
  const authenticatePin = useEmployeeSessionStore((s) => s.authenticatePin);
  const loading = useEmployeeSessionStore((s) => s.loading);
  const error = useEmployeeSessionStore((s) => s.error);
  const clearError = useEmployeeSessionStore((s) => s.clearError);

  const [roster, setRoster] = useState<RosterMember[]>([]);
  const [rosterLoading, setRosterLoading] = useState(true);
  const [rosterError, setRosterError] = useState<string | null>(null);
  const [step, setStep] = useState<Step>('roster');
  const [member, setMember] = useState<RosterMember | null>(null);
  const [pin, setPin] = useState('');
  const [confirmPin, setConfirmPin] = useState('');
  const [settingPin, setSettingPin] = useState(false);
  const [revealPin, setRevealPin] = useState(false);
  const [shake, setShake] = useState(false);
  const submittingRef = React.useRef(false);
  const mountedRef = React.useRef(true);

  const busy = loading || settingPin;

  const loadRoster = useCallback(async () => {
    if (!locationId) return;
    setRosterLoading(true);
    setRosterError(null);
    try {
      const data = await accessClient.getLocationRoster({ locationId: String(locationId) });
      setRoster(
        (data.roster ?? []).flatMap((entry) => {
          const userId = Number(entry.userId);
          const name = entry.name ?? '';
          if (!Number.isFinite(userId) || userId <= 0 || !name.trim()) return [];
          return [
            {
              userId,
              name,
              hasPin: entry.hasPin ?? false,
              isManager: entry.isManager ?? false,
              posRole: entry.posRole ?? '',
            },
          ];
        })
      );
    } catch {
      setRosterError('Unable to load the staff list. Check your connection and try again.');
    } finally {
      setRosterLoading(false);
    }
  }, [locationId]);

  useEffect(() => {
    void loadRoster();
  }, [loadRoster]);

  useEffect(() => {
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const flashShake = useCallback(() => {
    if (!mountedRef.current) return;
    setShake(true);
    window.setTimeout(() => {
      if (mountedRef.current) setShake(false);
    }, 500);
  }, []);

  const resetToRoster = (): void => {
    if (busy) return;
    setStep('roster');
    setMember(null);
    setPin('');
    setConfirmPin('');
    clearError();
  };

  const pickMember = (m: RosterMember): void => {
    clearError();
    setMember(m);
    setPin('');
    setConfirmPin('');
    setStep(m.hasPin ? 'enter' : 'setup');
  };

  const submitEnter = useCallback(
    async (value: string) => {
      if (!member || !locationId) return;
      if (submittingRef.current) return;
      submittingRef.current = true;
      try {
        await authenticatePin(locationId, member.userId, value);
        // success → AppShell sees isActive and renders the POS
      } catch {
        flashShake();
        if (mountedRef.current) setPin('');
      } finally {
        submittingRef.current = false;
      }
    },
    [authenticatePin, flashShake, locationId, member]
  );

  const submitSetup = useCallback(async () => {
    if (!member || !locationId) return;
    if (submittingRef.current) return;
    if (pin !== confirmPin) {
      flashShake();
      setConfirmPin('');
      toast.error('PINs do not match');
      return;
    }
    submittingRef.current = true;
    setSettingPin(true);
    try {
      await userClient.setPIN({
        userId: String(member.userId),
        locationId: String(locationId),
        name: member.name,
        pin,
      });
      // Immediately sign in with the PIN just set.
      await authenticatePin(locationId, member.userId, pin);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Could not set PIN';
      toast.error(msg);
      if (mountedRef.current) {
        setPin('');
        setConfirmPin('');
        setStep('setup');
      }
    } finally {
      submittingRef.current = false;
      if (mountedRef.current) setSettingPin(false);
    }
  }, [authenticatePin, confirmPin, flashShake, locationId, member, pin]);

  // A PIN is 4-12 digits now, so no single "field is full" length exists to
  // submit on: submitting at 4 would truncate a 6-12 digit PIN and submitting at
  // 6 would lock out a legacy 4-digit one. Both steps confirm explicitly.
  const isSetup = step === 'setup';
  const canSubmit = isSetup ? isNewPin(pin) && isNewPin(confirmPin) : isVerifiablePin(pin);

  const confirmSubmit = useCallback(() => {
    if (busy || submittingRef.current || !canSubmit) return;
    if (isSetup) void submitSetup();
    else void submitEnter(pin);
  }, [busy, canSubmit, isSetup, pin, submitEnter, submitSetup]);

  // The active field for the keypad: PIN first, then confirm during setup.
  const activeValue = isSetup && isNewPin(pin) ? confirmPin : pin;

  useEffect(() => {
    if (step !== 'setup' || activeValue.length === 0) {
      setRevealPin(false);
      return;
    }
    setRevealPin(true);
    const timeout = window.setTimeout(() => setRevealPin(false), 900);
    return () => window.clearTimeout(timeout);
  }, [activeValue, step]);

  // The keypad writes to the confirm field once the first field is a whole PIN.
  const setActive = (updater: (prev: string) => string): void => {
    if (isSetup && isNewPin(pin)) setConfirmPin(updater);
    else setPin(updater);
  };

  const pushDigit = useCallback(
    (d: string) => {
      if (busy) return;
      clearError();
      setActive((p) => (p.length < PIN_MAX_LENGTH ? p + d : p));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [busy, clearError, step, pin]
  );

  const deleteDigit = useCallback(() => {
    if (busy) return;
    clearError();
    setActive((p) => p.slice(0, -1));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [busy, clearError, step, pin]);

  useEffect(() => {
    if (step === 'roster' || busy) return;
    const onKey = (e: KeyboardEvent): void => {
      // An Arabic-locale keyboard reports ٤ rather than 4; comparing the raw
      // key against '0'-'9' would drop the keystroke and the PIN never fills.
      const digit = normalizeDigits(e.key);
      if (digit.length === 1 && digit >= '0' && digit <= '9') pushDigit(digit);
      if (e.key === 'Backspace') deleteDigit();
      if (e.key === 'Enter') confirmSubmit();
      if (e.key === 'Escape') resetToRoster();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [step, busy, pushDigit, deleteDigit, confirmSubmit]);

  const handleKey = (key: string): void => {
    if (busy) return;
    if (key === '⌫') deleteDigit();
    else if (key !== '') pushDigit(key);
  };

  // ── Roster step ────────────────────────────────────────────────
  if (step === 'roster') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-pos-bg p-6">
        <div className="w-full max-w-md">
          <div className="flex flex-col items-center gap-1 text-center mb-8">
            <RtylrWordmark className="text-2xl" />
            <h1 className="text-lg font-bold text-pos-text mt-2">Who's working?</h1>
            <p className="text-xs text-pos-muted">Tap your name to sign in</p>
          </div>

          {rosterLoading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="w-7 h-7 animate-spin text-pos-muted" />
            </div>
          ) : rosterError ? (
            <div className="text-center py-10">
              <p className="text-sm text-destructive mb-3">{rosterError}</p>
              <button
                onClick={() => void loadRoster()}
                className="text-xs text-pos-muted hover:text-pos-text underline"
              >
                Retry
              </button>
            </div>
          ) : roster.length === 0 ? (
            <div className="text-center py-10">
              <p className="text-sm text-pos-muted">No staff assigned to this location yet.</p>
              <p className="text-xs text-pos-muted mt-1">
                Ask your manager to assign you to this location in the dashboard.
              </p>
            </div>
          ) : (
            <div
              className={
                roster.length === 1 ? 'flex justify-center' : 'grid grid-cols-2 gap-3'
              }
            >
              {roster.map((m) => (
                <button
                  key={m.userId}
                  onClick={() => pickMember(m)}
                  className={`flex w-full flex-col items-center gap-2 rounded-2xl border border-pos-border bg-pos-card p-5 transition-colors hover:border-blue-500/50 ${
                    roster.length === 1 ? 'max-w-72' : ''
                  }`}
                >
                  <div className="w-14 h-14 rounded-full bg-blue-500/10 flex items-center justify-center text-lg font-bold text-blue-400">
                    {m.name.charAt(0).toUpperCase()}
                  </div>
                  <span className="text-sm font-semibold text-pos-text text-center leading-tight">
                    {m.name}
                  </span>
                  <span className="text-[10px] text-pos-muted capitalize flex items-center gap-1">
                    {m.isManager && <ShieldCheck className="w-3 h-3 text-emerald-500" />}
                    {m.hasPin ? m.posRole || 'staff' : 'Set up PIN'}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    );
  }

  // ── PIN entry / setup step ─────────────────────────────────────
  const confirming = isSetup && isNewPin(pin);
  const title = isSetup
    ? confirming
      ? 'Confirm your PIN'
      : 'Create your PIN'
    : `Hi ${member?.name?.split(' ')[0] ?? ''}`;
  const subtitle = isSetup
    ? confirming
      ? 'Re-enter to confirm'
      : `Pick a ${PIN_MIN_LENGTH}-${PIN_MAX_LENGTH} digit PIN for this location`
    : 'Enter your PIN';
  const busyTitle = settingPin ? 'Setting up PIN' : 'Verifying PIN';
  const busySubtitle = settingPin ? 'Saving this PIN and signing you in' : 'Checking your PIN';
  const displayTitle = busy ? busyTitle : title;
  const displaySubtitle = busy ? busySubtitle : subtitle;

  return (
    <div className="min-h-screen flex items-center justify-center bg-pos-bg">
      <div className="w-full max-w-sm bg-pos-card border border-pos-border rounded-2xl p-8 flex flex-col items-center gap-6">
        <button
          onClick={resetToRoster}
          disabled={busy}
          className="self-start flex items-center gap-1 text-xs text-pos-muted hover:text-pos-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <ChevronLeft className="w-3.5 h-3.5" /> Back
        </button>

        <div className="flex flex-col items-center gap-1 text-center">
          <h1 className="text-lg font-bold text-pos-text">{displayTitle}</h1>
          <p className="text-xs text-pos-muted">{displaySubtitle}</p>
        </div>

        {busy ? (
          <div
            role="status"
            aria-live="polite"
            className="flex h-64 w-full flex-col items-center justify-center gap-4"
          >
            <Loader2 className="h-9 w-9 animate-spin text-pos-muted" />
            <span className="sr-only">{busyTitle}</span>
          </div>
        ) : (
          <>
            <div className={`flex w-full gap-2 ${shake ? 'animate-[shake_0.5s_ease-in-out]' : ''}`}>
              {Array.from({ length: PIN_MAX_LENGTH }).map((_, i) => (
                <span
                  key={i}
                  className={`flex-1 h-12 rounded-xl border flex items-center justify-center text-lg font-bold tabular-nums transition-colors ${
                    error
                      ? 'border-destructive bg-destructive/15 text-destructive'
                      : i < activeValue.length
                        ? 'border-pos-text/70 bg-white/8 text-pos-text'
                        : 'border-pos-border bg-transparent text-pos-muted'
                  }`}
                >
                  {i < activeValue.length ? (isSetup && revealPin ? activeValue[i] : '•') : ''}
                </span>
              ))}
            </div>

            {error && <p className="text-xs text-destructive text-center">{error}</p>}

            <div className="flex flex-col gap-2 w-full">
              {KEYS.map((row, ri) => (
                <div key={ri} className="flex gap-2 justify-center">
                  {row.map((key, ki) => (
                    <button
                      key={ki}
                      className={`w-16 h-14 rounded-xl text-base font-semibold transition-colors ${
                        key === ''
                          ? 'invisible'
                          : key === '⌫'
                            ? 'bg-transparent text-pos-muted hover:bg-white/8'
                            : 'bg-white/8 text-pos-text hover:bg-white/14 active:scale-95'
                      }`}
                      onClick={() => handleKey(key)}
                      disabled={key === ''}
                      aria-label={key === '⌫' ? 'Delete' : key}
                    >
                      {key}
                    </button>
                  ))}
                </div>
              ))}
            </div>

            <button
              onClick={confirmSubmit}
              disabled={!canSubmit}
              className="w-full h-11 rounded-xl text-sm font-semibold bg-pos-text text-pos-bg disabled:opacity-40"
            >
              {isSetup ? 'Set PIN' : 'Sign in'}
            </button>
          </>
        )}
      </div>
    </div>
  );
};

export default PinLogin;
