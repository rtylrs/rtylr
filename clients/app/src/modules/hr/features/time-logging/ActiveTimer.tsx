import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Play, Square } from 'lucide-react';
import { Button, Card, ACCENT } from '../../shared/components/ui';

const TIMER_KEY = 'rtylr:hr:active-timer';

interface TimerState {
  startedAt: number;
  description: string;
}

interface ActiveTimerProps {
  onStop: (durationMinutes: number, description: string) => void;
}

function readTimerState(): TimerState | null {
  try {
    const raw = localStorage.getItem(TIMER_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as TimerState;
  } catch {
    return null;
  }
}

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

const ActiveTimer = ({ onStop }: ActiveTimerProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [timerState, setTimerState] = useState<TimerState | null>(readTimerState);
  const [elapsed, setElapsed] = useState(0);
  const [desc, setDesc] = useState('');
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const updateElapsed = useCallback((): void => {
    const state = readTimerState();
    if (!state) {
      // Another tab can stop the timer; drop back to the idle input instead of
      // freezing on a record that no longer exists.
      setTimerState(null);
      setElapsed(0);
      return;
    }
    setElapsed(Math.floor((Date.now() - state.startedAt) / 1000));
  }, []);

  useEffect(() => {
    if (timerState) {
      updateElapsed();
      intervalRef.current = setInterval(updateElapsed, 1000);
      return (): void => {
        if (intervalRef.current) clearInterval(intervalRef.current);
      };
    }
    setElapsed(0);
    return undefined;
  }, [timerState, updateElapsed]);

  // The record lives in localStorage, so a timer stopped in another tab removes
  // it without telling this one. `storage` fires in the tabs that did not write.
  useEffect(() => {
    const handleStorage = (event: StorageEvent): void => {
      if (event.key !== null && event.key !== TIMER_KEY) return;
      setTimerState(readTimerState());
    };
    window.addEventListener('storage', handleStorage);
    return (): void => window.removeEventListener('storage', handleStorage);
  }, []);

  const handleStart = (): void => {
    const state: TimerState = { startedAt: Date.now(), description: desc };
    localStorage.setItem(TIMER_KEY, JSON.stringify(state));
    setTimerState(state);
    setDesc('');
  };

  const handleStop = (): void => {
    const state = readTimerState();
    if (!state) {
      setTimerState(null);
      setElapsed(0);
      return;
    }
    const durationMinutes = Math.max(1, Math.round((Date.now() - state.startedAt) / 60000));
    localStorage.removeItem(TIMER_KEY);
    onStop(durationMinutes, state.description);
    setTimerState(null);
  };

  if (timerState) {
    return (
      <Card
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          padding: '12px 16px',
          border: `1px solid ${ACCENT}40`,
          background: `${ACCENT}08`,
        }}
      >
        <div
          style={{
            width: 8,
            height: 8,
            borderRadius: '50%',
            background: 'var(--hr-error)',
            animation: 'hr-shimmer 1.5s ease-in-out infinite',
            flexShrink: 0,
          }}
        />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 20,
              fontWeight: 700,
              fontFamily: 'monospace',
              color: 'var(--hr-text)',
            }}
          >
            {formatElapsed(elapsed)}
          </div>
          {timerState.description && (
            <div
              style={{
                fontSize: 12,
                color: 'var(--hr-text-muted)',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {timerState.description}
            </div>
          )}
        </div>
        <Button variant="danger" onClick={handleStop} icon={<Square size={14} />}>
          {t('hrTimeLogging.stopTimer')}
        </Button>
      </Card>
    );
  }

  return (
    <Card
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '10px 12px',
      }}
    >
      <input
        value={desc}
        onChange={(e) => setDesc(e.target.value)}
        placeholder={t('hrTimeLogging.timerDescription')}
        style={{
          flex: 1,
          padding: '8px 10px',
          borderRadius: 8,
          border: '1px solid var(--hr-border)',
          background: 'var(--hr-input-bg)',
          color: 'var(--hr-text)',
          fontSize: 13,
          outline: 'none',
          minWidth: 0,
        }}
        onKeyDown={(e) => {
          if (e.key === 'Enter') handleStart();
        }}
      />
      <Button variant="primary" onClick={handleStart} icon={<Play size={14} />}>
        {t('hrTimeLogging.startTimer')}
      </Button>
    </Card>
  );
};

export default ActiveTimer;
