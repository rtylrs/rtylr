import { useEffect, useRef } from 'react';
import { useSettingsStore } from '../stores/settingsStore';
import { useEmployeeSessionStore } from '../stores/employeeSessionStore';

// Logs the employee out after N minutes of inactivity (mouse/keyboard/touch).
export function useAutoLogout() {
  const minutes = useSettingsStore((s) => s.autoLogoutMinutes);
  const endSession = useEmployeeSessionStore((s) => s.endSession);
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    if (!minutes || minutes <= 0) return;
    const reset = () => {
      if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
      timeoutRef.current = window.setTimeout(() => endSession(), minutes * 60 * 1000);
    };
    const events = ['mousemove', 'keydown', 'pointerdown', 'touchstart'];
    events.forEach((e) => window.addEventListener(e, reset));
    reset();
    return () => {
      events.forEach((e) => window.removeEventListener(e, reset));
      if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    };
  }, [minutes, endSession]);
}
