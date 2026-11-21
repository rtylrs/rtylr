import { useSyncExternalStore } from 'react';
import type { Alarm, AlarmSnapshot } from './alarm';

export function useAlarm(alarm: Alarm): AlarmSnapshot {
  return useSyncExternalStore(alarm.subscribe, alarm.getSnapshot);
}
