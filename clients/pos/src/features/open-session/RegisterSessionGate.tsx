import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useRegisterSessionStore } from '../../shared/stores/registerSessionStore';
import { useUserStore } from '../../shared/stores/userStore';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';
import OpenRegisterScreen from './index';

// Closing a session lands here; the summary stays readable before the next open.
const SESSION_FREE_PATHS = new Set(['/day-summary']);

// The lookup records the status itself. A failed lookup (offline) leaves the
// status untouched so the offline queues keep working as before.
function checkRegisterSession(): Promise<void> {
  return resolveOpenSessionId().then(
    () => undefined,
    () => undefined
  );
}

const RegisterSessionGate: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const status = useRegisterSessionStore((s) => s.status);
  const userId = useUserStore((s) => s.userId);
  const { pathname } = useLocation();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let active = true;
    void checkRegisterSession().finally(() => {
      if (active) setChecking(false);
    });
    const recheck = (): void => {
      void checkRegisterSession();
    };
    window.addEventListener('online', recheck);
    return () => {
      active = false;
      window.removeEventListener('online', recheck);
    };
  }, [userId]);

  if (checking) {
    return (
      <div role="status" className="min-h-screen flex items-center justify-center bg-pos-bg">
        <Loader2 className="h-7 w-7 animate-spin text-pos-muted" />
        <span className="sr-only">Checking register session</span>
      </div>
    );
  }

  if (status === 'none' && !SESSION_FREE_PATHS.has(pathname)) {
    return <OpenRegisterScreen />;
  }

  return <>{children}</>;
};

export default RegisterSessionGate;
